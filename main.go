package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	mastodon "github.com/mattn/go-mastodon"
	"github.com/spf13/cobra"
)

var (
	client     *mastodon.Client
	self       *mastodon.Account
	configFile string

	// RootCmd is the core command used for cli-arg parsing
	RootCmd = &cobra.Command{
		Use:           "mastotool",
		Short:         "mastotool offers a collection of tools to work with your Mastodon account",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
)

func registerApp(config *Config) (string, error) {
	app, err := mastodon.RegisterApp(context.Background(), &mastodon.AppConfig{
		Server:     config.Value("instance").(string),
		ClientName: "mastotool",
		Scopes:     "read",
		Website:    "",
	})
	if err != nil {
		return "", err
	}

	config.Set("id", app.ClientID)
	config.Set("secret", app.ClientSecret)
	config.Set("redirectURI", app.RedirectURI)

	return app.AuthURI, nil
}

func initClient() error {
	var err error
	var instance, token, redirectURI, authURI, id, secret string
	config, err := LoadConfig(configFile)
	if err == nil {
		instance = config.Value("instance").(string)
		id = config.Value("id").(string)
		secret = config.Value("secret").(string)
		token = config.Value("token").(string)
		redirectURI = config.Value("redirectURI").(string)
	}

	scanner := bufio.NewScanner(os.Stdin)
	if len(instance) == 0 {
		fmt.Print("Which instance to connect to (e.g. https://mastodon.social): ")
		scanner.Scan()
		if scanner.Err() != nil {
			return fmt.Errorf("Can't open input: %s", err)
		}
		instance = scanner.Text()
		config.Set("instance", instance)
	}

	if len(id) == 0 {
		authURI, err = registerApp(&config)
		if err != nil {
			return fmt.Errorf("Can't register app: %s", err)
		}

		id = config.Value("id").(string)
		secret = config.Value("secret").(string)
		redirectURI = config.Value("redirectURI").(string)
	}

	mConfig := &mastodon.Config{
		AccessToken:  token,
		Server:       instance,
		ClientID:     id,
		ClientSecret: secret,
	}
	client = mastodon.NewClient(mConfig)

	if len(mConfig.AccessToken) == 0 {
		fmt.Printf("Please visit %s and enter the generated token: ", authURI)
		scanner.Scan()
		if scanner.Err() != nil {
			return fmt.Errorf("Can't open input: %s", err)
		}
		code := scanner.Text()

		err = client.AuthenticateToken(context.Background(), code, redirectURI)
		if err != nil {
			return fmt.Errorf("Can't retrieve authentication token: %s", err)
		}

		config.Set("token", mConfig.AccessToken)
		err = config.Save(configFile)
		if err != nil {
			return fmt.Errorf("Can't save config: %s", err)
		}
	}

	return nil
}

func main() {
	RootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "mastodon.json", "uses the specified config file")

	if err := initClient(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var err error
	self, err = client.GetAccountCurrentUser(context.Background())
	if err != nil {
		fmt.Printf("Can't retrieve user: %s\n", err)
		os.Exit(1)
	}

	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
