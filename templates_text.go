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

const (
	statsTempl = `
{{define "base"}}
# of keys sampled: {{.KeyCount}}

{{ if .StringKeys }}
--- Strings ({{summarize .StringSizes}}) ---
{{template "exampleKeys" .StringKeys}}
{{template "exampleValues" .StringValues}}
Sizes ({{template "stats" .StringSizes}}):
{{template "freq" .StringSizes}}
^2 Sizes:{{template "freq" power .StringSizes}}{{end}}

{{ if .SetKeys }}
--- Sets ({{summarize .SetSizes}}) ---
{{template "exampleKeys" .SetKeys}}
Sizes ({{template "stats" .SetSizes}}):
{{template "freq" .SetSizes}}
^2 Sizes:{{template "freq" power .SetSizes}}
{{template "exampleElements" .SetElements}}
Element Sizes:{{template "freq" .SetElementSizes}}
Element ^2 Sizes:{{template "freq" power .SetElementSizes}}{{end}}

{{ if .SortedSetKeys }}
--- Sorted Sets ({{summarize .SortedSetSizes}}) ---
{{template "exampleKeys" .SortedSetKeys}}
Sizes ({{template "stats" .SortedSetSizes}}):
{{template "freq" .SortedSetSizes}}
^2 Sizes:{{template "freq" power .SortedSetSizes}}
{{template "exampleElements" .SortedSetElements}}
Element Sizes ({{template "stats" .SortedSetElementSizes}}):
{{template "freq" .SortedSetElementSizes}}
Element ^2 Sizes:{{template "freq" power .SortedSetElementSizes}}{{end}}

{{ if .HashKeys }}
--- Hashes ({{summarize .HashSizes}}) ---
{{template "exampleKeys" .HashKeys}}
Sizes ({{template "stats" .HashSizes}}):
{{template "freq" .HashSizes}}
^2 Sizes:{{template "freq" power .HashSizes}}
{{template "exampleElements" .HashElements}}
Element Sizes ({{template "stats" .HashElementSizes}}):
{{template "freq" .HashElementSizes}}
^2 Element Sizes:{{template "freq" power .HashElementSizes}}
{{template "exampleValues" .HashValues}}
Value Sizes ({{template "stats" .HashValueSizes}}):
{{template "freq" .HashValueSizes}}
^2 Value Sizes:{{template "freq" power .HashValueSizes}}{{end}}

{{ if .ListKeys }}
--- Lists ({{summarize .ListSizes}}) ---
{{template "exampleKeys" .ListKeys}}
Sizes ({{template "stats" .ListSizes}}):
{{template "freq" .ListSizes}}
^2 Sizes:{{template "freq" power .ListSizes}}
{{template "exampleElements" .ListElements}}
Element Sizes ({{template "stats" .ListElementSizes}}):
{{template "freq" .ListElementSizes}}
^2 Element Sizes{{template "freq" power .ListElementSizes}}
{{end}}{{end}}

{{define "stats"}}{{ with stats . }}min: {{.Min}} max: {{.Max}} mean: {{fmtFloat .Mean}} std dev: {{fmtFloat .StdDev}}{{end}}{{end}}

{{define "exampleKeys"}}Example Keys:
{{range $k, $v := .}} {{$k}}
{{end}}{{end}}

{{define "exampleValues"}}Example Values:
{{range $k, $v := .}} {{$k}}
{{end}}{{end}}

{{define "exampleElements"}}Example Elements:
{{range $k, $v := .}} {{$k}}
{{end}}{{end}}

{{define "freq"}}
{{ $ss := summarize . }}{{ range $s, $c := .}} {{$s}}: {{$c}} ({{percentage $c $ss }})
{{end}}{{end}}
`
)
