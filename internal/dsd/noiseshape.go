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

// NoiseShaper applies noise shaping via second-order sections (SOS) IIR filtering.
// It is used to preserve dynamic range when quantizing float PCM to 16-bit integer.
//
// Usage pattern for each sample:
//
//	correction := ns.Get()
//	shaped := sample + correction
//	quantized := round(shaped)
//	ns.Update(quantized - shaped)  // feed back quantization error
type NoiseShaper struct {
	coeffs []float32 // b1, b2, a1, a2 per section (length = sosCount * 4)
	t1     []float32
	t2     []float32
}

// NewNoiseShaper creates a noise shaper with the given SOS coefficients.
// The coeffs slice contains 4 values per section: b1, b2, a1, a2.
// Its length must be a positive multiple of 4.
func NewNoiseShaper(coeffs []float32) *NoiseShaper {
	sosCount := len(coeffs) / 4

	return &NoiseShaper{
		coeffs: coeffs,
		t1:     make([]float32, sosCount),
		t2:     make([]float32, sosCount),
	}
}

// Clone returns a deep copy of the noise shaper with identical state.
// The coefficients slice is shared (not copied).
func (ns *NoiseShaper) Clone() *NoiseShaper {
	clone := &NoiseShaper{
		coeffs: ns.coeffs,
		t1:     make([]float32, len(ns.t1)),
		t2:     make([]float32, len(ns.t2)),
	}
	copy(clone.t1, ns.t1)
	copy(clone.t2, ns.t2)

	return clone
}

// Get computes the next noise shaping correction sample.
// Get and Update must be called in strict alternation.
func (ns *NoiseShaper) Get() float32 {
	var acc float32

	coeffs := ns.coeffs
	sosCount := len(ns.t1)

	for i := range sosCount {
		t1i := ns.t1[i]
		t2i := ns.t2[i]

		ns.t2[i] = acc - t1i*coeffs[2] - t2i*coeffs[3]
		acc = ns.t2[i] + t1i*coeffs[0] + t2i*coeffs[1]
		coeffs = coeffs[4:]
	}

	return acc
}

// Update feeds the quantization error back into the noise shaper state.
// Get and Update must be called in strict alternation.
func (ns *NoiseShaper) Update(qerror float32) {
	for i := range ns.t2 {
		ns.t2[i] += qerror
	}

	ns.t1, ns.t2 = ns.t2, ns.t1
}
