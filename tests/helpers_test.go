/*
   Copyright Mycophonic.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package tests_test

import (
	"testing"
	"time"

	"github.com/mycophonic/agar/pkg/agar"
)

// Benchmark infrastructure.

const benchIterations = agar.DefaultBenchIterations

// benchResult wraps agar.BenchResult with a string-based Format for simpler test code.
type benchResult struct {
	Format  string
	Tool    string
	Op      string
	Median  time.Duration
	Mean    time.Duration
	Min     time.Duration
	Max     time.Duration
	Stddev  time.Duration
	PCMSize int
}

func computeResult(format, tool, op string, durations []time.Duration, pcmSize int) benchResult {
	r := agar.ComputeResult(agar.BenchFormat{Name: format}, tool, op, durations, pcmSize)

	return benchResult{
		Format:  format,
		Tool:    r.Tool,
		Op:      r.Op,
		Median:  r.Median,
		Mean:    r.Mean,
		Min:     r.Min,
		Max:     r.Max,
		Stddev:  r.Stddev,
		PCMSize: r.PCMSize,
	}
}

func printResults(t *testing.T, results []benchResult) {
	t.Helper()

	agarResults := make([]agar.BenchResult, len(results))
	for i, r := range results {
		agarResults[i] = agar.BenchResult{
			Format:  agar.BenchFormat{Name: r.Format},
			Tool:    r.Tool,
			Op:      r.Op,
			Median:  r.Median,
			Mean:    r.Mean,
			Min:     r.Min,
			Max:     r.Max,
			Stddev:  r.Stddev,
			PCMSize: r.PCMSize,
		}
	}

	agar.PrintResults(t, agar.BenchOptions{Iterations: benchIterations}, agarResults)
}
