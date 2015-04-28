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
	return fmt.Sprintf("%.2f", 100.0*float64(n)/float64(total))
}

// chartJS returns the static js what we need on the HTML templates in order to
// render charts.  The js itself has been turned into Go src using go-bindata.
// This func panics if there is any error accessing the embedded asset data.
func chartJS() string {
	data, err := Asset("Chart.min.js")
	if err != nil {
		panic(err)
	}
	return string(data)
}

type chartData struct {
	DOMElement string
	Data       map[int]int64
}

func barChart(domElement string, freq map[int]int64) chartData {
	return chartData{
		DOMElement: domElement,
		Data:       freq,
	}
}

// RenderHTML renders a plaintext report for a Results instance to the supplied
// io.Writer
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
		"barChart":   barChart,
		"chartJS":    chartJS,
	}
	t := template.Must(template.New("htmloutput").Funcs(fm).Parse(htmlTmpl))
	return t.ExecuteTemplate(out, "base", s)
}

// RenderText renders an HTML report for a Results instance to the supplied
// io.Writer
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
