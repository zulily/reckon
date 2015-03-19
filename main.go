package main

import (
	"flag"
	"log"
	"strings"

	"github.com/kr/pretty"

	sampler "core-gitlab.corp.zulily.com/core/sampler/pkg"
)

// An FirstLetterAggregator aggregates redis stats by the first letter of the key
func aggregateByFirstLetter(key, valueType string) []string {
	return []string{key[:1]}
}

// aggregateByNamespace aggregates redis stats by the reasoning KB namespace of the key.
// Reasoning KB keys typically follow the pattern: `namespace:kbkey:{id}`, where the
// curly brace `{` and `}` chars are literals present in the key.
func aggregateByNamespace(key string, valueType sampler.ValueType) []string {
	splits := strings.SplitN(key, ":", 2)
	if len(splits) > 1 {
		return []string{splits[0]}
	}
	return []string{}
}

func aggregateByValueType(key string, valueType sampler.ValueType) []string {
	return []string{string(valueType)}
}

func main() {

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	opts := sampler.Options{}
	flag.StringVar(&opts.Host, "host", "localhost", "the hostname of the redis server")
	flag.IntVar(&opts.Port, "port", 6379, "the port of the redis server")
	flag.IntVar(&opts.NumKeys, "num-keys", 50, "number of keys to sample (should be <= the number of keys in the redis instance")
	flag.Parse()

	aggregator := sampler.AggregatorFunc(aggregateByNamespace)

	stats, err := sampler.Sample(opts, aggregator)

	if err != nil {
		log.Fatal(err.Error())
	}

	for k, v := range stats {
		//		fmt.Printf("%s: %#v\n", k, v)
		pretty.Print(k, v, "\n")
		pretty.Print(sampler.ComputeStatistics(v.SetSizes), "\n")
		pretty.Print(sampler.ComputePowerOfTwoFreq(v.SetSizes), "\n")
		//fmt.Printf("stats for: %s\n", k)
		//sampler.RenderStats(v)
	}
}
