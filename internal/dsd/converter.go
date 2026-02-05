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

// silencePattern is a DSD byte that produces minimal audible energy.
// On repeat it makes a low energy 352.8 kHz tone and a high energy
// 1.0584 MHz tone, both filtered out by any playback system.
const silencePattern = 0x69

// Converter performs DSD-to-PCM conversion for a single channel using
// a symmetric 96-tap FIR lowpass filter with 8:1 decimation.
type Converter struct {
	fifo    [fifoSize]byte
	fifoPos uint
}

// NewConverter creates a new DSD-to-PCM converter initialized to silence.
func NewConverter() *Converter {
	initTables()

	c := &Converter{}
	c.Reset()

	return c
}

// Reset reinitializes the converter state to silence.
func (c *Converter) Reset() {
	for i := range c.fifo {
		c.fifo[i] = silencePattern
	}

	c.fifoPos = 0
}

// Clone returns a deep copy of the converter with identical state.
func (c *Converter) Clone() *Converter {
	cp := *c

	return &cp
}

// Translate converts DSD octets to PCM float64 samples (8:1 decimation).
//
// Each input octet produces one output float sample. For multi-channel
// interleaved data, use srcStride and dstStride to skip between channels.
//
// Parameters:
//   - samples: number of octets to process
//   - src: input DSD data
//   - srcStride: increment between consecutive input octets for this channel
//   - lsbFirst: true if least-significant bit is first in each octet
//   - dst: output PCM samples (float64, nominally in [-1, +1])
//   - dstStride: increment between consecutive output samples for this channel
//
//revive:disable-next-line:flag-parameter // lsbFirst is inherent to DSD container format distinction.
func (c *Converter) Translate(
	samples int,
	src []byte, srcStride int,
	lsbFirst bool,
	dst []float64, dstStride int,
) {
	ffp := c.fifoPos
	srcIdx := 0
	dstIdx := 0

	for range samples {
		bite1 := uint(src[srcIdx]) & 0xFF
		if lsbFirst {
			bite1 = uint(bitReverse[bite1])
		}

		c.fifo[ffp] = byte(bite1)
		srcIdx += srcStride

		p := (ffp - ctablesLen) & fifoMask
		c.fifo[p] = bitReverse[c.fifo[p]&0xFF]

		acc := 0.0

		for i := range uint(ctablesLen) {
			b1 := c.fifo[(ffp-i)&fifoMask] & 0xFF
			b2 := c.fifo[(ffp-(ctablesLen*2-1)+i)&fifoMask] & 0xFF
			acc += ctables[i][b1] + ctables[i][b2]
		}

		dst[dstIdx] = acc
		dstIdx += dstStride
		ffp = (ffp + 1) & fifoMask
	}

	c.fifoPos = ffp
}
