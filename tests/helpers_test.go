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
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mycophonic/agar/pkg/agar"

	dsd "github.com/mycophonic/saprobe-dsd"
)

// decodeDSDFile opens and decodes a DSF or DFF file, returning per-channel DSD bytes.
func decodeDSDFile(path string) ([][]byte, dsd.DSDFormat, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, dsd.DSDFormat{}, err
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".dsf":
		return dsd.DecodeDSF(file)
	default:
		return dsd.DecodeDFF(file)
	}
}

// discoverDSDFiles returns all .dsf and .dff files in the given directory, sorted by name.
func discoverDSDFiles(t *testing.T, dir string) []string {
	t.Helper()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading %s: %v", dir, err)
	}

	var files []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext == ".dsf" || ext == ".dff" {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}

	if len(files) == 0 {
		t.Fatalf("no .dsf or .dff files found in %s", dir)
	}

	return files
}

// Benchmark infrastructure.

const (
	benchIterations = agar.DefaultBenchIterations
	benchDuration   = 10 // seconds of synthetic DSD audio
)

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
