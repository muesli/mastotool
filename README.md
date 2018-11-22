statootstics
============

Mastodon Statistics Generator

## Installation

Make sure you have a working Go environment (Go 1.7 or higher is required).
See the [install instructions](http://golang.org/doc/install.html).

To install statootstics, simply run:

    go get github.com/muesli/statootstics

## Usage

```
$ statootstics -help
Usage of ./statootstics:
  -columns int
        displays tables with N columns (default 80)
  -config string
        uses the specified config file (default "mastodon.json")
  -recent int
        only account for the N most recent toots (excl replies & boosts)
  -top int
        shows the top N items in each category (default 10)

$ statootstics
Loading toots for some_user     100 of 100 [#>---------------------------] 100.00%

Total toots: 100 (excluding replies & boosts)
Toots per day: 1.00 (account created 100 days ago)
Ratio toots/replies: 0.33
New followers per day: 7.41
New followings per day: 3.67
Likes per toot: 9.00 (total likes: 900)
Boosts per toot: 2.50 (total boosts: 250)

Users you mentioned most                                              Interactions
----------------------------------------------------------------------------------
abc                                                                              3

Users you boosted most                                                Interactions
----------------------------------------------------------------------------------
xyz                                                                              7

Most replied-to toots                                                      Replies
----------------------------------------------------------------------------------
Some toot                                                                       20

Most liked toots                                                             Likes
----------------------------------------------------------------------------------
Some toot                                                                       50

Most boosted toots                                                          Boosts
----------------------------------------------------------------------------------
Some toot                                                                       10

Highest scoring toots                                                        Score
----------------------------------------------------------------------------------
Some toot                                                                       80

Tags used that got the most likes                                            Likes
----------------------------------------------------------------------------------
Some tag                                                                        10

Tags used that got the most boosts                                          Boosts
----------------------------------------------------------------------------------
Some tag                                                                         5
```

## Development

[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://godoc.org/github.com/mueslistatootstics)
[![Build Status](https://travis-ci.org/muesli/statootstics.svg?branch=master)](https://travis-ci.org/muesli/statootstics)
[![Go ReportCard](http://goreportcard.com/badge/muesli/statootstics)](http://goreportcard.com/report/muesli/statootstics)
