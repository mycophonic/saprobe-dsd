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

package dsd

import (
	"fmt"

	intdsd "github.com/mycophonic/saprobe-dsd/internal/dsd"
)

// Pipeline orchestrates multi-stage DSD-to-PCM conversion.
type Pipeline = intdsd.Pipeline

// Converter performs DSD-to-PCM conversion for a single channel using
// a symmetric 96-tap FIR lowpass filter with 8:1 decimation.
type Converter = intdsd.Converter

// NoiseShaper applies noise shaping via second-order sections (SOS) IIR filtering.
type NoiseShaper = intdsd.NoiseShaper

// NewPipeline creates a conversion pipeline for the given rate combination.
//
// All 25 combinations of DSD64..DSD1024 and 44.1k..705.6k are supported.
func NewPipeline(dsdRate InputRate, outputRate OutputRate) (*Pipeline, error) {
	pipe, err := intdsd.NewPipeline(dsdRate, outputRate)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnsupportedRate, err)
	}

	return pipe, nil
}

// NewConverter creates a new DSD-to-PCM converter initialized to silence.
func NewConverter() *Converter {
	return intdsd.NewConverter()
}

// NewNoiseShaper creates a noise shaper with the given SOS coefficients.
func NewNoiseShaper(coeffs []float32) *NoiseShaper {
	return intdsd.NewNoiseShaper(coeffs)
}

// DefaultNoiseShapeCoeffs contains noise shaping coefficients for 16-bit
// quantization of DSD-to-PCM output.
//
//nolint:gochecknoglobals
var DefaultNoiseShapeCoeffs = intdsd.DefaultNoiseShapeCoeffs
