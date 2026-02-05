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
	"testing"

	dsd "github.com/mycophonic/saprobe-dsd"
	intdsd "github.com/mycophonic/saprobe-dsd/internal/dsd"
)

// BenchmarkPipelineAllocs measures heap allocations in the DSD-to-PCM pipeline.
// After warmup, Process() should report 0 allocs/op.
//
// Run with: cd tests && go test -bench=BenchmarkPipelineAllocs -benchmem -count=5
func BenchmarkPipelineAllocs(b *testing.B) {
	// 1 second of DSD128 mono silence.
	const dsdRate = 5644800
	bytesPerCh := dsdRate / 8
	src := make([]byte, bytesPerCh)

	for i := range src {
		src[i] = dsdSilenceByte
	}

	pipe, err := intdsd.NewPipeline(intdsd.DSD128, intdsd.Rate352800)
	if err != nil {
		b.Fatalf("NewPipeline: %v", err)
	}

	outLen := pipe.OutputSamples(len(src))
	dst := make([]float32, outLen)

	// Warmup: one call to grow internal buffers.
	pipe.Process(src, false, dst)
	pipe.Reset()

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		pipe.Process(src, false, dst)
		pipe.Reset()
	}
}

// BenchmarkDecodeAllocs measures heap allocations in the full streaming decode path
// (container parse + pipeline + interleave). Matches the saprobe-alac benchmark pattern.
//
// Run with: BENCH_FILE=/path/to/file.dsf go test -bench=BenchmarkDecodeAllocs -benchmem -count=5
func BenchmarkDecodeAllocs(b *testing.B) {
	filePath := os.Getenv("BENCH_FILE")
	if filePath == "" {
		b.Skip("set BENCH_FILE to run this benchmark")
	}

	encoded, err := os.ReadFile(filePath)
	if err != nil {
		b.Fatalf("read: %v", err)
	}

	outBuf := make([]byte, 64*1024)

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		dec, err := dsd.NewDecoder(bytes.NewReader(encoded), dsd.Rate352800)
		if err != nil {
			b.Fatalf("new decoder: %v", err)
		}

		for {
			_, err := dec.Read(outBuf)
			if err != nil {
				break
			}
		}
	}
}

// BenchmarkResetAllocs measures heap allocations when reusing a Decoder via Reset
// for same-format files. After warmup, Reset + full decode should report 0 allocs/op
// (excluding negligible header parsing temporaries).
//
// Run with: BENCH_FILE=/path/to/file.dsf go test -bench=BenchmarkResetAllocs -benchmem -count=5
func BenchmarkResetAllocs(b *testing.B) {
	filePath := os.Getenv("BENCH_FILE")
	if filePath == "" {
		b.Skip("set BENCH_FILE to run this benchmark")
	}

	encoded, err := os.ReadFile(filePath)
	if err != nil {
		b.Fatalf("read: %v", err)
	}

	// Initial construction (allocates).
	dec, err := dsd.NewDecoder(bytes.NewReader(encoded), dsd.Rate352800)
	if err != nil {
		b.Fatalf("new decoder: %v", err)
	}

	outBuf := make([]byte, 64*1024)

	// Warmup: decode once to grow all internal buffers.
	for {
		_, readErr := dec.Read(outBuf)
		if readErr != nil {
			break
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		if err := dec.Reset(bytes.NewReader(encoded), dsd.Rate352800); err != nil {
			b.Fatalf("reset: %v", err)
		}

		for {
			_, readErr := dec.Read(outBuf)
			if readErr != nil {
				break
			}
		}
	}
}
