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
	htmlTmpl = `
{{define "base"}}

<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>reckoning</title>
    <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.4/css/bootstrap.min.css">

    <style>
      canvas {
        width: 75%;
        height: auto;
			  margin-left: auto;
			  margin-right: auto;
			  display: block;
      }
    </style>

		<script type="text/javascript">{{chartJS}}</script>
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
						{{template "barchart" barChart "StringSizes" .StringSizes}}
						<h3>2<sup><var>n</var></sup> Value Sizes:</h3>
						{{template "freq" power .StringSizes}}
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
						{{template "barchart" barChart "SetSizes" .SetSizes}}
						<h3>2<sup><var>n</var></sup> Sizes:</h3>
						{{template "freq" power .SetSizes}}

						<h3>Example elements:</h3> {{template "examples" .SetElements}}
						<h3>Element Sizes: {{template "stats" .SetElementSizes}}</h3>
						{{template "freq" .SetElementSizes}}
						{{template "barchart" barChart "SetElementSizes" .SetElementSizes}}
						<h3>2<sup><var>n</var></sup> Element Sizes:</h3>
						{{template "freq" power .SetElementSizes}}
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
						{{template "barchart" barChart "SortedSetSizes" .SortedSetSizes}}
						<h3>2<sup><var>n</var></sup> Sizes:</h3>
						{{template "freq" power .SortedSetSizes}}

						<h3>Example elements:</h3> {{template "examples" .SortedSetElements}}
						<h3>Element Sizes: {{template "stats" .SortedSetElementSizes}}</h3>
						{{template "freq" .SortedSetElementSizes}}
						{{template "barchart" barChart "SortedSetElementSizes" .SortedSetElementSizes}}
						<h3>2<sup><var>n</var></sup> Element Sizes:</h3>
						{{template "freq" power .SortedSetElementSizes}}
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
						{{template "barchart" barChart "ListSizes" .ListSizes}}
						<h3>2<sup><var>n</var></sup> Sizes:</h3>
						{{template "freq" power .ListSizes}}

						<h3>Example elements:</h3> {{template "examples" .ListElements}}
						<h3>Element Sizes: {{template "stats" .ListElementSizes}}</h3>
						{{template "freq" .ListElementSizes}}
						{{template "barchart" barChart "ListElementSizes" .ListElementSizes}}
						<h3>2<sup><var>n</var></sup> Element Sizes:</h3>
						{{template "freq" power .ListElementSizes}}
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
						{{template "barchart" barChart "HashSizes" .HashSizes}}
						<h3>2<sup><var>n</var></sup> Sizes:</h3>
						{{template "freq" power .HashSizes}}

						<h3>Example elements:</h3> {{template "examples" .HashElements}}
						<h3>Element Sizes: {{template "stats" .HashElementSizes}}</h3>
						{{template "freq" .HashElementSizes}}
						{{template "barchart" barChart "HashElementSizes" .HashElementSizes}}
						<h3>2<sup><var>n</var></sup> Element Sizes:</h3>
						{{template "freq" power .HashElementSizes}}

						<h3>Example values:</h3> {{template "examples" .HashValues}}
						<h3>Value Sizes: {{template "stats" .HashValueSizes}}</h3>
						{{template "freq" .HashValueSizes}}
						{{template "barchart" barChart "HashValueSizes" .HashValueSizes}}
						<h3>2<sup><var>n</var></sup> Value Sizes:</h3>
						{{template "freq" power .HashValueSizes}}
					</div>
				</div>
			{{ end }}

		 </container>

		<script src="https://ajax.googleapis.com/ajax/libs/jquery/1.11.2/jquery.min.js"></script>
		<script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.4/js/bootstrap.min.js"></script>
	</body>
</html>

{{end}}

{{define "barchart"}}
	{{/* Supplies the necessary DOM elements and JS to render a simple bar chart (if there are >= 4 data points) */}}

  {{ $l := len .Data }}
  {{ if ge $l 4}}
	{{ $total := summarize .Data }}
	<button class="btn btn-primary" type="button" data-toggle="collapse" data-target="#{{.DOMElement}}Collapse">toggle chart</button>
	<div class="collapse in" id="{{.DOMElement}}Collapse">
		<canvas id="{{.DOMElement}}"></canvas>
  </div>
	<script type="text/javascript">
    // Chart.defaults.global.responsive = true;
		var ctx = document.getElementById("{{.DOMElement}}").getContext("2d");

		var data = {
			labels: [ {{range $k, $v := .Data}} "{{$k}}", {{end}} ],
			datasets: [
			{
				label: "size frequencies",
				fillColor: "rgba(151,187,205,0.5)",
				strokeColor: "rgba(151,187,205,0.8)",
				highlightFill: "rgba(151,187,205,0.75)",
				highlightStroke: "rgba(151,187,205,1)",
				data: [ {{range $k, $v := .Data}} {{percentage $v $total}}, {{end}} ]
			}
			]
		};
		new Chart(ctx).Bar(data, {"scaleLabel": "<%=value%>%"});
	</script>
	{{end}}
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
			<tr><td>{{$s}}</td> <td>{{$c}}</td> <td>{{percentage $c $ss}}%</td></tr>
		{{end}}
		</tbody>
	</table>
{{end}}

`
)
