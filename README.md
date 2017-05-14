# iterscraper

[![Build Status](https://travis-ci.org/philipithomas/iterscraper.svg?branch=master)](https://travis-ci.org/philipithomas/iterscraper)

A basic package used for scraping information from a website where URLs contain an incrementing integer. Information is retrieved from HTML5 elements, and outputted as a CSV.

Thanks [Francesc](https://github.com/campoy) for featuring this repo in episode #1 of [Just For Func](https://twitter.com/justforfunc). [Watch The Video](https://www.youtube.com/watch?list=PL64wiCrrxh4Jisi7OcCJIUpguV_f5jGnZ&v=eIWFnNz8mF4) or [Review Francesc's pull request](https://github.com/philipithomas/iterscraper/pull/1).

## Flags

Flags are all optional except `columns`, and are set with a single dash on the command line, e.g.

```
iterscraper \
-url            "http://foo.com/%d" \
-from           1                   \
-to             10                  \
-concurrency    10                  \
-output         foo.csv             \
```

For an explanation of the options, type `iterscraper -help`

General usage of iterscraper:

```
  -concurrency int
        How many scrapers to run in parallel. (More scrapers are faster, but more prone to rate limiting or bandwith issues) (default 1)
  -from int
        The first ID that should be searched in the URL - inclusive.
  -output string
        Filename to export the CSV results (default "output.csv")
  -to int
        The last ID that should be searched in the URL - exclusive (default 1)
  -url string
        The URL you wish to scrape, containing "%d" where the id should be substituted (default "http://example.com/v/%d")
  -columns string
        QueryString mapping Column title to Selector - ie: "name=.name&address=#address-container span"
```

## NOTE ABOUT SELECTORS

If there are multiple matches for a given selector in the same page, it is expected that other selectors return same amount of matches. Useful for lists, tables, etc.

## URL Structure

Successive pages must look like:

```
http://example.com/foo/1/bar
http://example.com/foo/2/bar
http://example.com/foo/3/bar
```

iterscraper would then accept the url in the following style, in `Printf` style such that numbers may be substituted into the url:

```
http://example.com/foo/%d/bar
```

## Installation

Building the source requires the [Go programming language](https://golang.org/doc/install) and the [Glide](http://glide.sh) package manager.

```
# Dependency is GoQuery
go get github.com/PuerkitoBio/goquery
# Get and build source
go get github.com/philipithomas/iterscraper
# If your $PATH is configured correctly, you can call it directly
iterscraper [flags]

```


## Errata

* This is purpose-built for some internal scraping. It's not meant to be the scraping tool for every user case, but you're welcome to modify it for your purposes
* On a `429 - too many requests` error, the app logs and continues, ignoring the request.
* The package will [follow up to 10 redirects](https://golang.org/pkg/net/http/#Get)
* On a `404 - not found` error, the system will log the miss, then continue. It is not exported to the CSV.
