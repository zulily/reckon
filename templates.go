package sampler

import (
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

const (
	statsTempl = `# of keys: {{.Keys}}

{{if .StringKeys}}
Strings ({{sumValues .StringSizes}}):
  Sample Keys:
	{{ range $k, $v := .StringKeys}} {{$k}}
	{{end}}

	Sizes:
	{{ range $size, $count := .StringSizes}} {{$size}}: {{$count}},
	{{end}}
{{end}}

{{if .SetKeys}}
Sets ({{sumValues .SetSizes}}):
  Sample Keys:
	{{ range $k, $v := .SetKeys}} {{$k}}
	{{end}}

	Sizes:
	{{ range $size, $count := .SetSizes}} {{$size}}: {{$count}},
	{{end}}

	Element Sizes:
	{{ range $s, $c := .SetElementSizes}} {{$s}}: {{$c}},
	{{end}}
{{end}}

{{if .SortedSetKeys}}
SortedSets ({{sumValues .SortedSetSizes}}):
  Sample Keys:
	{{ range $k, $v := .SortedSetKeys}} {{$k}}
	{{end}}

	Sizes:
	{{ range $size, $count := .SortedSetSizes}} {{$size}}: {{$count}},
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
	{{ range $size, $count := .HashSizes}} {{$size}}: {{$count}},
	{{end}}
	Element Sizes:
	{{ range $s, $c := .HashElementSizes}} {{$s}}: {{$c}},
	{{end}}
	Value Sizes:
	{{ range $s, $c := .HashValueSizes}} {{$s}}: {{$c}},
	{{end}}
{{end}}`
)

// RenderText renders a Stats instance to the supplied io.Writer
func RenderText(s *Results, out io.Writer) error {
	t := template.Must(template.New("stats").Funcs(template.FuncMap{"sumValues": sumValues}).Parse(statsTempl))
	return t.Execute(out, s)
}
