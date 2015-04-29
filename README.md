# Reckon

> A Go package for sampling and reporting on random keys on a set of redis instances

Inspired/influenced by [redis-sampler](https://github.com/antirez/redis-sampler)
from [antirez](https://github.com/antirez), the author of redis.

## Background

We love redis here at [zulily](https://github.com/zulily/). We store millions
of keys across many redis instances, and we've built our own internal distributed
cache on top of it.

One problem with running a large, distributed cache using redis is the opaque
nature of the keyspaces; it's hard to tell what the composition of your redis
dataset is, especially when you've got multiple codebases or teams using the
same redis instance(s), or you're sharding your dataset over a large number of
redis instances.

While there are some [existing](https://github.com/antirez/redis-sampler)
[solutions](https://github.com/snmaynard/redis-audit) for sampling a redis
keyspace, the `reckon` package has a few advantages:

### Programmatic access to sampling results:

Results are returned in data structures, not just printed to stdout or a file.
This is what allows a user of reckon to sample data across a cluster of redis
instances and merge the results to get an overall picture of the keyspaces.
We've included some sample code to do just that, in the
[examples](https://github.com/zulily/reckon/tree/master/examples/reckoning-multiple-instances).

### Aggregation

`reckon` also allows you to define arbitrary buckets based on the name of the
sampled key and/or the redis data type (hash, set, list, etc.). During
sampling, `reckon` compiles statistics about the various redis data types, and
aggregates those statistics according to the buckets you defined.

Any type that implements the `Aggregator` interface can instruct `reckon` as to
how to aggregate the redis keys that it samples. This is best illustrated with some
simple, contrived examples:

To aggregate only redis sets whose keys start with the letter a:

    func setsThatStartWithA(key string, valueType reckon.ValueType) []string {
      if strings.HasPrefix(key, "a") && valueType == reckon.TypeSet {
        return []string{"setsThatStartWithA"}
      }
      return []string{}
    }

To aggregate sampled keys of any redis data type that are longer than 80 characters:

    func longKeys(key string, valueType reckon.ValueType) []string {
    if len(key) > 80 {
      return []string{"long-keys"}
      }
      return []string{}
    }

### Reports

When you are done sampling, aggregating, and/or combining the results produced
by `reckon` you can easily produce a report of the findings in either plain-text
or static HTML. An example HTML report is shown below:

![Sample HTML report](https://github.com/zulily/reckon/blob/master/random-sets.png)


## Quick Start

Get the code:

    $ go get github.com/zulily/reckon

Use one of the provided example binaries to sample from a redis instance and
output results to static HTML files in the current directory:

    $ reckoning-single-instance -host=localhost -port=6379 \
        -sample-rate=0.1 -min-samples=100

Or to sample from multiple instances:

    $ reckoning-multiple-instances -sample-rate=0.1 \
        -redis=localhost:6379 \
        -redis=localhost:6380 \
        -redis=localhost:6381

Or, use the package in your own binary:

    package main

    import (
      "log"
      "os"

      "github.com/zulily/reckon"
    )

    func main() {

      opts := reckon.Options{
        Host:       "localhost",
        Port:       6379,
        MinSamples: 10000,
      }

      stats, keyCount, err := reckon.Run(opts, reckon.AggregatorFunc(reckon.AnyKey))
      if err != nil {
        panic(err)
      }

      log.Printf("total key count: %d\n", keyCount)
      for k, v := range stats {
        log.Printf("stats for: %s\n", k)

        v.Name = k
        if f, err := os.Create(fmt.Sprintf("output-%s.html", k)); err != nil {
          panic(err)
        } else {
          defer f.Close()
          log.Printf("Rendering totals for: '%s' to %s:\n", k, f.Name())
          if err := reckon.RenderHTML(v, f); err != nil {
            panic(err)
          }
        }
      }
    }

## Limitations

Since `reckon` makes use of redis' `RANDOMKEY` and `INFO` commands, it is not
able to sample data via a [twemproxy](https://github.com/twitter/twemproxy)
proxy, since twemproxy implements a subset of the redis protocol that does not
include these commands.

However, instead of sampling through a proxy, you can easily run `reckon`
against multiple redis instances, and merge the results.  We include code
that does just that in the
[examples](https://github.com/zulily/reckon/tree/master/examples/reckoning-multiple-instances).
