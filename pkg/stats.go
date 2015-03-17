package sampler

import "strconv"

const (
	maxKeys = 10
)

// inc increments the map value for `e` by 1 in the supplied map `m`, adding an
// entry if one does not already exist
func inc(m map[string]int64, e string) {
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

// Stats stores data about sampled redis data structures
type Stats struct {
	Keys int64
	// Strings
	StringSizes map[string]int64
	StringKeys  map[string]bool

	// Sets
	SetSizes        map[string]int64
	SetElementSizes map[string]int64
	SetKeys         map[string]bool

	// Sorted Sets
	SortedSetSizes        map[string]int64
	SortedSetElementSizes map[string]int64
	SortedSetKeys         map[string]bool

	// Hashes
	HashSizes        map[string]int64
	HashElementSizes map[string]int64
	HashValueSizes   map[string]int64
	HashKeys         map[string]bool

	// Lists
	ListSizes        map[string]int64
	ListElementSizes map[string]int64
	ListKeys         map[string]bool
}

func NewStats() *Stats {
	return &Stats{
		StringSizes:           make(map[string]int64),
		StringKeys:            make(map[string]bool),
		SetSizes:              make(map[string]int64),
		SetElementSizes:       make(map[string]int64),
		SetKeys:               make(map[string]bool),
		SortedSetSizes:        make(map[string]int64),
		SortedSetElementSizes: make(map[string]int64),
		SortedSetKeys:         make(map[string]bool),
		HashSizes:             make(map[string]int64),
		HashElementSizes:      make(map[string]int64),
		HashValueSizes:        make(map[string]int64),
		HashKeys:              make(map[string]bool),
		ListSizes:             make(map[string]int64),
		ListElementSizes:      make(map[string]int64),
		ListKeys:              make(map[string]bool),
	}
}

func (s *Stats) ObserveSet(key string, length int64, member string) {
	s.Keys += 1
	inc(s.SetSizes, strconv.FormatInt(length, 10))
	inc(s.SetElementSizes, strconv.Itoa(len(member)))
	add(s.SetKeys, key, maxKeys)
}

// TODO: Remove the score?
func (s *Stats) ObserveSortedSet(key string, length int64, score float32, member string) {
	s.Keys += 1
	inc(s.SortedSetSizes, strconv.FormatInt(length, 10))
	inc(s.SortedSetElementSizes, strconv.Itoa(len(member)))
	add(s.SortedSetKeys, key, maxKeys)
}

func (s *Stats) ObserveHash(key string, length int64, field string, value string) {
	s.Keys += 1
	inc(s.HashSizes, strconv.FormatInt(length, 10))
	inc(s.HashValueSizes, strconv.Itoa(len(value)))
	inc(s.HashElementSizes, strconv.Itoa(len(field)))
	add(s.HashKeys, key, maxKeys)
}

func (s *Stats) ObserveList(key string, length int64, member string) {
	s.Keys += 1
	inc(s.ListSizes, strconv.FormatInt(length, 10))
	inc(s.ListElementSizes, strconv.Itoa(len(member)))
	add(s.ListKeys, key, maxKeys)
}

func (s *Stats) ObserveString(key, value string) {
	s.Keys += 1
	inc(s.StringSizes, strconv.Itoa(len(value)))
	add(s.StringKeys, key, maxKeys)
}
