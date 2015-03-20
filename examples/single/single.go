package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/zulily/sampler"
)

// aggregateByFirst letter aggregates redis stats according the first letter of the redis key
func aggregateByFirstLetter(key string, valueType sampler.ValueType) []string {
	return []string{key[:1]}
}

// setsThatStartWithA ignores any sampled key that is not a set or does not
// start with the letter 'a'.  It aggregates keys that DO meet this criteria up
// to a group named (appropriately) "setsThatStartWithA".
func setsThatStartWithA(key string, valueType sampler.ValueType) []string {
	if strings.HasPrefix(key, "a") && valueType == sampler.TypeSet {
		return []string{"setsThatStartWithA"}
	}
	return []string{}
}

func main() {

	opts := sampler.Options{}
	flag.StringVar(&opts.Host, "host", "localhost", "the hostname of the redis server")
	flag.IntVar(&opts.Port, "port", 6379, "the port of the redis server")
	flag.IntVar(&opts.NumKeys, "num-keys", 50, "number of keys to sample (should be <= the number of keys in the redis instance")
	flag.Parse()

	stats, err := sampler.Run(opts, sampler.AggregatorFunc(setsThatStartWithA))

	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	for k, v := range stats {
		fmt.Printf("stats for: %s\n", k)
		err := sampler.RenderText(v, os.Stdout)
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			return
		}
	}
}
