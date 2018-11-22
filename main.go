package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"syscall"
	"time"

	mastodon "github.com/mattn/go-mastodon"
	"github.com/muesli/goprogressbar"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	client *mastodon.Client

	topN       = flag.Int("top", 10, "shows the top N items in each category")
	maxToots   = flag.Int("recent", 0, "only account for the N most recent toots (excl replies & boosts)")
	columns    = flag.Int("columns", 80, "displays tables with N columns")
	configFile = flag.String("config", "mastodon.json", "uses the specified config file")
	// user     = flag.String("user", "@fribbledom@mastodon.social", "shows stats for this user")
)

func readPassword(prompt string) (string, error) {
	var tty io.WriteCloser
	tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		tty = os.Stdout
	} else {
		defer tty.Close()
	}

	fmt.Fprint(tty, prompt+" ")
	buf, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Fprintln(tty)

	return string(buf), err
}

func registerApp(config *Config) error {
	app, err := mastodon.RegisterApp(context.Background(), &mastodon.AppConfig{
		Server:     config.Value("instance").(string),
		ClientName: "statootstics",
		Scopes:     "read write follow",
		Website:    "",
	})
	if err != nil {
		return err
	}

	config.Set("id", app.ClientID)
	config.Set("secret", app.ClientSecret)
	return nil
}

func initClient() {
	var err error
	var instance, username, password, id, secret string
	config, err := LoadConfig(*configFile)
	if err == nil {
		instance = config.Value("instance").(string)
		username = config.Value("username").(string)
		secret = config.Value("secret").(string)
		id = config.Value("id").(string)
		if config.Value("password") != nil {
			password = config.Value("password").(string)
		}
	}

	scanner := bufio.NewScanner(os.Stdin)
	if len(instance) == 0 {
		fmt.Print("Which instance to connect to (e.g. https://mastodon.social): ")
		scanner.Scan()
		if scanner.Err() != nil {
			panic(err)
		}
		instance = scanner.Text()
	}

	if len(username) == 0 {
		fmt.Print("Username (email): ")
		scanner.Scan()
		if scanner.Err() != nil {
			panic(err)
		}
		username = scanner.Text()
	}

	config.Set("instance", instance)
	config.Set("username", username)

	if len(id) == 0 {
		err = registerApp(&config)
		if err != nil {
			panic(err)
		}

		id = config.Value("id").(string)
		secret = config.Value("secret").(string)
	}
	config.Save(*configFile)

	if len(password) == 0 {
		password, err = readPassword("Password:")
		if err != nil {
			panic(err)
		}
	}

	client = mastodon.NewClient(&mastodon.Config{
		Server:       instance,
		ClientID:     id,
		ClientSecret: secret,
	})
	err = client.Authenticate(context.Background(), username, password)
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()

	initClient()
	self, err := client.GetAccountCurrentUser(context.Background())
	if err != nil {
		panic(err)
	}
	/*
		accounts, err := client.AccountsSearch(context.Background(), *user, 1)
		if err != nil {
			panic(err)
		}
		self := accounts[0]
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
