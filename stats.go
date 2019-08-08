package main

import (
	"context"
	"fmt"
	"html"
	"sort"
	"strconv"
	"strings"

	mastodon "github.com/mattn/go-mastodon"
	"github.com/microcosm-cc/bluemonday"
	"github.com/muesli/gotable"
)

var (
	stripper = bluemonday.StrictPolicy()
)

const (
	SortByLikes = iota
	SortByBoosts
	SortByScore
	SortByReplies
)

type tootStat struct {
	Likes   int64
	Boosts  int64
	Replies int64
}

type stats struct {
	DaysActive int
	Followers  int64
	Following  int64
	Toots      map[string]*tootStat
	Tags       map[string]*tootStat
	Replies    map[string]*tootStat
	Mentions   map[string]int64
	Boosts     map[string]int64
	Responses  map[string]int64
}

func cleanupContent(content string) string {
	// clean up toot for terminal output
	content = strings.Replace(content, "<br>", "\n", -1)
	content = strings.Replace(content, "<p>", "\n", -1)
	content = strings.Replace(content, "</p>", "", -1)
	content = html.UnescapeString(stripper.Sanitize(content))
	content = strings.TrimSpace(strings.Replace(content, "\n", " ", -1))

	return content
}

func parseToot(status *mastodon.Status, stats *stats) error {
	// handle mentions
	for _, m := range status.Mentions {
		stats.Mentions[m.Acct]++
	}

	// handle boosts
	if status.Reblog != nil {
		stats.Boosts[status.Reblog.Account.Acct]++
		return nil
	}

	var replies int64

	// parse tags
	if status.InReplyToID == nil {
		contexts, err := client.GetStatusContext(context.Background(), status.ID)
		if err != nil {
			return err
		}

		// handle replies for this status
		for _, d := range contexts.Descendants {
			if d.Account.ID == self.ID {
				continue
			}
			replies++
			stats.Responses[d.Account.Acct]++
		}

		for _, t := range status.Tags {
			tag := strings.ToLower(t.Name)

			stat, ok := stats.Tags[tag]
			if ok {
				stat.Likes += status.FavouritesCount
				stat.Boosts += status.ReblogsCount
				stat.Replies += replies
			} else {
				stat = &tootStat{
					Likes:   status.FavouritesCount,
					Boosts:  status.ReblogsCount,
					Replies: replies,
				}

				stats.Tags[tag] = stat
			}
		}
	}

	content := cleanupContent(status.Content)

	// handle replies
	if status.InReplyToID != nil {
		stats.Replies[content] = &tootStat{
			Likes:   status.FavouritesCount,
			Boosts:  status.ReblogsCount,
			Replies: replies,
		}
	} else {
		stats.Toots[content] = &tootStat{
			Likes:   status.FavouritesCount,
			Boosts:  status.ReblogsCount,
			Replies: replies,
		}
	}

	return nil
}

type StatSorter struct {
	SortKey int
	Key     []string
	Stats   []*tootStat
}

func (a StatSorter) Len() int {
	return len(a.Stats)
}

func (a StatSorter) Swap(i, j int) {
	a.Key[i], a.Key[j] = a.Key[j], a.Key[i]
	a.Stats[i], a.Stats[j] = a.Stats[j], a.Stats[i]
}

func (a StatSorter) Less(i, j int) bool {
	switch a.SortKey {
	case SortByReplies:
		return a.Stats[i].Replies < a.Stats[j].Replies
	case SortByLikes:
		return a.Stats[i].Likes < a.Stats[j].Likes
	case SortByBoosts:
		return a.Stats[i].Boosts < a.Stats[j].Boosts
	case SortByScore:
		return (a.Stats[i].Boosts*3)+a.Stats[i].Likes <
			(a.Stats[j].Boosts*3)+a.Stats[j].Likes
	default:
		panic("SortKey is incorrect")
	}
}

type kv struct {
	Key   string
	Value int64
}

func printTable(cols []string, emptyText string, data []kv) {
	sort.Slice(data, func(i, j int) bool {
		return data[i].Value > data[j].Value
	})

	col1 := *columns - len(cols[1])
	col2 := len(cols[1])
	tab := gotable.NewTable(cols,
		[]int64{-int64(col1), int64(col2)},
		emptyText)

	for i, kv := range data {
		if i >= *topN {
			break
		}
		if len(kv.Key) > col1-4 {
			kv.Key = kv.Key[:col1-4] + "..."
		}

		tab.AppendRow([]interface{}{kv.Key, strconv.FormatInt(int64(kv.Value), 10)})
	}
	tab.Print()
	fmt.Println()
}

func printTootTable(cols []string, emptyText string, toots []string, tootStats []*tootStat, sortKey int) {
	sort.Sort(sort.Reverse(StatSorter{sortKey, toots, tootStats}))

	var ss []kv
	for k, v := range toots {
		switch sortKey {
		case SortByReplies:
			if tootStats[k].Replies == 0 {
				continue
			}
			ss = append(ss, kv{v, tootStats[k].Replies})
		case SortByLikes:
			if tootStats[k].Likes == 0 {
				continue
			}
			ss = append(ss, kv{v, tootStats[k].Likes})
		case SortByBoosts:
			if tootStats[k].Boosts == 0 {
				continue
			}
			ss = append(ss, kv{v, tootStats[k].Boosts})
		case SortByScore:
			score := (tootStats[k].Boosts * 3) + tootStats[k].Likes
			if score == 0 {
				continue
			}
			ss = append(ss, kv{v, score})
		}
	}

	printTable(cols, emptyText, ss)
}

func printAccountStats(stats *stats) {
	var likes, boosts, replies int64
	for _, t := range stats.Toots {
		likes += t.Likes
		boosts += t.Boosts
		replies += t.Replies
	}

	fmt.Printf("Total toots: %d (excluding replies & boosts)\n", len(stats.Toots))
	fmt.Printf("Toots per day: %.2f (account created %d days ago)\n",
		float64(len(stats.Toots))/float64(stats.DaysActive),
		stats.DaysActive)
	fmt.Printf("Ratio toots/replies: %.2f\n",
		float64(len(stats.Toots))/float64(len(stats.Replies)))
	fmt.Printf("New followers per day: %.2f\n",
		float64(stats.Followers)/float64(stats.DaysActive))
	fmt.Printf("New followings per day: %.2f\n",
		float64(stats.Following)/float64(stats.DaysActive))
	fmt.Printf("Likes per toot: %.2f (total likes: %d)\n",
		float64(likes)/float64(len(stats.Toots)),
		likes)
	fmt.Printf("Boosts per toot: %.2f (total boosts: %d)\n",
		float64(boosts)/float64(len(stats.Toots)),
		boosts)
	fmt.Printf("Replies per toot: %.2f (total replies: %d)\n",
		float64(replies)/float64(len(stats.Toots)),
		replies)
	fmt.Println()
}

func printInteractionStats(stats *stats) {
	var ss []kv
	for k, v := range stats.Mentions {
		ss = append(ss, kv{k, v})
	}
	printTable([]string{"Users you mentioned most", "Interactions"},
		"No interactions found.",
		ss)

	ss = []kv{}
	for k, v := range stats.Boosts {
		ss = append(ss, kv{k, v})
	}
	printTable([]string{"Users you boosted most", "Interactions"},
		"No interactions found.",
		ss)

	ss = []kv{}
	for k, v := range stats.Responses {
		ss = append(ss, kv{k, v})
	}
	printTable([]string{"Users that responded most", "Interactions"},
		"No interactions found.",
		ss)
}

func printTootStats(stats *stats) {
	var toots []string
	var tootStats []*tootStat
	for toot, s := range stats.Toots {
		toots = append(toots, toot)
		tootStats = append(tootStats, s)
	}

	// most replied-to toots
	printTootTable([]string{"Most replied-to toots", "Replies"},
		"No toots found.",
		toots, tootStats, SortByReplies)

	// most liked toots
	printTootTable([]string{"Most liked toots", "Likes"},
		"No toots found.",
		toots, tootStats, SortByLikes)

	// most boosted toots
	printTootTable([]string{"Most boosted toots", "Boosts"},
		"No toots found.",
		toots, tootStats, SortByBoosts)

	// highest scoring toots
	printTootTable([]string{"Highest scoring toots", "Score"},
		"No toots found.",
		toots, tootStats, SortByScore)
}

func printTagStats(stats *stats) {
	var tags []string
	var tagStats []*tootStat
	for tag, s := range stats.Tags {
		tags = append(tags, tag)
		tagStats = append(tagStats, s)
	}

	// most liked tags
	printTootTable([]string{"Tags used that got the most likes", "Likes"},
		"No toots found.",
		tags, tagStats, SortByLikes)

	// most boosted tags
	printTootTable([]string{"Tags used that got the most boosts", "Boosts"},
		"No toots found.",
		tags, tagStats, SortByBoosts)
}
