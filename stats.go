package sampler

import "math"

const (
	maxKeys = 10
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
		StdDev: math.Sqrt(sd / float64(len(m))),
	}
}

// inc increments the map value for `e` by 1 in the supplied map `m`, adding an
// entry if one does not already exist
func inc(m map[int]int64, e int) {
	if existing, ok := m[e]; ok {
		m[e] = existing + 1
	} else {
		m[e] = 1
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
	StringSizes map[int]int64
	StringKeys  map[string]bool

	// Sets
	SetSizes        map[int]int64
	SetElementSizes map[int]int64
	SetKeys         map[string]bool

	// Sorted Sets
	SortedSetSizes        map[int]int64
	SortedSetElementSizes map[int]int64
	SortedSetKeys         map[string]bool

	// Hashes
	HashSizes        map[int]int64
	HashElementSizes map[int]int64
	HashValueSizes   map[int]int64
	HashKeys         map[string]bool

	// Lists
	ListSizes        map[int]int64
	ListElementSizes map[int]int64
	ListKeys         map[string]bool
}

// NewResults constructs a new, zero-valued Results struct
func NewResults() *Results {
	return &Results{
		StringSizes:           make(map[int]int64),
		StringKeys:            make(map[string]bool),
		SetSizes:              make(map[int]int64),
		SetElementSizes:       make(map[int]int64),
		SetKeys:               make(map[string]bool),
		SortedSetSizes:        make(map[int]int64),
		SortedSetElementSizes: make(map[int]int64),
		SortedSetKeys:         make(map[string]bool),
		HashSizes:             make(map[int]int64),
		HashElementSizes:      make(map[int]int64),
		HashValueSizes:        make(map[int]int64),
		HashKeys:              make(map[string]bool),
		ListSizes:             make(map[int]int64),
		ListElementSizes:      make(map[int]int64),
		ListKeys:              make(map[string]bool),
	}
}

// merge inserts all key/value pairs in `b` into `a`.  If `b` contains keys
// that are present in `a`, their values will be summed
func merge(a map[int]int64, b map[int]int64) {
	for k, v := range b {
		if existing, ok := a[k]; ok {
			a[k] = existing + v
		} else {
			a[k] = v
		}
	}
}

// union performs a set union of `a` and `b`, storing the results in `a`
func union(a map[string]bool, b map[string]bool) {
	for k := range b {
		if !a[k] {
			a[k] = true
		}
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
	union(r.SetKeys, other.SetKeys)
	union(r.SortedSetKeys, other.SortedSetKeys)
	union(r.HashKeys, other.HashKeys)
	union(r.ListKeys, other.ListKeys)

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
	inc(r.SetSizes, length)
	inc(r.SetElementSizes, len(member))
	add(r.SetKeys, key, maxKeys)
}

func (r *Results) observeSortedSet(key string, length int, member string) {
	r.KeyCount++
	inc(r.SortedSetSizes, length)
	inc(r.SortedSetElementSizes, len(member))
	add(r.SortedSetKeys, key, maxKeys)
}

func (r *Results) observeHash(key string, length int, field string, value string) {
	r.KeyCount++
	inc(r.HashSizes, length)
	inc(r.HashValueSizes, len(value))
	inc(r.HashElementSizes, len(field))
	add(r.HashKeys, key, maxKeys)
}

func (r *Results) observeList(key string, length int, member string) {
	r.KeyCount++
	inc(r.ListSizes, length)
	inc(r.ListElementSizes, len(member))
	add(r.ListKeys, key, maxKeys)
}

func (r *Results) observeString(key, value string) {
	r.KeyCount++
	inc(r.StringSizes, len(value))
	add(r.StringKeys, key, maxKeys)
}
