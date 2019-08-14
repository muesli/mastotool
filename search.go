package main

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	mastodon "github.com/mattn/go-mastodon"
	"github.com/muesli/goprogressbar"
	"github.com/spf13/cobra"
)

var (
	searchCmd = &cobra.Command{
		Use:   "search <string>",
		Short: "searches your toots",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("search requires a search token")
			}
			return search(args[0])
		},
	}
)

func search(token string) error {
	pb := &goprogressbar.ProgressBar{
		Text:  fmt.Sprintf("Searching toots for %s", token),
		Total: self.StatusesCount,
		PrependTextFunc: func(p *goprogressbar.ProgressBar) string {
			return fmt.Sprintf("%d of %d", p.Current, int64(math.Max(float64(p.Current), float64(self.StatusesCount))))
		},
		Current: 0,
		Width:   40,
	}

	var pg mastodon.Pagination
	for {
		pg.SinceID = ""
		pg.MinID = ""
		pg.Limit = 40
		statuses, err := client.GetAccountStatuses(context.Background(), self.ID, &pg)
		if err != nil {
			return fmt.Errorf("Can't retrieve statuses: %s", err)
		}

		abort := false
		for _, s := range statuses {
			if strings.Contains(strings.ToLower(cleanupContent(s.Content)), token) {
				fmt.Println("\nFound toot:", cleanupContent(s.Content))
				fmt.Println("Date:", s.CreatedAt.Format(time.RFC822))
				fmt.Println("URL:", s.URL)
				fmt.Println()
			}

			pb.Current += 1
			pb.LazyPrint()
		}

		// For some reason, either because it's Pleroma or because I have too few toots,
		// `pg.MaxID` never equals `""` and we get stuck looping forever. Add a simple
		// break condition on "no statuses fetched" to avoid the issue.
		if abort || pg.MaxID == "" || len(statuses) == 0 {
			break
		}
	}

	return nil
}

func init() {
	RootCmd.AddCommand(searchCmd)
}
