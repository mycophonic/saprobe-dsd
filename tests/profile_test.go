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

	"github.com/mycophonic/agar/pkg/agar"
)

// profileDuration is longer than syntheticDuration to give pprof enough samples.
const profileDuration = 10

// TestProfileDecode runs saprobe-only decoding for clean pprof profiling.
// Use with: hack/bench.sh TestProfileDecode
//
//nolint:paralleltest // Profile must run sequentially for accurate sampling.
func TestProfileDecode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping profile in short mode")
	}

	var results []agar.BenchResult

	for _, bf := range benchFormats {
		t.Logf("=== %s ===", bf.Name)

		dsfData := generateDSFContainerN(bf.DSDRate, 2, profileDuration)

		t.Logf("  DSF container: %.1f MB", float64(len(dsfData))/(1024*1024))

		results = append(results, benchDecoderSaprobe(t, bf.BenchFormat, dsfData))
	}

	agar.PrintResults(t, benchOpts, results)
}

// TestProfileDecodeFile runs saprobe-only decoding of a real file for clean pprof profiling.
// Set BENCH_FILE to a .dsf or .dff file path.
// Use with: hack/bench.sh TestProfileDecodeFile /path/to/file.dsf
//
//nolint:paralleltest // Profile must run sequentially for accurate sampling.
func TestProfileDecodeFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping profile in short mode")
	}

	filePath := os.Getenv("BENCH_FILE")
	if filePath == "" {
		t.Skip("set BENCH_FILE to run this profile")
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".dsf" && ext != ".dff" {
		t.Fatalf("unsupported extension %q: expected .dsf or .dff", ext)
	}

	encoded, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	t.Logf("File: %s (%.1f MB)", filePath, float64(len(encoded))/(1024*1024))

	bf := agar.BenchFormat{Name: filepath.Base(filePath)}

	results := []agar.BenchResult{benchFileSaprobe(t, bf, encoded)}

	agar.PrintResults(t, benchOpts, results)
}
