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

// Package reckon provides support for sampling and reporting on the keys and
// values in one or more redis instances
package reckon

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
)

// Options is a configuration struct that instructs the reckon pkg to sample
// the redis instance listening on a particular host/port with a specified
// number/percentage of random keys.
type Options struct {
	Host string
	Port int

	scanMode bool

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

	// Glob is a glob expression to use when running reckon in scan mode.  The expression
	// uses the glob-style paterns described at: the at: http://redis.io/commands/keys
	// An empty string will match no keys and is therefore an invalid glob expression.
	// This setting has no effect when running reckon in "sample mode".
	Glob string
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
	ErrNoKeys = errors.New("No keys are present in the configured redis instance")

	// keysExpr captures the key count from the matching line of output from
	// redis' "INFO" command
	keysExpr = regexp.MustCompile("^db\\d+:keys=(\\d+),")
)

// AnyKey is an AggregatorFunc that puts any key (regardless of key
// name or redis data type) into a generic "any-key" bucket.
func AnyKey(key string, valueType ValueType) []string {
	return []string{"any-key"}
}

// An Aggregator returns 0 or more arbitrary strings, to be used as aggregation
// groups or "buckets". For example, an Aggregator that takes the first letter
// of the key would cause reckon to aggregate stats by each letter of the
// alphabet.
type Aggregator interface {
	Groups(key string, valueType ValueType) []string
}

// The AggregatorFunc type is an adapter to allow the use of
// ordinary functions as Aggregators.  If f is a function
// with the appropriate signature, AggregatorFunc(f) is an
// Aggregator object that calls f.
type AggregatorFunc func(key string, valueType ValueType) []string

// Groups provides 0 or more groups to aggregate `key` to, when examining redis keys.
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

// A keyResult is a convenience struct that allows scanning/sampling methods to
// return a key, it's (redis) type, and/or an error, using a single channel
type keyResult struct {
	key string
	vt  ValueType
	err error
}

// scan uses a connection from the given `pool` to perform a SCAN operation on
// the redis instance. The `glob` pattern is used to limit the results of the
// SCAN operation.  The scan runs until termination (as defined by the SCAN
// documentation at: http://redis.io/commands/scan), or until a value is
// received on the supplied `quit` channel.
//
// Detail on the caveats and semantics of the SCAN operation:
// http://redis.io/commands/scan
func scan(pool *redis.Pool, glob string, quit chan struct{}) (chan *keyResult, error) {
	// A redis scan cursor starts at "0", and is complete only when the cursor value returns to "0"
	if glob == "" {
		return nil, errors.New("glob expression is empty; no keys will ever match")
	}

	var cursor string
	results := make(chan *keyResult, 2)

	go func() {
		// a conn is NOT go-routine safe, but the pool is.
		conn := pool.Get()
		defer func() {
			close(results)
			conn.Close()
		}()

		for cursor != "0" {
			if cursor == "" {
				cursor = "0"
			}

			scanVals, err := redis.Values(conn.Do("SCAN", cursor, "MATCH", glob))
			if err != nil {
				results <- &keyResult{err: err}
				return
			}

			var keys []string
			// NOTE: this Scan method does NOT perform a redis SCAN operation!  It's just
			// an unfortunately named redigo method.
			if _, err := redis.Scan(scanVals, &cursor, &keys); err != nil {
				results <- &keyResult{err: err}
				return
			}
			for _, key := range keys {
				select {
				case <-quit:
					fmt.Println("received on quit chan, exiting")
					return
				default:
					typeStr, err := redis.String(conn.Do("TYPE", key))
					if err != nil {
						results <- &keyResult{err: err}
						return
					}
					results <- &keyResult{key: key, vt: ValueType(typeStr)}
				}
			}
		}
		// scan operation complete
	}()

	return results, nil
}

// sample uses a connection from the given pool to return random redis keys,
// along with (redis) type and error information.  Results are supplied
// over the returned channel until a value is received on the supplied `quit`
// chan.
func sample(pool *redis.Pool, quit chan struct{}) (chan *keyResult, error) {

	// This buffering prevents the select below from blocking on the results
	// channel send while there's a quit channel receive waiting
	results := make(chan *keyResult, 2)

	go func() {
		// a conn is NOT go-routine safe, but the pool is.
		conn := pool.Get()

		defer close(results)
		defer conn.Close()

		for {
			select {
			case <-quit:
				fmt.Println("recieved on quit chan, exiting")
				return
			default:
				key, err := redis.String(conn.Do("RANDOMKEY"))
				if err != nil {
					results <- &keyResult{err: err}
					return
				}

				typeStr, err := redis.String(conn.Do("TYPE", key))
				if err != nil {
					results <- &keyResult{err: err}
					return
				}

				results <- &keyResult{key, ValueType(typeStr), nil}
			}
		}
	}()

	return results, nil
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
	}
	return b
}

// WithHostPort sets the host and port to use when connecting to a redis
// instance.
func WithHostPort(host string, port int) func(*Options) error {
	return func(opts *Options) error {
		if host == "" {
			return errors.New("host cannot be empty")
		}
		opts.Host = host
		opts.Port = port
		return nil
	}
}

// SampleMode returns an Options function that instructs reckon to randomly
// sample keys.
//
// `minSamples` indicates the minimum number of random keys to sample from the redis
// instance.  Note that this does not mean **unique** keys, just an absolute
// number of random keys.  Therefore, this number should be small relative to
// the number of keys in the redis instance.

// `sampleRate` indicates the percentage of the keyspace to sample.
// Accordingly, the value must be between 0.0 and 1.0.  If a non-zero value is
// given for both `sampleRate` and `minSamples`, the actual number of keys
// sampled will be the greater of the two values, once the key count has been
// calculated using the `sampleRate`.
func SampleMode(minSamples int, sampleRate float32) func(*Options) error {
	return func(opts *Options) error {
		if sampleRate < 0.0 || sampleRate > 1.0 {
			return errors.New("sample rate must be between 0.0 and 1.0")
		}

		if minSamples <= 0 && sampleRate == 0.0 {
			return errors.New("minSamples cannot be <= 0")
		}

		opts.MinSamples = minSamples
		opts.SampleRate = sampleRate
		opts.scanMode = false
		return nil
	}
}

// ScanMode returns an Options function that instructs reckon to perform a
// redis SCAN operation on the configured redis instance.
//
// The `glob` pattern is used to limit the results of the SCAN operation.
// Detail on the caveats and semantics of the SCAN operation:
// http://redis.io/commands/scan
func ScanMode(glob string) func(*Options) error {
	return func(opts *Options) error {
		if glob == "" {
			return errors.New("glob expression is empty; no keys will ever match")
		}
		opts.Glob = glob
		opts.scanMode = true
		return nil
	}
}

// Run performs the configured sampling operation against the redis instance,
// returning aggregated statistics using the provided Aggregator, as well as
// the actual key count for the redis instance.  If any errors occur, the
// operation is short-circuited, and the error is returned.  In such a case, the
// results should be considered invalid.
func Run(aggregator Aggregator, fns ...func(*Options) error) (map[string]*Results, int64, error) {

	stats := make(map[string]*Results)
	var keys int64
	var err error

	opts := &Options{
		Host:     "localhost",
		Port:     6379,
		scanMode: false,
	}

	for _, fn := range fns {
		if err == nil {
			err = fn(opts)
		}
	}

	if err != nil {
		return stats, keys, err
	}

	pool := newConnectionPool(net.JoinHostPort(opts.Host, strconv.Itoa(opts.Port)))

	conn := pool.Get()
	defer conn.Close()

	if err = conn.Err(); err != nil {
		return stats, keys, fmt.Errorf("Error connecting to the redis instance at: %s:%d : %s", opts.Host, opts.Port, err)
	}

	numSamples := opts.MinSamples

	if keys, err = keyCount(conn); err != nil {
		return stats, keys, err
	}

	fmt.Printf("redis at %s:%d has %d keys\n", opts.Host, opts.Port, keys)
	if opts.SampleRate > 0.0 {
		v := int(float32(keys) * opts.SampleRate)
		numSamples = max(max(v, numSamples), 1)
	}

	interval := numSamples / 100
	if interval == 0 {
		interval = 1
	}
	lastInterval := 0

	interval = 100
	quit := make(chan struct{})

	var results chan *keyResult
	if opts.scanMode {
		results, err = scan(pool, opts.Glob, quit)
	} else {
		results, err = sample(pool, quit)
	}

	if err != nil {
		return stats, keys, err
	}

	// Ensure that the other goroutines exit cleanly by telling them explicitly to quit
	defer func() {
		go func() { quit <- struct{}{} }()
	}()

	observed := 0
	for {
		result, ok := <-results
		if !ok {
			fmt.Printf("results channel closed; examined %d keys from redis at: %s:%d...\n", observed, opts.Host, opts.Port)
			return stats, keys, err
		}

		if result.err != nil {
			fmt.Printf("%#v\n", result.err)
			return stats, keys, err
		}

		observed++
		if observed/interval != lastInterval {
			fmt.Printf("examined %d keys from redis at: %s:%d...\n", observed, opts.Host, opts.Port)
			lastInterval = observed / interval
		}

		// Return if we're sampling, and have sampled enough keys
		if !opts.scanMode && observed == numSamples {
			return stats, keys, err
		}

		switch ValueType(result.vt) {
		case TypeString:
			if err = sampleString(result.key, conn, aggregator, stats); err != nil {
				return stats, keys, err
			}
		case TypeList:
			if err = sampleList(result.key, conn, aggregator, stats); err != nil {
				return stats, keys, err
			}
		case TypeSet:
			if err = sampleSet(result.key, conn, aggregator, stats); err != nil {
				return stats, keys, err
			}
		case TypeSortedSet:
			if err = sampleSortedSet(result.key, conn, aggregator, stats); err != nil {
				return stats, keys, err
			}
		case TypeHash:
			if err = sampleHash(result.key, conn, aggregator, stats); err != nil {
				return stats, keys, err
			}
		default:
			return stats, keys, fmt.Errorf("unknown type for redis key: %s", result.key)
		}
	}

	return stats, keys, nil
}

// newConnectionPool creates a goroutine-safe connection pool to the redis
// instance on specified host:port address. Users of the pool are responsible
// for getting connections from the pool and returning them when done (by
// calling Close) on the connection.
func newConnectionPool(address string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", address)
			if err != nil {
				return nil, err
			}
			return c, err
		},

		// Since we've already said that reckon can't run against twemproxy, go
		// ahead and use the PING cmd
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}
