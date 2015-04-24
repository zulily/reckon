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
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/zulily/reckon"
)

// aggregateByFirst letter aggregates redis stats according the first letter of the redis key
func aggregateByFirstLetter(key string, valueType reckon.ValueType) []string {
	return []string{key[:1]}
}

// setsThatStartWithA ignores any sampled key that is not a set or does not
// start with the letter 'a'.  It aggregates keys that DO meet this criteria up
// to a group named (appropriately) "setsThatStartWithA".
func setsThatStartWithA(key string, valueType reckon.ValueType) []string {
	if strings.HasPrefix(key, "a") && valueType == reckon.TypeSet {
		return []string{"setsThatStartWithA"}
	}
	return []string{}
}

func main() {

	opts := reckon.Options{}
	flag.StringVar(&opts.Host, "host", "localhost", "the hostname of the redis server")
	flag.IntVar(&opts.Port, "port", 6379, "the port of the redis server")
	flag.IntVar(&opts.MinSamples, "min-samples", 50, "number of random samples to take (should be <= the number of keys in the redis instance")
	flag.Parse()

	stats, keyCount, err := reckon.Run(opts, reckon.AggregatorFunc(reckon.AnyKey))

	if err != nil {
		log.Fatalf("ERROR: %v\n", err)
	}

	log.Printf("total key count: %d\n", keyCount)
	for k, v := range stats {
		log.Printf("stats for: %s\n", k)

		v.Name = k
		if f, err := os.Create(fmt.Sprintf("output/output-%s.html", k)); err != nil {
			panic(err)
		} else {
			defer f.Close()
			if err := reckon.RenderHTML(v, f); err != nil {
				panic(err)
			}
		}
	}
}
