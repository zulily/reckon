package sampler

import (
	"fmt"
	"net"
	"strconv"

	"github.com/garyburd/redigo/redis"
)

type Options struct {
	Host    string
	Port    int
	NumKeys int
}

// A ValueType represents the various data types that redis can store. The
// string representation of a ValueType matches what is returned from redis'
// `TYPE` command.
type ValueType string

var (
	TypeString    ValueType = "string"
	TypeSortedSet ValueType = "zset"
	TypeSet       ValueType = "set"
	TypeHash      ValueType = "hash"
	TypeList      ValueType = "list"
	TypeUnknown   ValueType = "unknown"
)

// An Aggregator returns 0 or more arbitrary strings, to be used by a
// Sampler as aggregation buckets for stats. For example, an Aggregator
// that takes the first litter of the key would cause a Sampler to aggregate
// stats by each letter of the alphabet
type Aggregator interface {
	Groups(key string, valueType ValueType) []string
}

// The AggregatorFunc type is an adapter to allow the use of
// ordinary functions as Aggregators.  If f is a function
// with the appropriate signature, AggregatorFunc(f) is an
// Aggregator object that calls f.
type AggregatorFunc func(key string, valueType ValueType) []string

// Groups provides 0 or more groups to aggregate `key` to when aggregating across redis keys.
func (f AggregatorFunc) Groups(key string, valueType ValueType) []string {
	return f(key, valueType)
}

// flush is a convenience fn for flushing a redis pipeline, receiving the
// replies, and returning them, along with any error
func flush(conn redis.Conn) ([]interface{}, error) {
	return redis.Values(conn.Do(""))
}

// ensureEntry is a convenience func for obtaining the Stats instance for the
// specified `group`, creating a new one if no such entry already exists
func ensureEntry(m map[string]*Stats, group string, init func() *Stats) *Stats {
	var stats *Stats
	var ok bool
	if stats, ok = m[group]; !ok {
		stats = init()
		m[group] = stats
	}
	return stats
}

// randomKey obtains a random key and it's ValueType from the supplied redis connection
func randomKey(conn redis.Conn) (key string, vt ValueType, err error) {
	key, err = redis.String(conn.Do("RANDOMKEY"))
	if err != nil {
		return key, TypeUnknown, err
	}

	typeStr, err := redis.String(conn.Do("TYPE", key))
	if err != nil {
		return key, TypeUnknown, err
	}

	return key, ValueType(typeStr), nil
}

func sampleString(key string, conn redis.Conn, aggregator Aggregator, stats map[string]*Stats) error {
	val, err := redis.String(conn.Do("GET", key))
	if err != nil {
		return err
	}

	for _, agg := range aggregator.Groups(key, TypeString) {
		s := ensureEntry(stats, agg, NewStats)
		s.ObserveString(key, val)
	}
	return nil
}

func sampleList(key string, conn redis.Conn, aggregator Aggregator, stats map[string]*Stats) error {
	// TODO: Let's not always get the first element, like the orig. sampler
	conn.Send("LLEN", key)
	conn.Send("LRANGE", key, 0, 0)
	replies, err := flush(conn)
	if err != nil {
		return err
	}

	if len(replies) >= 2 {
		l, err := redis.Int(replies[0], err)
		if err != nil {
			return err
		}
		ms, err := redis.Strings(replies[1], err)
		if err != nil {
			return err
		}

		for _, g := range aggregator.Groups(key, TypeList) {
			s := ensureEntry(stats, g, NewStats)
			s.ObserveList(key, l, ms[0])
		}
	}
	return nil
}

func sampleSet(key string, conn redis.Conn, aggregator Aggregator, stats map[string]*Stats) error {
	conn.Send("SCARD", key)
	conn.Send("SRANDMEMBER", key)
	replies, err := flush(conn)
	if err != nil {
		return err
	}

	if len(replies) >= 2 {
		l, err := redis.Int(replies[0], err)
		if err != nil {
			return err
		}
		m, err := redis.String(replies[1], err)
		if err != nil {
			return err
		}

		for _, g := range aggregator.Groups(key, TypeSet) {
			s := ensureEntry(stats, g, NewStats)
			s.ObserveSet(key, l, m)
		}
	}
	return nil
}

func sampleSortedSet(key string, conn redis.Conn, aggregator Aggregator, stats map[string]*Stats) error {
	conn.Send("ZCARD", key)
	// TODO: Let's not always get the first element, like the orig. sampler
	conn.Send("ZRANGE", key, 0, 0)
	replies, err := flush(conn)
	if err != nil {
		return err
	}

	if len(replies) >= 2 {
		l, err := redis.Int(replies[0], err)
		if err != nil {
			return err
		}
		ms, err := redis.Strings(replies[1], err)
		if err != nil {
			return err
		}

		for _, g := range aggregator.Groups(key, TypeSortedSet) {
			s := ensureEntry(stats, g, NewStats)
			s.ObserveSortedSet(key, l, 0.0, ms[0])
		}
	}
	return nil
}

func sampleHash(key string, conn redis.Conn, aggregator Aggregator, stats map[string]*Stats) error {
	conn.Send("HLEN", key)
	conn.Send("HKEYS", key)
	replies, err := flush(conn)
	if err != nil {
		return err
	}

	if len(replies) >= 2 {
		for _, g := range aggregator.Groups(key, TypeHash) {

			// TODO: Let's not always get the first hash field, like the orig. sampler
			l, err := redis.Int(replies[0], err)
			if err != nil {
				return err
			}
			fields, err := redis.Strings(replies[1], err)
			if err != nil {
				return err
			}
			val, err := redis.String(conn.Do("HGET", key, fields[0]))
			if err != nil {
				return err
			}
			s := ensureEntry(stats, g, NewStats)
			s.ObserveHash(key, l, fields[0], val)
		}
	}
	return nil
}

// Sample runs the configured sampling operation against the redis instance, aggregating
// statistics using the provided Aggregator.  If any errors occurr, the sampling is short-circuited,
// and the error is returned.  In such a case, the results should be considered invalid.
func Sample(opts Options, aggregator Aggregator) (map[string]*Stats, error) {

	allStats := make(map[string]*Stats)
	var err error

	conn, err := redis.Dial("tcp", net.JoinHostPort(opts.Host, strconv.Itoa(opts.Port)))
	if err != nil {
		return allStats, err
	}

	for i := 0; i < opts.NumKeys; i++ {
		key, vt, err := randomKey(conn)
		if err != nil {
			return allStats, err
		}

		switch ValueType(vt) {
		case TypeString:
			err = sampleString(key, conn, aggregator, allStats)
			if err != nil {
				return allStats, err
			}
		case TypeList:
			err = sampleList(key, conn, aggregator, allStats)
			if err != nil {
				return allStats, err
			}
		case TypeSet:
			err = sampleSet(key, conn, aggregator, allStats)
			if err != nil {
				return allStats, err
			}
		case TypeSortedSet:
			err = sampleSortedSet(key, conn, aggregator, allStats)
			if err != nil {
				return allStats, err
			}
		case TypeHash:
			err = sampleHash(key, conn, aggregator, allStats)
			if err != nil {
				return allStats, err
			}
		default:
			return allStats, fmt.Errorf("unknown type for redis key: %s", key)
		}
	}
	return allStats, nil
}
