# Sampler

> A golang package for sampling and reporting on random keys on a set of redis instances

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

While there is an [existing solution](https://github.com/antirez/redis-sampler) for
sampling a redis keyspace, the `sampler` package has a few advantages:

### Programmatic access to sampling results:

Results are returned in data structures, not just printed to stdout. This
allows for some interesting use cases, like sampling data across a cluster of
redis instances, and merging the results to get an overall picture of the
keyspaces.  We've included some sample code to do just that, in the
[examples](https://github.com/zulily/sampler/tree/master/examples/sampler-cluster).

### Arbitrary aggregation based on key and redis data type:

`sampler` affords you the ability to sample, examine, and aggregate statistics
about particular redis data types (e.g. hashes, sets, ...) and/or keys with
particular names/patterns. You can then define arbitrary aggregation "buckets",
based on the aforementioned properties of each sampled key. Details about the
aggregations [below](https://github.com/zulily/sampler#aggregation)

### Written in [Go](https://golang.org/):

We use a lot of Go. Without delving into all the reasons we love Go, suffice it
to say that Go's ability to compile to a fully-contained static binary means
that it's easy to run `sampler` on any host in our fleet.  Just `scp` a binary,
and run it.

## Aggregation

`sampler` can aggregate redis statistics by arbitrary groups, based on the
redis key and/or datatype:

Any type that implements the `Aggregator` interface can instruct `sampler`
about how to group the redis keys that it samples.  This is best illustrated
with some simple examples:

To aggregate only redis sets whose keys start with the letter `a`:

    func setsThatStartWithA(key string, valueType sampler.ValueType) []string {
      if strings.HasPrefix(key, "a") && valueType == sampler.TypeSet {
        return []string{"setsThatStartWithA"}
      }
      return []string{}
    }

To aggregate sampled keys of any redis data type that are longer than 80 characters:

    func longKeys(key string, valueType sampler.ValueType) []string {
      if len(key) > 80 {
        return []string{"long-keys"}
      }
      return []string{}
    }

## Quick Start

Get the code:

    $ go get github.com/zulily/sampler

Use the package in a binary:

    package main

    import (
      "log"
      "os"

      "github.com/zulily/sampler"
    )

    func main() {

      opts := sampler.Options{
        Host:    "localhost",
        Port:    6379,
        NumKeys: 10000,
      }

      stats, err := sampler.Run(opts, sampler.AggregatorFunc(sampler.AnyKey))
      if err != nil {
        log.Fatalf("ERROR: %v\n", err)
      }

      for k, v := range stats {
        log.Printf("stats for: %s\n", k)
        if err := sampler.RenderText(v, os.Stdout); err != nil {
          log.Fatalf("ERROR: %v\n", err)
        }
      }
    }


## Examples

Some example binaries are included that demonstrate various usages of the
`sampler` package, the simplest of which samples from a single redis instance.

To sample 10K keys from a redis instance running on `yourserver:6379` and print
the results to `stdout`:

    $ sampler-single -host=yourserver -port=6379 -num-keys=10000

## Limitations

Since `sampler` makes use of redis' `RANDOMKEY` and `INFO` commands, it is not
able to sample data via a [twemproxy](https://github.com/twitter/twemproxy)
proxy, since twemproxy implements a subset of the redis protocol that does not
include these commands.

However, instead of sampling through a proxy, you can easily run `sampler`
against multiple redis instances, and merge the results.  We include code
that does just that in the
[examples](https://github.com/zulily/sampler/tree/master/examples/sampler-cluster).
