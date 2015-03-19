package sampler

import "math"

const (
	maxKeys = 10
)

type Statistics struct {
	mean   float64
	min    int
	max    int
	stdDev float64
}

// computes descriptive statistics about our frequency maps
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
		mean:   mean,
		min:    min,
		max:    max,
		stdDev: math.Sqrt(sd / float64(len(m))),
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

// Stats stores data about sampled redis data structures. Map keys represent
// lengths/sizes, while map values represent the frequency with which those
// lengths/sizes occurred in the sampled data.
type Stats struct {
	Keys int64
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

func NewStats() *Stats {
	return &Stats{
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

func (s *Stats) ObserveSet(key string, length int, member string) {
	s.Keys++
	inc(s.SetSizes, length)
	inc(s.SetElementSizes, len(member))
	add(s.SetKeys, key, maxKeys)
}

// TODO: Remove the score?
func (s *Stats) ObserveSortedSet(key string, length int, score float32, member string) {
	s.Keys++
	inc(s.SortedSetSizes, length)
	inc(s.SortedSetElementSizes, len(member))
	add(s.SortedSetKeys, key, maxKeys)
}

func (s *Stats) ObserveHash(key string, length int, field string, value string) {
	s.Keys++
	inc(s.HashSizes, length)
	inc(s.HashValueSizes, len(value))
	inc(s.HashElementSizes, len(field))
	add(s.HashKeys, key, maxKeys)
}

func (s *Stats) ObserveList(key string, length int, member string) {
	s.Keys++
	inc(s.ListSizes, length)
	inc(s.ListElementSizes, len(member))
	add(s.ListKeys, key, maxKeys)
}

func (s *Stats) ObserveString(key, value string) {
	s.Keys++
	inc(s.StringSizes, len(value))
	add(s.StringKeys, key, maxKeys)
}
