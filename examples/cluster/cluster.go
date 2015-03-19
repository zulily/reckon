package main

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"core-gitlab.corp.zulily.com/core/sampler"
)

// keysThatStartWithA aggregates sampled keys that start with the letter 'a',
// and ignores any other key.  It aggregates such keys to a group named
// (appropriately) "starts-with-a".
func keysThatStartWithA(key string, valueType sampler.ValueType) []string {
	if strings.HasPrefix(key, "a") {
		return []string{"starts-with-a"}
	}
	return []string{}
}

// samplerResult allow us to return results OR an error on the same chan
type samplerResult struct {
	s   map[string]*sampler.Results
	err error
}

func main() {

	// Sample 100 keys from each of three redis instances, all running on different ports on localhost
	redises := []sampler.Options{
		sampler.Options{Host: "localhost", Port: 6379, NumKeys: 100},
		sampler.Options{Host: "localhost", Port: 6380, NumKeys: 100},
		sampler.Options{Host: "localhost", Port: 6381, NumKeys: 100},
	}

	aggregator := sampler.AggregatorFunc(keysThatStartWithA)

	var wg sync.WaitGroup
	results := make(chan samplerResult)

	wg.Add(len(redises))

	// Sample each redis in its own goroutine
	for _, redis := range redises {
		go func(opts sampler.Options) {
			defer wg.Done()
			s, err := sampler.Run(opts, aggregator)
			results <- samplerResult{s: s, err: err}
		}(redis)
	}

	// Collect and merge all the results
	totals := make(map[string]*sampler.Results)
	go func() {
		for r := range results {

			if r.err != nil {
				fmt.Printf("ERROR: %v\n", r.err.Error())
				return
			}
			fmt.Println("Got results back from a redis!")

			for k, v := range r.s {
				if existing, ok := totals[k]; ok {
					existing.Merge(v)
					totals[k] = existing
				} else {
					totals[k] = v
				}
			}
		}
	}()

	// render the final results when everything is complete
	wg.Wait()
	close(results)

	for k, v := range totals {
		fmt.Printf("Totals for: %s:\n", k)
		err := sampler.RenderText(v, os.Stdout)
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			return
		}
	}
}
