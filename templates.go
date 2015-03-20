package sampler

import (
	"fmt"
	"io"
	"text/template"
)

func summarize(m map[int]int64) int64 {
	// trim off entries that constitute < 1% of the total
	return trimAndSum(m, 0.01)
}

func fmtFloat(n float64) string {
	return fmt.Sprintf("%.2f", n)
}

func percentage(n, total int64) string {
	return fmt.Sprintf("%.2f%%", 100.0*float64(n)/float64(total))
}

const (
	statsTempl = `
{{define "base"}}
# of keys: {{.KeyCount}}

{{ if .StringKeys }}
--- Strings ({{summarize .StringSizes}}) ---
{{template "exampleKeys" .StringKeys}}
Sizes ({{template "stats" .StringSizes}}):
{{template "freq" .StringSizes}}
{{template "freq" power .StringSizes}}{{end}}

{{ if .SetKeys }}
--- Sets ({{summarize .SetSizes}}) ---
{{template "exampleKeys" .SetKeys}}
Sizes ({{template "stats" .SetSizes}}):
{{template "freq" .SetSizes}}
^2 Sizes:{{template "freq" power .SetSizes}}
Element Sizes:{{template "freq" .SetElementSizes}}
Element ^2 Sizes:{{template "freq" power .SetElementSizes}}{{end}}

{{ if .SortedSetKeys }}
--- Sorted Sets ({{summarize .SortedSetSizes}}) ---
{{template "exampleKeys" .SortedSetKeys}}
Sizes ({{template "stats" .SortedSetSizes}}):
{{template "freq" .SortedSetSizes}}
^2 Sizes:{{template "freq" power .SortedSetSizes}}
Element Sizes ({{template "stats" .SortedSetElementSizes}}):
{{template "freq" .SortedSetElementSizes}}
Element ^2 Sizes:{{template "freq" power .SortedSetElementSizes}}{{end}}

{{ if .HashKeys }}
--- Hashes ({{summarize .HashSizes}}) ---
{{template "exampleKeys" .HashKeys}}
Sizes ({{template "stats" .HashSizes}}):
{{template "freq" .HashSizes}}
^2 Sizes:{{template "freq" power .HashSizes}}
Element Sizes ({{template "stats" .HashElementSizes}}):
{{template "freq" .HashElementSizes}}
^2 Element Sizes:{{template "freq" power .HashElementSizes}}
Value Sizes ({{template "stats" .HashValueSizes}}):
{{template "freq" .HashValueSizes}}
^2 Value Sizes:{{template "freq" power .HashValueSizes}}{{end}}

{{ if .ListKeys }}
--- Lists ({{summarize .ListSizes}}) ---
{{template "exampleKeys" .ListKeys}}
Sizes ({{template "stats" .ListSizes}}):
{{template "freq" .ListSizes}}
^2 Sizes:{{template "freq" power .ListSizes}}
Element Sizes ({{template "stats" .ListElementSizes}}):
{{template "freq" .ListElementSizes}}
^2 Element Sizes{{template "freq" power .ListElementSizes}}
{{end}}{{end}}

{{define "stats"}}{{ with stats . }}min: {{.Min}} max: {{.Max}} mean: {{fmtFloat .Mean}} std dev: {{fmtFloat .StdDev}}{{end}}{{end}}

{{define "exampleKeys"}}Example Keys:
{{range $k, $v := .}} {{$k}}
{{end}}{{end}}

{{define "freq"}}
{{ $ss := summarize . }}{{ range $s, $c := .}} {{$s}}: {{$c}} ({{percentage $c $ss }})
{{end}}{{end}}
`
)

// RenderText renders a Results instance to the supplied io.Writer
func RenderText(s *Results, out io.Writer) error {

	s.StringKeys = trim(s.StringKeys, 10)
	s.SetKeys = trim(s.SetKeys, 10)
	s.SortedSetKeys = trim(s.SortedSetKeys, 10)
	s.HashKeys = trim(s.HashKeys, 10)
	s.ListKeys = trim(s.ListKeys, 10)

	fm := template.FuncMap{
		"summarize":  summarize,
		"percentage": percentage,
		"power":      ComputePowerOfTwoFreq,
		"stats":      ComputeStatistics,
		"fmtFloat":   fmtFloat,
	}
	t := template.Must(template.New("output").Funcs(fm).Parse(statsTempl))
	return t.ExecuteTemplate(out, "base", s)
}
