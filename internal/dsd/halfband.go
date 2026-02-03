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

// Halfband performs 2:1 decimation using a symmetric FIR halfband filter.
//
// The filter exploits the halfband property: every other coefficient (except
// the center tap) is zero, halving the number of multiplications.
type Halfband struct {
	taps    []float64 // right-half coefficients (including center tap)
	ring    []float64 // circular buffer (length = full filter length)
	ringPos int
}

func newHalfband() *Halfband {
	n := len(halfbandTaps)
	fullLen := 2*n - 1 // total filter length (odd)

	taps := make([]float64, n)
	copy(taps, halfbandTaps[:])

	return &Halfband{
		taps: taps,
		ring: make([]float64, fullLen),
	}
}

// Reset clears the filter state.
func (h *Halfband) Reset() {
	for i := range h.ring {
		h.ring[i] = 0
	}

	h.ringPos = 0
}

// Clone returns a deep copy of the halfband filter with identical state.
func (h *Halfband) Clone() *Halfband {
	clone := &Halfband{
		taps:    h.taps, // shared, immutable
		ring:    make([]float64, len(h.ring)),
		ringPos: h.ringPos,
	}
	copy(clone.ring, h.ring)

	return clone
}

// Process decimates the input by 2:1. The input length must be even.
// Returns the output slice (length = len(src) / 2).
//
// The dst slice is used as output storage if it has sufficient capacity;
// otherwise a new slice is allocated.
func (h *Halfband) Process(src, dst []float64) []float64 {
	outLen := len(src) / 2
	if cap(dst) >= outLen {
		dst = dst[:outLen]
	} else {
		dst = make([]float64, outLen)
	}

	ringLen := len(h.ring)
	halfLen := len(h.taps) // right-half length including center

	for sample := range outLen {
		// Push two input samples into the ring buffer.
		h.ring[h.ringPos] = src[sample*2]
		h.ringPos = (h.ringPos + 1) % ringLen
		h.ring[h.ringPos] = src[sample*2+1]
		h.ringPos = (h.ringPos + 1) % ringLen

		// Compute output: center tap + symmetric pairs.
		// The newest sample is at ringPos-1, oldest at ringPos.
		// Center tap is at offset halfLen-1 from oldest.
		centerIdx := (h.ringPos + halfLen - 1) % ringLen
		acc := h.taps[0] * h.ring[centerIdx]

		// Symmetric pairs: taps[k] multiplies (sample[center-k] + sample[center+k])
		// for k = 1, 2, ..., halfLen-1.
		// But halfband property means every other non-center tap is zero.
		// Non-zero taps are at odd offsets from center: k = 1, 3, 5, ...
		for k := 1; k < halfLen; k += 2 {
			left := (centerIdx - k + ringLen) % ringLen
			right := (centerIdx + k) % ringLen
			acc += h.taps[k] * (h.ring[left] + h.ring[right])
		}

		dst[sample] = acc
	}

	return dst
}
