// iterscraper scrapes information from a website where URLs contain an incrementing integer.
// Information is retrieved from HTML5 elements, and outputted as a CSV.
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	var (
		urlTemplate = flag.String("url", "http://example.com/v/%d", "The URL you wish to scrape, containing \"%d\" where the id should be substituted")
		idLow       = flag.Int("from", 0, "The first ID that should be searched in the URL - inclusive.")
		idHigh      = flag.Int("to", 1, "The last ID that should be searched in the URL - exclusive")
		concurrency = flag.Int("concurrency", 1, "How many scrapers to run in parallel. (More scrapers are faster, but more prone to rate limiting or bandwith issues)")
		outfile     = flag.String("output", "output.csv", "Filename to export the CSV results")
		cols        = flag.String("columns", "", "QueryString of Column to Selector - ie: name=.name&address=.addr")
	)
	flag.Parse()

	if len(*cols) == 0 {
		log.Fatal("columns is a mandatory parameter")
	}

	colsParsed, err := url.ParseQuery(*cols)
	if err != nil {
		log.Fatalf("Invalid columns param: %s. QueryString format expected", *cols)
	}
	columns := make([]string, 0, 5)
	headers := make([]string, 0, 5)

	for header, selectors := range colsParsed {
		columns = append(columns, selectors[0])
		headers = append(headers, header)
	}
	// url and id are added as the first two columns.
	headers = append([]string{"url", "id"}, headers...)

	// create all tasks and send them to the channel.
	type task struct {
		url string
		id  int
	}
	tasks := make(chan task)
	go func() {
		for i := *idLow; i < *idHigh; i++ {
			tasks <- task{url: fmt.Sprintf(*urlTemplate, i), id: i}
		}
		close(tasks)
	}()

	// create workers and schedule closing results when all work is done.
	results := make(chan []string)
	var wg sync.WaitGroup
	wg.Add(*concurrency)
	go func() {
		wg.Wait()
		close(results)
	}()

	for i := 0; i < *concurrency; i++ {
		go func() {
			defer wg.Done()
			for t := range tasks {
				r, err := fetch(t.url, t.id, columns)
				if err != nil {
					log.Printf("could not fetch %v: %v", t.url, err)
					continue
				}
				for _, itemResult := range r {
					results <- itemResult
				}
			}
		}()
	}

	if err := dumpCSV(*outfile, headers, results); err != nil {
		log.Printf("could not write to %s: %v", *outfile, err)
	}
}

func fetch(url string, id int, queries []string) ([][]string, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("could not get %s: %v", url, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusTooManyRequests {
			return nil, fmt.Errorf("you are being rate limited")
		}

		return nil, fmt.Errorf("bad response from server: %s", res.Status)
	}

	// parse body with goquery.
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("could not parse page: %v", err)
	}

	// extract info we want.
	baseColumns := []string{url, strconv.Itoa(id)}

	matchesResults := make([][]string, 0, 1)
	for queryIndex, q := range queries {
		matches := doc.Find(q)
		matchesResults = append(matchesResults, make([]string, 0, matches.Length()))
		matches.Each(func(i int, s *goquery.Selection) {
			matchesResults[queryIndex] = append(matchesResults[queryIndex], strings.TrimSpace(s.Text()))
		})
	}

	r := make([][]string, len(matchesResults[0]), len(matchesResults[0]))
	for i := range matchesResults[0] {
		r[i] = make([]string, len(matchesResults), len(matchesResults))
		for j := range matchesResults {
			r[i][j] = matchesResults[j][i]
		}
		r[i] = append(baseColumns, r[i]...)
	}

	return r, nil
}

func dumpCSV(path string, headers []string, records <-chan []string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("unable to create file %s: %v", path, err)
	}
	defer f.Close()

	w := csv.NewWriter(f)

	// write headers to file.
	if err := w.Write(headers); err != nil {
		log.Fatalf("error writing record to csv: %v", err)
	}

	// write all records.
	for r := range records {
		if err := w.Write(r); err != nil {
			log.Fatalf("could not write record to csv: %v", err)
		}
	}

	w.Flush()

	// check for extra errors.
	if err := w.Error(); err != nil {
		return fmt.Errorf("writer failed: %v", err)
	}
	return nil
}
