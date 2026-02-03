package dsd

// Converter4to1 performs DSD-to-PCM conversion with 4:1 decimation using
// direct bit-level FIR convolution. Used exclusively for DSD64 -> 705.6 kHz.
//
// Unlike the byte-level Converter (8:1), this operates on individual bits
// because the decimation factor (4) doesn't align with byte boundaries.
// Each input byte contains 8 DSD bits and produces 2 output samples
// (8 bits / 4 bits-per-sample = 2 samples).
type Converter4to1 struct {
	taps    []float64 // full symmetric FIR coefficients
	ring    []float64 // circular buffer of ±1.0 bit values
	ringPos int
}

func newConverter4to1() *Converter4to1 {
	halfLen := len(fir4to1Taps)
	fullLen := 2*halfLen - 1

	// Expand right-half to full symmetric coefficients.
	taps := make([]float64, fullLen)
	for i := range halfLen {
		taps[halfLen-1+i] = fir4to1Taps[i]
		taps[halfLen-1-i] = fir4to1Taps[i]
	}

	return &Converter4to1{
		taps: taps,
		ring: make([]float64, fullLen),
	}
}

// Reset clears the converter state to silence.
func (c *Converter4to1) Reset() {
	for i := range c.ring {
		c.ring[i] = 0
	}

	c.ringPos = 0
}

// Clone returns a deep copy of the converter with identical state.
func (c *Converter4to1) Clone() *Converter4to1 {
	clone := &Converter4to1{
		taps:    c.taps, // shared, immutable
		ring:    make([]float64, len(c.ring)),
		ringPos: c.ringPos,
	}
	copy(clone.ring, c.ring)

	return clone
}

// Translate converts DSD octets to PCM float32 samples with 4:1 decimation.
//
// Each input byte produces 2 output samples (8 bits / 4 = 2).
//
// Parameters:
//   - samples: number of input octets to process
//   - src: input DSD data
//   - srcStride: increment between consecutive input octets
//   - lsbFirst: true if least-significant bit is first in each octet
//   - dst: output PCM samples (float32)
//   - dstStride: increment between consecutive output samples
//
// Returns the number of output samples produced (= samples * 2).
//
//revive:disable-next-line:flag-parameter // lsbFirst is inherent to DSD container format distinction.
func (c *Converter4to1) Translate(
	samples int,
	src []byte, srcStride int,
	lsbFirst bool,
	dst []float32, dstStride int,
) int {
	ringLen := len(c.ring)
	tapsLen := len(c.taps)
	srcIdx := 0
	dstIdx := 0

	for range samples {
		octet := src[srcIdx]
		srcIdx += srcStride

		// Process 8 bits, producing a sample every 4 bits.
		for bit := 7; bit >= 0; bit-- {
			var bitVal float64

			if lsbFirst {
				bitVal = float64((octet>>(7-bit))&1)*2 - 1
			} else {
				bitVal = float64((octet>>bit)&1)*2 - 1
			}

			c.ring[c.ringPos] = bitVal
			c.ringPos = (c.ringPos + 1) % ringLen

			// Output a sample every 4 bits (at bits 4 and 0 of each byte).
			if (7-bit)%4 == 3 {
				// Convolve: newest sample is at ringPos-1, oldest at ringPos.
				acc := 0.0

				for k := range tapsLen {
					idx := (c.ringPos + k) % ringLen
					acc += c.taps[k] * c.ring[idx]
				}

				dst[dstIdx] = float32(acc)
				dstIdx += dstStride
			}
		}
	}

	return dstIdx / dstStride
}
