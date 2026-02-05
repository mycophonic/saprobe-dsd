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
	"math"
	"os"
	"testing"

	"github.com/mycophonic/agar/pkg/agar"

	intdsd "github.com/mycophonic/saprobe-dsd/internal/dsd"
)

// DSD rate / output rate combinations for the full 30-combination matrix.
//
//nolint:gochecknoglobals
var (
	dsdRates = []intdsd.InputRate{
		intdsd.DSD64,
		intdsd.DSD128,
		intdsd.DSD256,
		intdsd.DSD512,
		intdsd.DSD1024,
		intdsd.DSD2048,
	}

	outputRates = []intdsd.OutputRate{
		intdsd.Rate44100,
		intdsd.Rate88200,
		intdsd.Rate176400,
		intdsd.Rate352800,
		intdsd.Rate705600,
	}
)

// TestDSDPipelineSampleCount verifies that every rate combination produces
// the correct number of output samples and no corrupted values.
func TestDSDPipelineSampleCount(t *testing.T) {
	t.Parallel()

	for _, dsdRate := range dsdRates {
		for _, outRate := range outputRates {
			t.Run(dsdRate.String()+"/"+outRate.String(), func(t *testing.T) {
				t.Parallel()

				pipe, err := intdsd.NewPipeline(dsdRate, outRate)
				if err != nil {
					t.Fatalf("NewPipeline(%v, %v): %v", dsdRate, outRate, err)
				}

				dsdPath := agar.DSDSine(t.TempDir(), t, int(dsdRate), 1000)

				src, err := os.ReadFile(dsdPath)
				if err != nil {
					t.Fatalf("reading DSD data: %v", err)
				}

				expectedOut := pipe.OutputSamples(len(src))
				dst := make([]float32, expectedOut)

				got := pipe.Process(src, false, dst)
				if got != expectedOut {
					t.Errorf("sample count: got %d, want %d", got, expectedOut)
				}

				if hasNaNOrInf(dst[:got]) {
					t.Error("output contains NaN or Inf")
				}

				peak := maxAbs(dst[:got])
				if peak > 10.0 {
					t.Errorf("output peak %.4f exceeds reasonable range", peak)
				}
			})
		}
	}
}

// TestDSDPipelineSilence verifies that silence input produces near-zero output.
func TestDSDPipelineSilence(t *testing.T) {
	t.Parallel()

	for _, dsdRate := range dsdRates {
		for _, outRate := range outputRates {
			t.Run(dsdRate.String()+"/"+outRate.String(), func(t *testing.T) {
				t.Parallel()

				pipe, err := intdsd.NewPipeline(dsdRate, outRate)
				if err != nil {
					t.Fatalf("NewPipeline(%v, %v): %v", dsdRate, outRate, err)
				}

				dsdPath := agar.DSDSilence(t.TempDir(), t, int(dsdRate))

				src, err := os.ReadFile(dsdPath)
				if err != nil {
					t.Fatalf("reading DSD data: %v", err)
				}

				dst := make([]float32, pipe.OutputSamples(len(src)))

				got := pipe.Process(src, false, dst)

				peak := maxAbs(dst[:got])
				// Sigma-delta silence still has quantization noise; threshold is generous.
				const silenceThreshold = 0.15
				if peak > silenceThreshold {
					t.Errorf("silence peak %.4f exceeds threshold %.2f", peak, silenceThreshold)
				}
			})
		}
	}
}

// TestDSDPipelineDC verifies that a positive DC input produces positive output.
func TestDSDPipelineDC(t *testing.T) {
	t.Parallel()

	// Representative subset of rate combinations.
	dcCases := []struct {
		dsdRate intdsd.InputRate
		outRate intdsd.OutputRate
	}{
		{intdsd.DSD64, intdsd.Rate44100},
		{intdsd.DSD128, intdsd.Rate88200},
		{intdsd.DSD256, intdsd.Rate176400},
	}

	for _, tc := range dcCases {
		t.Run(tc.dsdRate.String()+"/"+tc.outRate.String(), func(t *testing.T) {
			t.Parallel()

			pipe, err := intdsd.NewPipeline(tc.dsdRate, tc.outRate)
			if err != nil {
				t.Fatalf("NewPipeline: %v", err)
			}

			dsdPath := agar.DSDDC(t.TempDir(), t, int(tc.dsdRate), 0.5)

			src, err := os.ReadFile(dsdPath)
			if err != nil {
				t.Fatalf("reading DSD data: %v", err)
			}

			dst := make([]float32, pipe.OutputSamples(len(src)))

			got := pipe.Process(src, false, dst)

			avg := meanF32(dst[:got])
			if avg <= 0 {
				t.Errorf("DC positive input produced non-positive mean: %.6f", avg)
			}
		})
	}
}

// TestDSDPipelineSineEnergy verifies that a 1 kHz sine produces non-trivial output energy.
func TestDSDPipelineSineEnergy(t *testing.T) {
	t.Parallel()

	// Representative subset.
	sineCases := []struct {
		dsdRate intdsd.InputRate
		outRate intdsd.OutputRate
	}{
		{intdsd.DSD64, intdsd.Rate44100},
		{intdsd.DSD64, intdsd.Rate352800},
		{intdsd.DSD128, intdsd.Rate88200},
		{intdsd.DSD256, intdsd.Rate44100},
		{intdsd.DSD1024, intdsd.Rate44100},
		{intdsd.DSD2048, intdsd.Rate44100},
	}

	for _, tc := range sineCases {
		t.Run(tc.dsdRate.String()+"/"+tc.outRate.String(), func(t *testing.T) {
			t.Parallel()

			pipe, err := intdsd.NewPipeline(tc.dsdRate, tc.outRate)
			if err != nil {
				t.Fatalf("NewPipeline: %v", err)
			}

			dsdPath := agar.DSDSine(t.TempDir(), t, int(tc.dsdRate), 1000)

			src, err := os.ReadFile(dsdPath)
			if err != nil {
				t.Fatalf("reading DSD data: %v", err)
			}

			dst := make([]float32, pipe.OutputSamples(len(src)))

			got := pipe.Process(src, false, dst)

			energy := rmsF32(dst[:got])
			// A decoded 1 kHz sine should have measurable energy.
			const minRMS = 0.001
			if energy < minRMS {
				t.Errorf("sine RMS %.6f below minimum %.4f", energy, minRMS)
			}
		})
	}
}

// TestDSDPipeline8to1Equivalence verifies that Pipeline produces identical
// output to the direct Converter for the 8:1 DSD64->352.8k path (zero halfband stages).
func TestDSDPipeline8to1Equivalence(t *testing.T) {
	t.Parallel()

	pipe, err := intdsd.NewPipeline(intdsd.DSD64, intdsd.Rate352800)
	if err != nil {
		t.Fatalf("NewPipeline: %v", err)
	}

	dsdPath := agar.DSDSine(t.TempDir(), t, int(intdsd.DSD64), 1000)

	src, err := os.ReadFile(dsdPath)
	if err != nil {
		t.Fatalf("reading DSD data: %v", err)
	}

	// Pipeline path (outputs float32).
	pipeDst := make([]float32, pipe.OutputSamples(len(src)))
	pipeGot := pipe.Process(src, false, pipeDst)

	// Direct Converter path (outputs float64, convert to float32 for comparison).
	conv := intdsd.NewConverter()
	convDst64 := make([]float64, len(src))
	conv.Translate(len(src), src, 1, false, convDst64, 1)

	if pipeGot != len(src) {
		t.Fatalf("pipeline produced %d samples, expected %d", pipeGot, len(src))
	}

	for idx := range pipeGot {
		convF32 := float32(convDst64[idx])
		if pipeDst[idx] != convF32 {
			t.Errorf("sample %d: pipeline=%.8f, converter=%.8f", idx, pipeDst[idx], convF32)

			break
		}
	}
}

// Test helpers.

func hasNaNOrInf(samples []float32) bool {
	for _, sample := range samples {
		if math.IsNaN(float64(sample)) || math.IsInf(float64(sample), 0) {
			return true
		}
	}

	return false
}

func maxAbs(samples []float32) float32 {
	var peak float32

	for _, sample := range samples {
		absVal := float32(math.Abs(float64(sample)))
		if absVal > peak {
			peak = absVal
		}
	}

	return peak
}

func rmsF32(samples []float32) float32 {
	if len(samples) == 0 {
		return 0
	}

	var sumSq float64

	for _, sample := range samples {
		sumSq += float64(sample) * float64(sample)
	}

	return float32(math.Sqrt(sumSq / float64(len(samples))))
}

func meanF32(samples []float32) float32 {
	if len(samples) == 0 {
		return 0
	}

	var sum float64

	for _, sample := range samples {
		sum += float64(sample)
	}

	return float32(sum / float64(len(samples)))
}
