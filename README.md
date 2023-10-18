mastotool
=========

[![Latest Release](https://img.shields.io/github/release/muesli/mastotool.svg)](https://github.com/muesli/mastotool/releases)
[![Build Status](https://github.com/muesli/mastotool/workflows/build/badge.svg)](https://github.com/muesli/mastotool/actions)
[![Go ReportCard](https://goreportcard.com/badge/muesli/mastotool)](https://goreportcard.com/report/muesli/mastotool)
[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://pkg.go.dev/github.com/muesli/mastotool)

mastotool offers a collection of command-line tools to work with your Mastodon account

## Installation

### Packages

#### Linux
- Arch Linux: [mastotool](https://aur.archlinux.org/packages/mastotool/)
- [Packages](https://github.com/muesli/mastotool/releases) in Debian & RPM formats

#### macOS
- macOS with [Homebrew](https://brew.sh/): `brew tap muesli/tap && brew install mastotool`

### Binaries
- [Binaries](https://github.com/muesli/mastotool/releases) for Linux, FreeBSD, OpenBSD, macOS & Windows

### From source

Make sure you have a working Go environment (Go 1.9 or higher is required).
See the [install instructions](https://golang.org/doc/install.html).

Compiling mastotool is easy, simply run:

    go install github.com/muesli/mastotool@latest

## Usage

```
mastotool [command]

Available Commands:
  help        Help about any command
  search      searches your toots
  stats       generates statistics about your account

Flags:
  -c, --config string   uses the specified config file (default "mastodon.json")
  -h, --help   help for mastotool

Use "mastotool [command] --help" for more information about a command.
```

### Generate Statistics

```
mastotool stats [flags]

Flags:
      --columns int   displays tables with N columns (default 80)
  -r, --recent int    only account for the N most recent toots (excl replies & boosts)
  -t, --top int       shows the top N items in each category (default 10)
```

### Search

```
mastotool search <string>
```

## Example

```
$ mastotool stats -r 100
Which instance to connect to: https://mastodon.social
Username (email): some_user@domain.tld
Password: ********

Loading toots for some_user     100 of 100 [############################>] 100.00%

Total toots: 100 (excluding replies & boosts)
Toots per day: 1.00 (account created 100 days ago)
Ratio toots/replies: 0.33
New followers per day: 7.41
New followings per day: 3.67
Likes per toot: 9.00 (total likes: 900)
Boosts per toot: 2.50 (total boosts: 250)
Replies per toot: 3.40 (total replies: 340)

Users you mentioned most                                              Interactions
----------------------------------------------------------------------------------
abc                                                                              9
def                                                                              3

Users you boosted most                                                Interactions
----------------------------------------------------------------------------------
xyz                                                                              7
zyx                                                                              2

Users that responded most                                             Interactions
----------------------------------------------------------------------------------
def                                                                             11
fed                                                                              9

Most replied-to toots                                                      Replies
----------------------------------------------------------------------------------
Some toot                                                                       20
Another toot                                                                     7

Most liked toots                                                             Likes
----------------------------------------------------------------------------------
Some toot                                                                       50
Another toot                                                                     8

Most boosted toots                                                          Boosts
----------------------------------------------------------------------------------
Some toot                                                                       10
Another toot                                                                     4

Highest scoring toots                                                        Score
----------------------------------------------------------------------------------
Some toot                                                                       80
Another toot                                                                    20

Tags used that got the most likes                                            Likes
----------------------------------------------------------------------------------
Some tag                                                                        10
Another tag                                                                      4

Tags used that got the most boosts                                          Boosts
----------------------------------------------------------------------------------
Some tag                                                                         5
Another tag                                                                      1
```
