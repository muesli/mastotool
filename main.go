package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"math"
	"os"
	"time"

	mastodon "github.com/mattn/go-mastodon"
	"github.com/muesli/goprogressbar"
)

var (
	client *mastodon.Client
	self   *mastodon.Account

	topN       = flag.Int("top", 10, "shows the top N items in each category")
	maxToots   = flag.Int("recent", 0, "only account for the N most recent toots (excl replies & boosts)")
	columns    = flag.Int("columns", 80, "displays tables with N columns")
	configFile = flag.String("config", "mastodon.json", "uses the specified config file")
	// user     = flag.String("user", "@fribbledom@mastodon.social", "shows stats for this user")
)

func registerApp(config *Config) (string, error) {
	app, err := mastodon.RegisterApp(context.Background(), &mastodon.AppConfig{
		Server:     config.Value("instance").(string),
		ClientName: "statootstics",
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

func initClient() {
	var err error
	var instance, token, redirectURI, authURI, id, secret string
	config, err := LoadConfig(*configFile)
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
			panic(err)
		}
		instance = scanner.Text()
		config.Set("instance", instance)
	}

	if len(id) == 0 {
		authURI, err = registerApp(&config)
		if err != nil {
			panic(err)
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
			panic(err)
		}
		code := scanner.Text()

		err = client.AuthenticateToken(context.Background(), code, redirectURI)
		if err != nil {
			panic(err)
		}

		config.Set("token", mConfig.AccessToken)
		config.Save(*configFile)
	}
}

func main() {
	flag.Parse()
	initClient()

	var err error
	self, err = client.GetAccountCurrentUser(context.Background())
	if err != nil {
		panic(err)
	}
	/*
		accounts, err := client.AccountsSearch(context.Background(), *user, 1)
		if err != nil {
			panic(err)
		}
		self = accounts[0]
	*/

	stats := &stats{
		DaysActive: int(time.Since(self.CreatedAt).Hours() / 24),
		Followers:  self.FollowersCount,
		Following:  self.FollowingCount,
		Toots:      make(map[string]*tootStat),
		Tags:       make(map[string]*tootStat),
		Replies:    make(map[string]*tootStat),
		Mentions:   make(map[string]int64),
		Boosts:     make(map[string]int64),
		Responses:  make(map[string]int64),
	}
	pb := &goprogressbar.ProgressBar{
		Text:  fmt.Sprintf("Loading toots for %s", self.Username),
		Total: self.StatusesCount,
		PrependTextFunc: func(p *goprogressbar.ProgressBar) string {
			return fmt.Sprintf("%d of %d", p.Current, int64(math.Max(float64(p.Current), float64(self.StatusesCount))))
		},
		Current: 0,
		Width:   40,
	}

	var pg mastodon.Pagination
	for {
		pg.Limit = 40
		statuses, err := client.GetAccountStatuses(context.Background(), self.ID, &pg)
		if err != nil {
			panic(err)
		}

		// For some reason, either because it's Pleroma or because I have too few toots,
		// `pg.MaxID` never equals `""` and we get stuck looping forever.  Add a simple
		// break condition on "no statuses fetched" to avoid the issue.
		if len(statuses) == 0 {
			break
		}

		for _, s := range statuses {
			err = parseToot(s, stats)
			if err != nil {
				panic(err)
			}

			pb.Current += 1
			pb.LazyPrint()

			if *maxToots > 0 && len(stats.Toots) >= *maxToots {
				break
			}
		}

		if *maxToots > 0 && len(stats.Toots) >= *maxToots {
			break
		}
		if pg.MaxID == "" {
			break
		}
		time.Sleep(1000 * time.Millisecond)
	}

	fmt.Printf("\n\n")
	printAccountStats(stats)
	printInteractionStats(stats)
	printTootStats(stats)
	printTagStats(stats)
}
