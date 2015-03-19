package sampler

import (
	"fmt"
	"io"
	"text/template"
)

func sumValues(m map[int]int64) int64 {
	var sum int64
	for _, v := range m {
		sum += v
	}
	return sum
}

func percentage(n, total int64) string {
	return fmt.Sprintf("%.2f%%", 100.0*float64(n)/float64(total))
}

const (
	statsTempl = `# of keys: {{.KeyCount}}

{{if .StringKeys}}
{{ $ss := sumValues .StringSizes }}
Strings ({{$ss}}):
  Sample Keys:
	{{ range $k, $v := .StringKeys}} {{$k}}
	{{end}}
	Sizes:
	{{ range $s, $c := .StringSizes}} {{$s}}: {{$c}} ({{percentage $c $ss }}),
	{{end}}
{{end}}

{{if .SetKeys}}
{{ $ss := sumValues .SetSizes}}
Sets ({{sumValues .SetSizes}}):
  Sample Keys:
	{{ range $k, $v := .SetKeys}} {{$k}}
	{{end}}
	Sizes:
	{{ range $s, $c := .SetSizes}} {{$s}}: {{$c}} ({{percentage $c $ss }}),
	{{end}}
	Power of 2 Sizes:
	{{ range $s, $c := power .SetSizes}} {{$s}}: {{$c}} ({{percentage $c $ss }}),
	{{end}}
	Element Sizes:
	{{ range $s, $c := .SetElementSizes}} {{$s}}: {{$c}} ({{percentage $c $ss }}),
	{{end}}
	Power of 2 Element Sizes:
	{{ range $s, $c := power .SetElementSizes}} {{$s}}: {{$c}} ({{percentage $c $ss }}),
	{{end}}
{{end}}

{{if .SortedSetKeys}}
SortedSets ({{sumValues .SortedSetSizes}}):
  Sample Keys:
	{{ range $k, $v := .SortedSetKeys}} {{$k}}
	{{end}}
	Sizes:
	{{ range $s, $c := .SortedSetSizes}} {{$s}}: {{$c}},
	{{end}}
	Element Sizes:
	{{ range $s, $c := .SortedSetElementSizes}} {{$s}}: {{$c}},
	{{end}}
{{end}}

{{if .HashKeys}}
Hashs ({{sumValues .HashSizes}}):
  Sample Keys:
	{{ range $k, $v := .HashKeys}} {{$k}}
	{{end}}
	Sizes:
	{{ range $s, $c := .HashSizes}} {{$s}}: {{$c}},
	{{end}}
	Element Sizes:
	{{ range $s, $c := .HashElementSizes}} {{$s}}: {{$c}},
	{{end}}
	Value Sizes:
	{{ range $s, $c := .HashValueSizes}} {{$s}}: {{$c}},
	{{end}}
{{end}}`
)

// RenderText renders a Results instance to the supplied io.Writer
func RenderText(s *Results, out io.Writer) error {

	s.StringKeys = trim(s.StringKeys, 5)
	s.SetKeys = trim(s.SetKeys, 5)
	s.SortedSetKeys = trim(s.SortedSetKeys, 5)
	s.HashKeys = trim(s.HashKeys, 5)
	s.ListKeys = trim(s.ListKeys, 5)

	fm := template.FuncMap{
		"sumValues":  sumValues,
		"percentage": percentage,
		"power":      ComputePowerOfTwoFreq,
	}
	t := template.Must(template.New("stats").Funcs(fm).Parse(statsTempl))
	return t.Execute(out, s)
}
