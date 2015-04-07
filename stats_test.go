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
package sampler

import (
	"math"
	"testing"
)

func assertInt(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("expected: %d, actual: %d", expected, actual)
	}
}

func assertFloat(t *testing.T, expected, actual, epsilon float64) {
	if math.Abs(expected-actual) >= epsilon {
		t.Errorf("expected: %.6f, actual: %.6f", expected, actual)
	}
}

func assertNaN(t *testing.T, actual float64) {
	if !math.IsNaN(actual) {
		t.Errorf("expected NaN, actual: %.6f", actual)
	}
}

const (
	epsilon float64 = 0.00001
)

func TestStatistics(t *testing.T) {

	m := make(map[int]int64)
	m[-1] = 1
	m[13] = 1
	m[67] = 1
	m[999] = 1
	m[342] = 1

	stats := ComputeStatistics(m)

	assertInt(t, 999, stats.Max)
	assertInt(t, -1, stats.Min)
	assertFloat(t, 284.0, stats.Mean, epsilon)
	assertFloat(t, 423.18554, stats.StdDev, epsilon)

	m = make(map[int]int64)
	m[45] = 4
	m[123] = 8
	m[99999] = 2
	m[77] = 1

	stats = ComputeStatistics(m)

	assertInt(t, 99999, stats.Max)
	assertInt(t, 45, stats.Min)
	assertFloat(t, 13415.93333, stats.Mean, epsilon)
	assertFloat(t, 35152.65287, stats.StdDev, epsilon)
}

func TestStatisticsZeroValues(t *testing.T) {

	m := make(map[int]int64)
	stats := ComputeStatistics(m)

	assertInt(t, 0, stats.Max)
	assertInt(t, 0, stats.Min)
	assertNaN(t, stats.Mean)
	assertNaN(t, stats.StdDev)
}
