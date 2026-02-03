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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mycophonic/agar/pkg/agar"

	dsd "github.com/mycophonic/saprobe-dsd"
)

type benchFormat struct {
	Name      string
	DSDRate   int
	InputRate dsd.InputRate
}

//nolint:gochecknoglobals
var benchFormats = []benchFormat{
	{"DSD64", 2822400, dsd.DSD64},
	{"DSD128", 5644800, dsd.DSD128},
	{"DSD256", 11289600, dsd.DSD256},
}

// dsdSilenceByte is the DSD idle pattern (alternating 1/0 bits, MSB-first).
const dsdSilenceByte = 0x69

// generateDSD creates benchDuration seconds of synthetic stereo DSD data.
// Uses DSD silence (0x69) for reproducible, signal-independent timing.
func generateDSD(dsdRate, channels int) [][]byte {
	bytesPerCh := dsdRate / 8 * benchDuration
	data := make([][]byte, channels)

	for chIdx := range channels {
		data[chIdx] = make([]byte, bytesPerCh)

		for i := range data[chIdx] {
			data[chIdx][i] = dsdSilenceByte
		}
	}

	return data
}

// TestBenchmarkDecode benchmarks the DSD→PCM pipeline with synthetic data
// across DSD64, DSD128, and DSD256 rates.
//
//nolint:paralleltest // Benchmark must run sequentially for accurate timing.
func TestBenchmarkDecode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping benchmark in short mode")
	}

	var results []benchResult

	for _, bf := range benchFormats {
		t.Logf("=== %s ===", bf.Name)

		channels := generateDSD(bf.DSDRate, 2)
		bytesPerCh := len(channels[0])

		t.Logf("  DSD data: %.1f MB (%d bytes/ch × 2 ch)",
			float64(bytesPerCh*2)/(1024*1024), bytesPerCh)

		results = append(results, benchPipelineSaprobe(t, bf, channels))
	}

	printResults(t, results)
}

func benchPipelineSaprobe(t *testing.T, bf benchFormat, channels [][]byte) benchResult {
	t.Helper()

	durations := make([]time.Duration, benchIterations)

	for iter := range benchIterations {
		start := time.Now()

		pipe, err := dsd.NewPipeline(bf.InputRate, dsd.Rate352800)
		if err != nil {
			t.Fatalf("saprobe pipeline: %v", err)
		}

		for _, chData := range channels {
			outLen := pipe.OutputSamples(len(chData))
			pcmOut := make([]float32, outLen)

			pipe.Process(chData, false, pcmOut)
			pipe.Reset()
		}

		durations[iter] = time.Since(start)
	}

	// Compute output PCM size for throughput reporting.
	pipe, _ := dsd.NewPipeline(bf.InputRate, dsd.Rate352800)

	var pcmSize int

	for _, chData := range channels {
		pcmSize += pipe.OutputSamples(len(chData)) * 4 // float32 = 4 bytes
	}

	return computeResult(bf.Name, "saprobe", "decode", durations, pcmSize)
}

// TestBenchmarkDecodeFile benchmarks DSD-to-PCM decoding using a real DSF or DFF file.
// Compares saprobe (container decode + pipeline) vs ffmpeg.
//
// Set BENCH_FILE to a .dsf or .dff file path to run.
//
//nolint:paralleltest // Benchmark must run sequentially for accurate timing.
func TestBenchmarkDecodeFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping benchmark in short mode")
	}

	filePath := os.Getenv("BENCH_FILE")
	if filePath == "" {
		t.Skip("set BENCH_FILE to run this benchmark")
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".dsf" && ext != ".dff" {
		t.Fatalf("unsupported extension %q: expected .dsf or .dff", ext)
	}

	t.Logf("File: %s", filePath)

	// Pre-decode once to get format info and validate.
	channels, dsdFmt, err := decodeDSDFile(filePath)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	inputRate := dsd.InputRate(dsdFmt.SampleRate)
	lsbFirst := dsdFmt.BitOrder == dsd.LSBFirst
	numChannels := len(channels)
	bytesPerCh := len(channels[0])

	name := fmt.Sprintf("%s (%s)", filepath.Base(filePath), inputRate.String())

	t.Logf("  Format: %s, %d ch, %d bytes/ch (%.1f MB DSD data)",
		inputRate.String(), numChannels, bytesPerCh,
		float64(bytesPerCh*numChannels)/(1024*1024))

	var results []benchResult

	// Benchmark saprobe: container decode + DSD→PCM pipeline.
	results = append(results, benchFileSaprobe(t, name, filePath, inputRate, lsbFirst))

	// Benchmark ffmpeg DSD→PCM.
	results = append(results, benchFileFFmpeg(t, name, filePath))

	printResults(t, results)
}

func benchFileSaprobe(
	t *testing.T,
	name, filePath string,
	inputRate dsd.InputRate,
	lsbFirst bool,
) benchResult {
	t.Helper()

	durations := make([]time.Duration, benchIterations)

	for iter := range benchIterations {
		start := time.Now()

		// Container decode.
		channels, _, err := decodeDSDFile(filePath)
		if err != nil {
			t.Fatalf("saprobe decode iter %d: %v", iter, err)
		}

		// DSD→PCM pipeline.
		pipe, err := dsd.NewPipeline(inputRate, dsd.Rate352800)
		if err != nil {
			t.Fatalf("saprobe pipeline: %v", err)
		}

		for _, chData := range channels {
			outLen := pipe.OutputSamples(len(chData))
			pcmOut := make([]float32, outLen)

			pipe.Process(chData, lsbFirst, pcmOut)
			pipe.Reset()
		}

		durations[iter] = time.Since(start)
	}

	return computeResult(name, "saprobe", "decode", durations, 0)
}

func benchFileFFmpeg(t *testing.T, name, filePath string) benchResult {
	t.Helper()

	durations := make([]time.Duration, benchIterations)

	for iter := range benchIterations {
		start := time.Now()

		agar.FFmpeg(t, agar.FFmpegOptions{
			Args: []string{
				"-i", filePath,
				"-ar", "352800",
				"-ac", "2",
				"-f", "f32le",
				"-acodec", "pcm_f32le",
				"-",
			},
		})

		durations[iter] = time.Since(start)
	}

	return computeResult(name, "ffmpeg", "decode", durations, 0)
}
