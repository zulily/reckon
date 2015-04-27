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
	htmlTmpl = `
{{define "base"}}

<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <!-- The above 3 meta tags *must* come first in the head; any other head content must come *after* these tags -->
    <title>reckoning</title>

    <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.4/css/bootstrap.min.css">
  </head>
  <body>
    <div class="container">
      <div class="jumbotron">
        <h1>{{.Name}} <small>{{.KeyCount}} keys</small></h1>
      </div>

			{{ if .StringKeys }}
			  <h1>Strings <small>{{summarize .StringSizes}}</small> </h1>
				<div class="panel panel-default">
					<div class="panel-body">
						<h3>Example keys:</h3> {{template "examples" .StringKeys}}
						<h3>Value Sizes: {{template "stats" .StringSizes}}</h3>
						{{template "freq" .StringSizes}}
						<h3>2<sup><var>n</var></sup> Value Sizes:</h3> {{template "freq" power .StringSizes}}
					</div>
				</div>
			{{ end }}

			{{ if .SetKeys }}
			  <h1>Sets <small>{{summarize .SetSizes}}</small> </h1>
				<div class="panel panel-default">
					<div class="panel-body">
						<h3>Example keys:</h3> {{template "examples" .SetKeys}}
						<h3>Sizes: {{template "stats" .SetSizes}}</h3>
						{{template "freq" .SetSizes}}
						<h3>2<sup><var>n</var></sup> Sizes:</h3> {{template "freq" power .SetSizes}}

						<h3>Example elements:</h3> {{template "examples" .SetElements}}
						<h3>Element Sizes: {{template "stats" .SetElementSizes}}</h3>
						{{template "freq" .SetElementSizes}}
						<h3>2<sup><var>n</var></sup> Element Sizes:</h3> {{template "freq" power .SetElementSizes}}
					</div>
				</div>
			{{ end }}

			{{ if .SortedSetKeys }}
			  <h1>Sorted Sets <small>{{summarize .SortedSetSizes}}</small> </h1>
				<div class="panel panel-default">
					<div class="panel-body">
						<h3>Example keys:</h3> {{template "examples" .SortedSetKeys}}
						<h3>Sizes: {{template "stats" .SortedSetSizes}}</h3>
						{{template "freq" .SortedSetSizes}}
						<h3>2<sup><var>n</var></sup> Sizes:</h3> {{template "freq" power .SortedSetSizes}}

						<h3>Example elements:</h3> {{template "examples" .SortedSetElements}}
						<h3>Element Sizes: {{template "stats" .SortedSetElementSizes}}</h3>
						{{template "freq" .SortedSetElementSizes}}
						<h3>2<sup><var>n</var></sup> Element Sizes:</h3> {{template "freq" power .SortedSetElementSizes}}
					</div>
				</div>
			{{ end }}

			{{ if .ListKeys }}
			  <h1>Lists <small>{{summarize .ListSizes}}</small> </h1>
				<div class="panel panel-default">
					<div class="panel-body">
						<h3>Example keys:</h3> {{template "examples" .ListKeys}}
						<h3>Sizes: {{template "stats" .ListSizes}}</h3>
						{{template "freq" .ListSizes}}
						<h3>2<sup><var>n</var></sup> Sizes:</h3> {{template "freq" power .ListSizes}}

						<h3>Example elements:</h3> {{template "examples" .ListElements}}
						<h3>Element Sizes: {{template "stats" .ListElementSizes}}</h3>
						{{template "freq" .ListElementSizes}}
						<h3>2<sup><var>n</var></sup> Element Sizes:</h3> {{template "freq" power .ListElementSizes}}
					</div>
				</div>
			{{ end }}

			{{ if .HashKeys}}
			  <h1>Hashes <small>{{summarize .HashSizes}}</small> </h1>
				<div class="panel panel-default">
					<div class="panel-body">
						<h3>Example keys:</h3> {{template "examples" .HashKeys}}
						<h3>Sizes: {{template "stats" .HashSizes}}</h3>
						{{template "freq" .HashSizes}}
						<h3>2<sup><var>n</var></sup> Sizes:</h3> {{template "freq" power .HashSizes}}

						<h3>Example elements:</h3> {{template "examples" .HashElements}}
						<h3>Element Sizes: {{template "stats" .HashElementSizes}}</h3>
						{{template "freq" .HashElementSizes}}
						<h3>2<sup><var>n</var></sup> Element Sizes:</h3> {{template "freq" power .HashElementSizes}}

						<h3>Example values:</h3> {{template "examples" .HashValues}}
						<h3>Value Sizes: {{template "stats" .HashValueSizes}}</h3>
						{{template "freq" .HashValueSizes}}
						<h3>2<sup><var>n</var></sup> Value Sizes:</h3> {{template "freq" power .HashValueSizes}}
					</div>
				</div>
			{{ end }}

		 </container>

		<!-- jQuery (necessary for Bootstrap's JavaScript plugins) -->
		<script src="https://ajax.googleapis.com/ajax/libs/jquery/1.11.2/jquery.min.js"></script>
		<script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.4/js/bootstrap.min.js"></script>
	</body>
</html>

{{end}}

{{define "stats"}}
	{{ with stats . }}
		<small>(min: {{.Min}} max: {{.Max}} mean: {{fmtFloat .Mean}} std dev: {{fmtFloat .StdDev}})</small>
	{{end}}
{{end}}

{{define "examples"}}
	<ul class="list-inline">
	{{range $k, $v := .}}
		<li><code>{{$k}}</code></li>
	{{end}}
{{end}}

{{define "freq"}}
{{ $ss := summarize . }}
  <table class="table table-striped">
		<thead>
			<tr>
				<th>Size</th>
				<th># of occurrences</th>
				<th>%</th>
			</tr>
		</thead>
		<tbody>
		{{ range $s, $c := .}}
			<tr><td>{{$s}}</td> <td>{{$c}}</td> <td>{{percentage $c $ss }}</td></tr>
		{{end}}
		</tbody>
	</table>
{{end}}

`
)

// RenderHTML renders a Results instance to the supplied io.Writer
func RenderHTML(s *Results, out io.Writer) error {

	s.StringKeys = trim(s.StringKeys, MaxExampleKeys)
	s.StringValues = trim(s.StringValues, MaxExampleValues)
	s.SetKeys = trim(s.SetKeys, MaxExampleKeys)
	s.SetElements = trim(s.SetElements, MaxExampleElements)
	s.SortedSetKeys = trim(s.SortedSetKeys, MaxExampleKeys)
	s.SortedSetElements = trim(s.SortedSetElements, MaxExampleElements)
	s.HashKeys = trim(s.HashKeys, MaxExampleKeys)
	s.HashElements = trim(s.HashElements, MaxExampleElements)
	s.HashValues = trim(s.HashValues, MaxExampleValues)
	s.ListKeys = trim(s.ListKeys, MaxExampleKeys)
	s.ListElements = trim(s.ListElements, MaxExampleElements)

	fm := template.FuncMap{
		"summarize":  summarize,
		"percentage": percentage,
		"power":      ComputePowerOfTwoFreq,
		"stats":      ComputeStatistics,
		"fmtFloat":   fmtFloat,
	}
	t := template.Must(template.New("htmloutput").Funcs(fm).Parse(htmlTmpl))
	return t.ExecuteTemplate(out, "base", s)
}

// RenderText renders a Results instance to the supplied io.Writer
func RenderText(s *Results, out io.Writer) error {

	s.StringKeys = trim(s.StringKeys, MaxExampleKeys)
	s.StringValues = trim(s.StringValues, MaxExampleValues)
	s.SetKeys = trim(s.SetKeys, MaxExampleKeys)
	s.SetElements = trim(s.SetElements, MaxExampleElements)
	s.SortedSetKeys = trim(s.SortedSetKeys, MaxExampleKeys)
	s.SortedSetElements = trim(s.SortedSetElements, MaxExampleElements)
	s.HashKeys = trim(s.HashKeys, MaxExampleKeys)
	s.HashElements = trim(s.HashElements, MaxExampleElements)
	s.HashValues = trim(s.HashValues, MaxExampleValues)
	s.ListKeys = trim(s.ListKeys, MaxExampleKeys)
	s.ListElements = trim(s.ListElements, MaxExampleElements)

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
