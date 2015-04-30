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
	"bytes"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"

	"github.com/zulily/reckon"
)

// Address represents a host:port address.
type Address struct {
	Host string
	Port int
}

// Addresses is a slice of Address instances.  It implements the flag.Value
// interface, and thus can be used with the Var func in the flag pkg
type Addresses []Address

func (a *Addresses) String() string {
	var buf bytes.Buffer
	for _, addr := range *a {
		buf.WriteString(net.JoinHostPort(addr.Host, strconv.Itoa(addr.Port)))
	}
	return buf.String()
}

// Set is part of the flag.Value interface to allow Addresses to be used as
// flag values
func (a *Addresses) Set(value string) error {
	host, port, err := net.SplitHostPort(value)
	if err != nil {
		return err
	}

	p, err := strconv.Atoi(port)
	if err != nil {
		return err
	}
	*a = append(*a, Address{Host: host, Port: p})
	return nil
}

// reckonResult allow us to return results OR an error on the same chan
type reckonResult struct {
	s        map[string]*reckon.Results
	keyCount int64
	err      error
}

type options struct {
	redises    Addresses
	minSamples int
	sampleRate float64
}

var (
	opts options
)

func main() {

	flag.Float64Var(&opts.sampleRate, "sample-rate", 0.1, "The percentage of the keyspace to sample on each redis")
	flag.IntVar(&opts.minSamples, "min-samples", 100, "minimum number of keys to sample on each redis")
	flag.Var(&opts.redises, "redis", "host:port address of a redis instance to sample (may be specified multiple times)")
	flag.Parse()

	// Sample 100 keys from each of three redis instances, all running on different ports on localhost
	var reckonOpts []reckon.Options

	for _, redis := range opts.redises {
		opt := reckon.Options{Host: redis.Host, Port: redis.Port, MinSamples: opts.minSamples, SampleRate: float32(opts.sampleRate)}
		reckonOpts = append(reckonOpts, opt)
	}

	aggregator := reckon.AggregatorFunc(reckon.AnyKey)

	var wg sync.WaitGroup
	results := make(chan reckonResult)
	wg.Add(len(reckonOpts))

	// Sample each redis in its own goroutine
	for _, instanceOpts := range reckonOpts {
		go func(opts reckon.Options) {
			defer wg.Done()
			log.Printf("Sampling %d keys from redis at: %s:%d...\n", opts.MinSamples, opts.Host, opts.Port)
			s, keyCount, err := reckon.Run(opts, aggregator)
			results <- reckonResult{s: s, keyCount: keyCount, err: err}
		}(instanceOpts)
	}

	// Collect and merge all the results
	totals := make(map[string]*reckon.Results)
	totalKeyCount := int64(0)

	go func() {
		for r := range results {
			if r.err != nil {
				panic(r.err)
			}
			log.Println("Got results back from a redis instance!")

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

	// render the final results to HTML when everything is complete
	wg.Wait()
	close(results)

	log.Printf("total key count: %d\n", totalKeyCount)
	for k, v := range totals {

		v.Name = k
		if f, err := os.Create(fmt.Sprintf("output-%s.html", k)); err != nil {
			panic(err)
		} else {
			defer f.Close()
			log.Printf("Rendering totals for: %s to %s:\n", k, f.Name())
			if err := reckon.RenderHTML(v, f); err != nil {
				panic(err)
			}
		}

	}
}
