/*
 * Copyright (C) 2015 zulily, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package main

import (
	"log"
	"os"
	"strings"
	"sync"

	"github.com/zulily/sampler"
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
	s        map[string]*sampler.Results
	keyCount int64
	err      error
}

func main() {

	// Sample 100 keys from each of three redis instances, all running on different ports on localhost
	redises := []sampler.Options{
		sampler.Options{Host: "localhost", Port: 6379, MinSamples: 100},
		sampler.Options{Host: "localhost", Port: 6380, MinSamples: 100},
		sampler.Options{Host: "localhost", Port: 6381, MinSamples: 100},
	}

	aggregator := sampler.AggregatorFunc(sampler.AnyKey)

	var wg sync.WaitGroup
	results := make(chan samplerResult)

	wg.Add(len(redises))

	// Sample each redis in its own goroutine
	for _, redis := range redises {
		go func(opts sampler.Options) {
			defer wg.Done()
			log.Printf("Sampling %d keys from redis at: %s:%d...\n", opts.MinSamples, opts.Host, opts.Port)
			s, keyCount, err := sampler.Run(opts, aggregator)
			results <- samplerResult{s: s, keyCount: keyCount, err: err}
		}(redis)
	}

	// Collect and merge all the results
	totals := make(map[string]*sampler.Results)
	totalKeyCount := int64(0)

	go func() {
		for r := range results {
			if r.err != nil {
				log.Fatalf("ERROR: %v\n", r.err.Error())
			}
			log.Printf("Got results back from a redis!")

			totalKeyCount += r.keyCount
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

	log.Printf("total key count: %d\n", totalKeyCount)
	for k, v := range totals {
		log.Printf("Totals for: %s:\n", k)
		err := sampler.RenderText(v, os.Stdout)
		if err != nil {
			log.Fatalf("ERROR: %v\n", err)
		}
	}
}
