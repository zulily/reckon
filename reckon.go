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
package reckon

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/garyburd/redigo/redis"
)

// Options is a configuration struct that instructs the reckon pkg to sample
// the redis instance listening on a particular host/port with a specified
// number/percentage of random keys.
type Options struct {
	Host string
	Port int

	// MinSamples indicates the minimum number of random keys to sample from the redis
	// instance.  Note that this does not mean **unique** keys, just an absolute
	// number of random keys.  Therefore, this number should be small relative to
	// the number of keys in the redis instance.
	MinSamples int

	// SampleRate indicates the percentage of the keyspace to sample.
	// Accordingly, values should be between 0.0 and 1.0.  If a non-zero value is
	// given for both `SampleRate` and `MinSamples`, the actual number of keys
	// sampled will be the greater of the two values, once the key count has been
	// calculated using the `SampleRate`.
	SampleRate float32
}

// A ValueType represents the various data types that redis can store. The
// string representation of a ValueType matches what is returned from redis'
// `TYPE` command.
type ValueType string

var (
	// TypeString represents a redis string value
	TypeString ValueType = "string"
	// TypeSortedSet represents a redis sorted set value
	TypeSortedSet ValueType = "zset"
	// TypeSet represents a redis set value
	TypeSet ValueType = "set"
	// TypeHash represents a redis hash value
	TypeHash ValueType = "hash"
	// TypeList represents a redis list value
	TypeList ValueType = "list"
	// TypeUnknown means that the redis value type is undefined, and indicates an error
	TypeUnknown ValueType = "unknown"

	// ErrNoKeys is the error returned when a specified redis instance contains
	// no keys, or the key count could not be determined
	ErrNoKeys error = errors.New("No keys are present in the configured redis instance")

	// keysExpr captures the key count from the matching line of output from
	// redis' "INFO" command
	keysExpr = regexp.MustCompile("^db\\d+:keys=(\\d+),")
)

// AnyKey is an AggregatorFunc that puts any sampled key (regardless of key
// name or redis data type) into a generic "any" bucket.
func AnyKey(key string, valueType ValueType) []string {
	return []string{"any"}
}

// An Aggregator returns 0 or more arbitrary strings, to be used during a
// sampling operation as aggregation groups or "buckets". For example, an
// Aggregator that takes the first letter of the key would cause reckon to
// aggregate stats by each letter of the alphabet
type Aggregator interface {
	Groups(key string, valueType ValueType) []string
}

// The AggregatorFunc type is an adapter to allow the use of
// ordinary functions as Aggregators.  If f is a function
// with the appropriate signature, AggregatorFunc(f) is an
// Aggregator object that calls f.
type AggregatorFunc func(key string, valueType ValueType) []string

// Groups provides 0 or more groups to aggregate `key` to, when sampling redis keys.
func (f AggregatorFunc) Groups(key string, valueType ValueType) []string {
	return f(key, valueType)
}

// flush is a convenience func for flushing a redis pipeline, receiving the
// replies, and returning them, along with any error
func flush(conn redis.Conn) ([]interface{}, error) {
	return redis.Values(conn.Do(""))
}

// ensureEntry is a convenience func for obtaining the Stats instance for the
// specified `group`, creating a new one if no such entry already exists
func ensureEntry(m map[string]*Results, group string, init func() *Results) *Results {
	var stats *Results
	var ok bool
	if stats, ok = m[group]; !ok {
		stats = init()
		m[group] = stats
	}
	return stats
}

// randomKey obtains a random redis key and its ValueType from the supplied redis connection
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

// keyCount obtains a the number of keys in the redis instance.
func keyCount(conn redis.Conn) (count int64, err error) {
	resp, err := redis.String(conn.Do("INFO"))
	if err != nil {
		return count, err
	}

	for _, str := range strings.Split(resp, "\n") {
		if matches := keysExpr.FindStringSubmatch(str); len(matches) >= 2 {
			if count, err = strconv.ParseInt(matches[1], 10, 64); err == nil && count != 0 {
				return count, nil
			}
			return count, ErrNoKeys
		}
	}

	return 0, ErrNoKeys
}

func sampleString(key string, conn redis.Conn, aggregator Aggregator, stats map[string]*Results) error {
	val, err := redis.String(conn.Do("GET", key))
	if err != nil {
		return err
	}

	for _, agg := range aggregator.Groups(key, TypeString) {
		s := ensureEntry(stats, agg, NewResults)
		s.observeString(key, val)
	}
	return nil
}

func sampleList(key string, conn redis.Conn, aggregator Aggregator, stats map[string]*Results) error {
	// TODO: Let's not always get the first element, like the orig. reckon
	conn.Send("LLEN", key)
	conn.Send("LRANGE", key, 0, 0)
	replies, err := flush(conn)
	if err != nil {
		return err
	}

	if len(replies) >= 2 {
		l, err := redis.Int(replies[0], nil)
		ms, err := redis.Strings(replies[1], err)
		if err != nil {
			return err
		}

		for _, g := range aggregator.Groups(key, TypeList) {
			s := ensureEntry(stats, g, NewResults)
			s.observeList(key, l, ms[0])
		}
	}
	return nil
}

func sampleSet(key string, conn redis.Conn, aggregator Aggregator, stats map[string]*Results) error {
	conn.Send("SCARD", key)
	conn.Send("SRANDMEMBER", key)
	replies, err := flush(conn)
	if err != nil {
		return err
	}

	if len(replies) >= 2 {
		l, err := redis.Int(replies[0], nil)
		m, err := redis.String(replies[1], err)
		if err != nil {
			return err
		}

		for _, g := range aggregator.Groups(key, TypeSet) {
			s := ensureEntry(stats, g, NewResults)
			s.observeSet(key, l, m)
		}
	}
	return nil
}

func sampleSortedSet(key string, conn redis.Conn, aggregator Aggregator, stats map[string]*Results) error {
	conn.Send("ZCARD", key)
	// TODO: Let's not always get the first element, like the orig. sampler
	conn.Send("ZRANGE", key, 0, 0)
	replies, err := flush(conn)
	if err != nil {
		return err
	}

	if len(replies) >= 2 {
		l, err := redis.Int(replies[0], nil)
		ms, err := redis.Strings(replies[1], err)
		if err != nil {
			return err
		}

		for _, g := range aggregator.Groups(key, TypeSortedSet) {
			s := ensureEntry(stats, g, NewResults)
			s.observeSortedSet(key, l, ms[0])
		}
	}
	return nil
}

func sampleHash(key string, conn redis.Conn, aggregator Aggregator, stats map[string]*Results) error {
	conn.Send("HLEN", key)
	conn.Send("HKEYS", key)
	replies, err := flush(conn)
	if err != nil {
		return err
	}

	if len(replies) >= 2 {
		for _, g := range aggregator.Groups(key, TypeHash) {

			// TODO: Let's not always get the first hash field, like the orig. sampler
			l, err := redis.Int(replies[0], nil)
			fields, err := redis.Strings(replies[1], err)
			if err != nil {
				return err
			}
			val, err := redis.String(conn.Do("HGET", key, fields[0]))
			if err != nil {
				return err
			}
			s := ensureEntry(stats, g, NewResults)
			s.observeHash(key, l, fields[0], val)
		}
	}
	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

// Run performs the configured sampling operation against the redis instance,
// returning aggregated statistics using the provided Aggregator, as well as
// the actual key count for the redis instance.  If any errors occur, the
// sampling is short-circuited, and the error is returned.  In such a case, the
// results should be considered invalid.
func Run(opts Options, aggregator Aggregator) (map[string]*Results, int64, error) {

	stats := make(map[string]*Results)
	var err error
	var keys int64

	if opts.SampleRate < 0.0 || opts.SampleRate > 1.0 {
		return stats, keys, errors.New("SampleRate must be between 0.0 and 1.0")
	}

	if opts.MinSamples <= 0 && opts.SampleRate == 0.0 {
		return stats, keys, errors.New("MinSamples cannot be 0")
	}

	conn, err := redis.Dial("tcp", net.JoinHostPort(opts.Host, strconv.Itoa(opts.Port)))
	if err != nil {
		return stats, keys, fmt.Errorf("Error connecting to the redis instance at: %s:%d : %s", opts.Host, opts.Port, err.Error())
	}

	numSamples := opts.MinSamples

	if keys, err = keyCount(conn); err != nil {
		return stats, keys, err
	} else {
		fmt.Printf("redis at %s:%d has %d keys\n", opts.Host, opts.Port, keys)
		if opts.SampleRate > 0.0 {
			v := int(float32(keys) * opts.SampleRate)
			numSamples = max(max(v, numSamples), 1)
		}
	}

	interval := numSamples / 100
	if interval == 0 {
		interval = 1
	}
	lastInterval := 0

	for i := 0; i < numSamples; i++ {
		key, vt, err := randomKey(conn)
		if err != nil {
			return stats, keys, err
		}

		if i/interval != lastInterval {
			fmt.Printf("sampled %d keys from redis at: %s:%d...\n", i, opts.Host, opts.Port)
			lastInterval = i / interval
		}

		switch ValueType(vt) {
		case TypeString:
			if err = sampleString(key, conn, aggregator, stats); err != nil {
				return stats, keys, err
			}
		case TypeList:
			if err = sampleList(key, conn, aggregator, stats); err != nil {
				return stats, keys, err
			}
		case TypeSet:
			if err = sampleSet(key, conn, aggregator, stats); err != nil {
				return stats, keys, err
			}
		case TypeSortedSet:
			if err = sampleSortedSet(key, conn, aggregator, stats); err != nil {
				return stats, keys, err
			}
		case TypeHash:
			if err = sampleHash(key, conn, aggregator, stats); err != nil {
				return stats, keys, err
			}
		default:
			return stats, keys, fmt.Errorf("unknown type for redis key: %s", key)
		}
	}
	return stats, keys, nil
}
