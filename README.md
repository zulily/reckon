# Sampler

> A golang package for sampling and reporting on random keys on a set of redis instances

Highly influenced by [redis-sampler](https://github.com/antirez/redis-sampler)
by [antirez](https://github.com/antirez), the creator of redis.

## Background

We love redis here at [zulily](https://github.com/zulily/). We store millions
of keys across many redis instances, including our own internal distributed
cache.

One problem with running a large, distributed cache using redis is the opaque
nature of the keyspaces; it's hard to tell what the composition of your redis
dataset is, especially when you've got multiple codebases or teams using the
same redis instance(s), or you're sharding your dataset over a large number of
redis instances.

While there is an [existing solution](https://github.com/zulily/) for sampling
a redis keyspace, we wanted to make a few improvements:

### Written in [go](https://golang.org/):

We use a lot of Go. Without delving into all the reasons we love Go, suffice it
to say that Go's ability to compile to a fully-contained static binary means
that it's easy to run `sampler` on any host in our fleet.  Just `scp` a binary,
and run it.

### Programmatic access to sampling results:

Results are returned in data structures, not just printed to stdout. This
allows for some interesting use cases, like sampling data across a cluster of
redis instances, and merging the results to get an overall picture of the
keyspaces.  We've included some sample code to do just that, in the
[examples](https://github.com/zulily/sampler/tree/master/examples/sampler-cluster).

### Arbitrary aggregation based on key and redis type:

Sometimes we only want to examine redis hashes, other times, we care more about
keys that have a certain naming convention. `sampler` allows you to define
arbitrary aggregation "buckets", based on the name and redis data type of each
sampled key. Details about the aggregations below:

## Aggregation

`sampler` can aggregate redis statistics by arbitrary groups, based on the
redis key and/or datatype:

Any type that implements the `Aggregator` interface can instruct `sampler`
about how to group the redis keys that it samples.  This is best illustrated
with an example:

To aggregate only redis sets whose keys start with the letter `a`:

    func setsThatStartWithA(key string, valueType sampler.ValueType) []string {
      if strings.HasPrefix(key, "a") && valueType == sampler.TypeSet {
        return []string{"setsThatStartWithA"}
      }
      return []string{}
    }

To aggregate sampled keys of any redis datatype that are longer than 80 characters:

    func longKeys(key string, valueType sampler.ValueType) []string {
      if len(key) > 80 {
        return []string{"long-keys"}
      }
      return []string{}
    }

## Quick Start

Get the code:

    $ go get github.com/zulily/sampler

To sample 10K keys from a redis instance running on `yourserver:6379` and
print the results to `stdout`:

    $ sampler-single -host=yourserver -port=6379 -num-keys=10000

## Limitations

Since `sampler` makes use of redis' `RANDOMKEY` and `INFO` commands, it is not
able to sample data via a [twemproxy](https://github.com/twitter/twemproxy)
proxy, since twemproxy implements a subset of the redis protocol that does not
include these commands.

However, instead of sampling through a proxy, you can easily run `sampler`
against multiple redis instances, and merge the results.  We include an example
that does just that in the
[examples](https://github.com/zulily/sampler/tree/master/examples/sampler-cluster).
