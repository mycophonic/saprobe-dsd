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
	"errors"
	"fmt"
	"math"
)

// Pipeline orchestrates multi-stage DSD-to-PCM conversion.
//
// Stage 1: Byte-level 8:1 decimation (Gesemann 96-tap filter) for most rate
// combinations, or 4:1 bit-level FIR for the DSD64->705.6k edge case.
//
// Stage 2: Zero or more cascaded halfband 2:1 decimation stages on float data.
type Pipeline struct {
	dsdRate    InputRate
	outputRate OutputRate

	// Stage 1 (exactly one is non-nil).
	stage1   *Converter     // 8:1 byte-level
	stage1x4 *Converter4to1 // 4:1 bit-level (DSD64->705.6k only)

	// Stage 2: cascaded halfband 2:1 decimators.
	halfbands []*Halfband

	// Intermediate buffers, reused across Process calls (grow-only, zero-alloc after warmup).
	stage1Buf []float32 // stage1 output (Converter writes float32)
	buf64a    []float64 // stage1 output converted to float64, or halfband input
	buf64b    []float64 // halfband output (ping-pong with buf64a)
}

// ErrUnsupportedRate is returned when the DSD/output rate combination is invalid.
var ErrUnsupportedRate = errors.New("dsd: unsupported rate combination")

// NewPipeline creates a conversion pipeline for the given rate combination.
//
// All 30 combinations of DSD64..DSD2048 and 44.1k..705.6k are supported.
// Returns ErrUnsupportedRate for invalid combinations.
func NewPipeline(dsdRate InputRate, outputRate OutputRate) (*Pipeline, error) {
	initTables()

	ratio := int(dsdRate) / int(outputRate)
	if int(dsdRate) != ratio*int(outputRate) {
		return nil, fmt.Errorf("%w: %v / %v = non-integer ratio", ErrUnsupportedRate, dsdRate, outputRate)
	}

	// Verify ratio is a power of 2 and >= 4.
	if ratio < 4 || ratio&(ratio-1) != 0 {
		return nil, fmt.Errorf(
			"%w: %v / %v = %d (must be power-of-2 >= 4)",
			ErrUnsupportedRate,
			dsdRate,
			outputRate,
			ratio,
		)
	}

	pipe := &Pipeline{
		dsdRate:    dsdRate,
		outputRate: outputRate,
	}

	if ratio == 4 {
		// DSD64 -> 705.6 kHz: use 4:1 bit-level converter, no halfband stages.
		pipe.stage1x4 = newConverter4to1()
	} else {
		// All other combinations: 8:1 byte-level + cascaded halfband 2:1.
		pipe.stage1 = NewConverter()

		// After 8:1, need log2(ratio/8) halfband stages.
		halfbandCount := int(math.Log2(float64(ratio / 8)))
		pipe.halfbands = make([]*Halfband, halfbandCount)

		for i := range halfbandCount {
			pipe.halfbands[i] = newHalfband()
		}
	}

	return pipe, nil
}

// DSDRate returns the pipeline's DSD input rate.
func (p *Pipeline) DSDRate() InputRate { return p.dsdRate }

// PCMRate returns the pipeline's PCM output rate.
func (p *Pipeline) PCMRate() OutputRate { return p.outputRate }

// Reset reinitializes all converter states to silence.
func (p *Pipeline) Reset() {
	if p.stage1 != nil {
		p.stage1.Reset()
	}

	if p.stage1x4 != nil {
		p.stage1x4.Reset()
	}

	for _, hb := range p.halfbands {
		hb.Reset()
	}
}

// Clone returns a deep copy of the pipeline with identical state.
func (p *Pipeline) Clone() *Pipeline {
	clone := &Pipeline{
		dsdRate:    p.dsdRate,
		outputRate: p.outputRate,
	}

	if p.stage1 != nil {
		clone.stage1 = p.stage1.Clone()
	}

	if p.stage1x4 != nil {
		clone.stage1x4 = p.stage1x4.Clone()
	}

	clone.halfbands = make([]*Halfband, len(p.halfbands))
	for i, hb := range p.halfbands {
		clone.halfbands[i] = hb.Clone()
	}

	return clone
}

// Process converts DSD input bytes to PCM float32 output.
//
// Each input byte contains 8 DSD bits. For the 8:1 path, each byte produces
// one intermediate sample that is then decimated by the halfband stages.
// For the 4:1 path, each byte produces 2 output samples.
//
// Parameters:
//   - src: input DSD data for a single channel
//   - lsbFirst: true if least-significant bit is first in each octet
//   - dst: output buffer (must have capacity for the output samples)
//
// Returns the number of output samples written to dst.
func (p *Pipeline) Process(src []byte, lsbFirst bool, dst []float32) int {
	nBytes := len(src)

	if p.stage1x4 != nil {
		// 4:1 path: direct to output, no halfband stages.
		return p.stage1x4.Translate(nBytes, src, 1, lsbFirst, dst, 1)
	}

	// 8:1 path: byte-level decimation produces nBytes float32 samples.
	// Convert to float64 for halfband processing.
	p.stage1Buf = ensureLenF32(p.stage1Buf, nBytes)
	stage1Out := p.stage1Buf
	p.stage1.Translate(nBytes, src, 1, lsbFirst, stage1Out, 1)

	// Convert stage1 output to float64 for halfband chain.
	p.buf64a = ensureLen(p.buf64a, nBytes)
	for i, v := range stage1Out {
		p.buf64a[i] = float64(v)
	}

	current := p.buf64a[:nBytes]

	// Cascade through halfband stages.
	for _, hb := range p.halfbands {
		p.buf64b = hb.Process(current, p.buf64b)
		current, p.buf64b = p.buf64b, current
	}

	// Convert final float64 output to float32.
	for i, v := range current {
		dst[i] = float32(v)
	}

	return len(current)
}

// OutputSamples returns the number of output samples produced from nBytes
// input bytes for a single channel.
func (p *Pipeline) OutputSamples(nBytes int) int {
	if p.stage1x4 != nil {
		return nBytes * 2
	}

	// 8:1 produces nBytes samples, each halfband halves.
	out := nBytes
	for range p.halfbands {
		out /= 2
	}

	return out
}

func ensureLen(buf []float64, n int) []float64 {
	if cap(buf) >= n {
		return buf[:n]
	}

	return make([]float64, n)
}

func ensureLenF32(buf []float32, n int) []float32 {
	if cap(buf) >= n {
		return buf[:n]
	}

	return make([]float32, n)
}
