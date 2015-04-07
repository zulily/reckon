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
package sampler

import "math"

const (
	// MaxExampleKeys sets an upper bound on the number of example keys that will
	// be captured during sampling
	MaxExampleKeys = 10
	// MaxExampleElements sets an upper bound on the number of example elements that
	// will be captured during sampling
	MaxExampleElements = 10
	// MaxExampleValues sets an upper bound on the number of example values that
	// will be captured during sampling
	MaxExampleValues = 10
)

// Statistics are basic descriptive statistics that summarize data in a frequency table
type Statistics struct {
	Mean   float64
	Min    int
	Max    int
	StdDev float64
}

// powerOfTwo returns the smallest power of two that is greater than or equal to `n`
func powerOfTwo(n int) int {
	p := 1
	for p < n {
		p = p * 2
	}
	return p
}

// ComputePowerOfTwoFreq converts a frequency map into a new frequency map,
// where each map key is the smallest power of two that is greater than or
// equal to the original map key.
func ComputePowerOfTwoFreq(m map[int]int64) map[int]int64 {
	pf := make(map[int]int64)

	for k, v := range m {
		p := powerOfTwo(k)
		if existing, ok := pf[p]; ok {
			pf[p] = existing + v
		} else {
			pf[p] = v
		}
	}

	return pf
}

// ComputeStatistics computes basic descriptive statistics about a frequency map
func ComputeStatistics(m map[int]int64) Statistics {
	stats := Statistics{}
	if len(m) == 0 {
		return stats
	}

	min := math.MaxInt32
	max := math.MinInt32
	accum, count, sd := int64(0), int64(0), float64(0)

	for k, v := range m {
		if k < min {
			min = k
		}
		if k > max {
			max = k
		}
		accum += int64(k) * v
		count += v
	}

	mean := float64(accum) / float64(count)

	for k, v := range m {
		kf, vf := float64(k), float64(v)
		sd += ((kf - mean) * (kf - mean)) * vf
	}

	return Statistics{
		Mean:   mean,
		Min:    min,
		Max:    max,
		StdDev: math.Sqrt(sd / float64(count)),
	}
}

// add adds `elem` to the "set" (a map[<type>]bool is an idiomatic golang "set") if the
// current size of the set is less than `maxsize`
func add(set map[string]bool, elem string, maxsize int) {
	if len(set) >= maxsize {
		return
	}
	set[elem] = true
}

// Results stores data about sampled redis data structures. Map keys represent
// lengths/sizes, while map values represent the frequency with which those
// lengths/sizes occurred in the sampled data. Example keys are stored in
// golang "sets", which are maps with bool values.
type Results struct {
	KeyCount int64

	// Strings
	StringSizes  map[int]int64
	StringKeys   map[string]bool
	StringValues map[string]bool

	// Sets
	SetSizes        map[int]int64
	SetElementSizes map[int]int64
	SetKeys         map[string]bool
	SetElements     map[string]bool

	// Sorted Sets
	SortedSetSizes        map[int]int64
	SortedSetElementSizes map[int]int64
	SortedSetKeys         map[string]bool
	SortedSetElements     map[string]bool

	// Hashes
	HashSizes        map[int]int64
	HashElementSizes map[int]int64
	HashValueSizes   map[int]int64
	HashKeys         map[string]bool
	HashElements     map[string]bool
	HashValues       map[string]bool

	// Lists
	ListSizes        map[int]int64
	ListElementSizes map[int]int64
	ListKeys         map[string]bool
	ListElements     map[string]bool
}

// NewResults constructs a new, zero-valued Results struct
func NewResults() *Results {
	return &Results{
		StringSizes:  make(map[int]int64),
		StringKeys:   make(map[string]bool),
		StringValues: make(map[string]bool),

		SetSizes:        make(map[int]int64),
		SetElementSizes: make(map[int]int64),
		SetKeys:         make(map[string]bool),
		SetElements:     make(map[string]bool),

		SortedSetSizes:        make(map[int]int64),
		SortedSetElementSizes: make(map[int]int64),
		SortedSetKeys:         make(map[string]bool),
		SortedSetElements:     make(map[string]bool),

		HashSizes:        make(map[int]int64),
		HashElementSizes: make(map[int]int64),
		HashValueSizes:   make(map[int]int64),
		HashKeys:         make(map[string]bool),
		HashElements:     make(map[string]bool),
		HashValues:       make(map[string]bool),

		ListSizes:        make(map[int]int64),
		ListElementSizes: make(map[int]int64),
		ListKeys:         make(map[string]bool),
		ListElements:     make(map[string]bool),
	}
}

// merge inserts all key/value pairs in `b` into `a`.  If `b` contains keys
// that are present in `a`, their values will be summed
func merge(a map[int]int64, b map[int]int64) {
	for k, v := range b {
		a[k] += v
	}
}

// union performs a set union of `a` and `b`, storing the results in `a`
func union(a map[string]bool, b map[string]bool) {
	for k := range b {
		a[k] = true
	}
}

// trim creates a new set, consisting of up to `n` random members from set `s`.
// If `len(s)` < `n`, the returned map will be of length `len(s)`. Set `s`
// remains unmodified.
func trim(s map[string]bool, n int) map[string]bool {
	t := make(map[string]bool)
	// map iteration is random in golang!
	for k := range s {
		t[k] = true
		if len(t) == n {
			break
		}
	}
	return t
}

// trimAndSum removes entries from the frequency map that comprise less than
// `threshold` % of the total, returning the sum of the **original** map
func trimAndSum(m map[int]int64, threshold float64) int64 {
	var s int64
	var sum float64
	for _, v := range m {
		s += v
	}
	sum = float64(s)
	for k, v := range m {
		if float64(v)/sum <= threshold {
			delete(m, k)
		}
	}
	return s
}

// Merge adds the results from `other` into the method receiver.  This method
// can be used to combine sampling results from multiple redis instances into a
// single result set.
func (r *Results) Merge(other *Results) {
	r.KeyCount += other.KeyCount

	// union all sets
	union(r.StringKeys, other.StringKeys)
	union(r.StringValues, other.StringValues)
	union(r.SetKeys, other.SetKeys)
	union(r.SetElements, other.SetElements)
	union(r.SortedSetKeys, other.SortedSetKeys)
	union(r.SortedSetElements, other.SortedSetElements)
	union(r.HashKeys, other.HashKeys)
	union(r.HashElements, other.HashElements)
	union(r.HashValues, other.HashValues)
	union(r.ListKeys, other.ListKeys)
	union(r.ListElements, other.ListElements)

	// merge all frequency tables
	merge(r.StringSizes, other.StringSizes)
	merge(r.SetSizes, other.SetSizes)
	merge(r.SetElementSizes, other.SetElementSizes)
	merge(r.SortedSetSizes, other.SortedSetSizes)
	merge(r.SortedSetElementSizes, other.SortedSetElementSizes)
	merge(r.HashSizes, other.HashSizes)
	merge(r.HashElementSizes, other.HashElementSizes)
	merge(r.HashValueSizes, other.HashValueSizes)
	merge(r.ListSizes, other.ListSizes)
	merge(r.ListElementSizes, other.ListElementSizes)
}

func (r *Results) observeSet(key string, length int, member string) {
	r.KeyCount++
	r.SetSizes[length] += 1
	r.SetElementSizes[len(member)] += 1
	add(r.SetKeys, key, MaxExampleKeys)
	add(r.SetElements, member, MaxExampleElements)
}

func (r *Results) observeSortedSet(key string, length int, member string) {
	r.KeyCount++
	r.SortedSetSizes[length] += 1
	r.SortedSetElementSizes[len(member)] += 1
	add(r.SortedSetKeys, key, MaxExampleKeys)
	add(r.SortedSetElements, member, MaxExampleElements)
}

func (r *Results) observeHash(key string, length int, field string, value string) {
	r.KeyCount++
	r.HashSizes[length] += 1
	r.HashValueSizes[len(value)] += 1
	r.HashElementSizes[len(field)] += 1
	add(r.HashKeys, key, MaxExampleKeys)
	add(r.HashElements, field, MaxExampleElements)
	add(r.HashValues, value, MaxExampleValues)
}

func (r *Results) observeList(key string, length int, member string) {
	r.KeyCount++
	r.ListSizes[length] += 1
	r.ListElementSizes[len(member)] += 1
	add(r.ListKeys, key, MaxExampleKeys)
	add(r.ListElements, member, MaxExampleElements)
}

func (r *Results) observeString(key, value string) {
	r.KeyCount++
	r.StringSizes[len(value)] += 1
	add(r.StringKeys, key, MaxExampleKeys)
	add(r.StringValues, value, MaxExampleValues)
}
