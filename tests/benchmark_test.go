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
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mycophonic/agar/pkg/agar"

	dsd "github.com/mycophonic/saprobe-dsd"
)

type benchFormat struct {
	Name    string
	DSDRate int
}

//nolint:gochecknoglobals
var benchFormats = []benchFormat{
	{"DSD64", 2822400},
	{"DSD128", 5644800},
	{"DSD256", 11289600},
}

const (
	// dsdSilenceByte is the DSD idle pattern (alternating 1/0 bits, MSB-first).
	dsdSilenceByte = 0x69

	// syntheticDuration is the duration of synthetic DSD data for benchmarks (seconds).
	// Kept short because the streaming decoder does significant work per byte
	// (container parsing + pipeline + interleaving) and must complete under race detector.
	syntheticDuration = 1
)

// TestBenchmarkDecode benchmarks the DSD->PCM streaming decoder with synthetic data
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

		dsfData := generateDSFContainer(bf.DSDRate, 2)

		t.Logf("  DSF container: %.1f MB", float64(len(dsfData))/(1024*1024))

		results = append(results, benchDecoderSaprobe(t, bf.Name, dsfData))
	}

	printResults(t, results)
}

func benchDecoderSaprobe(t *testing.T, name string, dsfData []byte) benchResult {
	t.Helper()

	durations := make([]time.Duration, benchIterations)
	outBuf := make([]byte, 64*1024)

	var pcmSize int

	for iter := range benchIterations {
		start := time.Now()

		dec, err := dsd.NewDecoder(bytes.NewReader(dsfData), dsd.Rate352800)
		if err != nil {
			t.Fatalf("saprobe decoder: %v", err)
		}

		iterSize := 0

		for {
			n, readErr := dec.Read(outBuf)
			iterSize += n

			if readErr != nil {
				break
			}
		}

		durations[iter] = time.Since(start)
		pcmSize = iterSize
	}

	return computeResult(name, "saprobe", "decode", durations, pcmSize)
}

// TestBenchmarkDecodeFile benchmarks DSD-to-PCM decoding using a real DSF or DFF file.
// Compares saprobe (streaming decoder) vs ffmpeg.
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

	// Read entire file for repeated decoding.
	encoded, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	name := filepath.Base(filePath)

	var results []benchResult

	// Benchmark saprobe: streaming decoder.
	results = append(results, benchFileSaprobe(t, name, encoded))

	// Benchmark ffmpeg DSD->PCM.
	results = append(results, benchFileFFmpeg(t, name, filePath))

	printResults(t, results)
}

func benchFileSaprobe(t *testing.T, name string, encoded []byte) benchResult {
	t.Helper()

	durations := make([]time.Duration, benchIterations)
	outBuf := make([]byte, 64*1024)

	for iter := range benchIterations {
		start := time.Now()

		dec, err := dsd.NewDecoder(bytes.NewReader(encoded), dsd.Rate352800)
		if err != nil {
			t.Fatalf("saprobe decoder iter %d: %v", iter, err)
		}

		for {
			_, readErr := dec.Read(outBuf)
			if readErr != nil {
				break
			}
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

// generateDSFContainer creates a minimal DSF container with synthetic DSD silence data
// using the default syntheticDuration.
func generateDSFContainer(dsdRate, channels int) []byte {
	return generateDSFContainerN(dsdRate, channels, syntheticDuration)
}

// generateDSFContainerN creates a minimal DSF container with synthetic DSD silence data
// for the specified duration in seconds.
//

func generateDSFContainerN(dsdRate, channels, durationSec int) []byte {
	bytesPerCh := dsdRate / 8 * durationSec
	blockSize := 4096

	// Round up to block boundary.
	blocksPerCh := (bytesPerCh + blockSize - 1) / blockSize
	paddedBytesPerCh := blocksPerCh * blockSize
	dataPayload := paddedBytesPerCh * channels

	// DSF structure: DSD chunk (28) + fmt chunk (52) + data chunk (12 + payload).
	dsdChunkSize := 28
	fmtChunkSize := 52
	dataChunkHeader := 12
	totalSize := dsdChunkSize + fmtChunkSize + dataChunkHeader + dataPayload

	buf := make([]byte, totalSize)
	off := 0

	// DSD chunk.
	copy(buf[off:], "DSD ")
	putLE64(buf[off+4:], uint64(dsdChunkSize))
	putLE64(buf[off+12:], uint64(totalSize))
	putLE64(buf[off+20:], 0) // no metadata
	off += dsdChunkSize

	// fmt chunk.
	copy(buf[off:], "fmt ")
	putLE64(buf[off+4:], uint64(fmtChunkSize))
	putLE32(buf[off+12:], 1)                           // format version
	putLE32(buf[off+16:], 0)                           // format ID (DSD raw)
	putLE32(buf[off+20:], uint32(channels))            // channel type (stereo=2)
	putLE32(buf[off+24:], uint32(channels))            // channel count
	putLE32(buf[off+28:], uint32(dsdRate))             // sample rate
	putLE32(buf[off+32:], 1)                           // bits per sample
	putLE64(buf[off+36:], uint64(int64(bytesPerCh)*8)) // sample count per channel
	putLE32(buf[off+44:], uint32(blockSize))           // block size per channel
	putLE32(buf[off+48:], 0)                           // reserved
	off += fmtChunkSize

	// data chunk.
	copy(buf[off:], "data")
	putLE64(buf[off+4:], uint64(dataChunkHeader+dataPayload))
	off += dataChunkHeader

	// Fill with block-interleaved DSD silence.
	// DSF format: [ch0_block][ch1_block][ch0_block][ch1_block]...
	for block := range blocksPerCh {
		for ch := range channels {
			start := off + (block*channels+ch)*blockSize

			// Don't write past actual data.
			actualBytes := min(blockSize, bytesPerCh-block*blockSize)

			for i := range actualBytes {
				buf[start+i] = dsdSilenceByte
			}
		}
	}

	return buf
}

func putLE32(b []byte, v uint32) {
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
}

func putLE64(b []byte, v uint64) {
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
	b[4] = byte(v >> 32)
	b[5] = byte(v >> 40)
	b[6] = byte(v >> 48)
	b[7] = byte(v >> 56)
}
