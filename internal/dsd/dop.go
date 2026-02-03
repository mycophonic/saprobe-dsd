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

	"github.com/mycophonic/saprobe-dsd/internal/format"
)

// DoP marker bytes per the dCS DoP v1.1 specification.
const (
	dopMarkerA = 0x05
	dopMarkerB = 0xFA
)

// bytesPerDoPFrame is the size of one DoP sample: 3 bytes (24-bit LE).
const bytesPerDoPFrame = 3

// dsdBytesPerDoPFrame is how many DSD bytes each DoP frame consumes.
const dsdBytesPerDoPFrame = 2

// ErrDoPUnsupportedRate indicates a DSD rate that cannot be transported via DoP.
var ErrDoPUnsupportedRate = errors.New("dop: DSD rate not supported (DSD512+ exceeds DoP transport capacity)")

// ErrDoPInvalidInput indicates malformed input to the DoP encoder.
var ErrDoPInvalidInput = errors.New("dop: invalid input")

// DoPRate returns the PCM sample rate required to transport the given DSD rate
// via DoP. Returns [ErrDoPUnsupportedRate] for DSD512 and above.
func DoPRate(dsdRate InputRate) (int, error) {
	// DoP PCM rate = DSD rate / 16 (each 24-bit frame carries 16 DSD bits).
	//nolint:exhaustive // DSD512+ explicitly rejected below; unknown rates fall through to error.
	switch dsdRate {
	case DSD64:
		return int(Rate176400), nil
	case DSD128:
		return int(Rate352800), nil
	case DSD256:
		return int(Rate705600), nil
	default:
		return 0, fmt.Errorf("%w: %s", ErrDoPUnsupportedRate, dsdRate)
	}
}

// DoPFormat describes the PCM container parameters for DoP output.
type DoPFormat struct {
	// SampleRate is the PCM frame rate (176400, 352800, or 705600).
	SampleRate int
	// BitDepth is always 24.
	BitDepth int
	// Channels is the number of audio channels.
	Channels int
}

// DoPEncoder packs raw DSD data into DoP (DSD over PCM) frames.
//
// It is stateful: the marker alternation carries across successive Encode calls,
// enabling streaming use. Each Encode call consumes an equal number of DSD bytes
// from every channel and writes interleaved 24-bit little-endian PCM frames.
type DoPEncoder struct {
	marker   byte
	lsbFirst bool
}

// NewDoPEncoder creates a DoP encoder for the given bit ordering.
// LSB-first data (DSF) is automatically bit-reversed to MSB-first for DoP packing.
func NewDoPEncoder(bitOrder format.BitOrder) *DoPEncoder {
	initTables() // ensure bitReverse table is populated

	return &DoPEncoder{
		marker:   dopMarkerA,
		lsbFirst: bitOrder == format.LSBFirst,
	}
}

// Reset restores the encoder to its initial state (marker = 0x05).
func (e *DoPEncoder) Reset() {
	e.marker = dopMarkerA
}

// OutputSize returns the number of output bytes that Encode will produce
// for the given number of DSD bytes per channel and channel count.
//
// Each pair of DSD bytes produces one DoP frame (3 bytes) per channel.
// dsdBytes must be even.
func OutputSize(dsdBytes, channels int) int {
	return (dsdBytes / dsdBytesPerDoPFrame) * bytesPerDoPFrame * channels
}

// Encode packs per-channel DSD data into interleaved 24-bit LE DoP frames.
//
// All channel slices must have equal length, and that length must be even
// (each DoP frame consumes 2 DSD bytes). The marker alternation state is
// preserved across calls for streaming.
//
// dst must be at least [OutputSize](len(channels[0]), len(channels)) bytes.
// Returns the number of bytes written to dst.
func (e *DoPEncoder) Encode(channels [][]byte, dst []byte) (int, error) {
	if len(channels) == 0 {
		return 0, fmt.Errorf("%w: no channels", ErrDoPInvalidInput)
	}

	chLen := len(channels[0])

	for idx := 1; idx < len(channels); idx++ {
		if len(channels[idx]) != chLen {
			return 0, fmt.Errorf("%w: channel length mismatch", ErrDoPInvalidInput)
		}
	}

	if chLen%dsdBytesPerDoPFrame != 0 {
		return 0, fmt.Errorf("%w: DSD byte count must be even (got %d)", ErrDoPInvalidInput, chLen)
	}

	needed := OutputSize(chLen, len(channels))
	if len(dst) < needed {
		return 0, fmt.Errorf("%w: dst too small (need %d, got %d)", ErrDoPInvalidInput, needed, len(dst))
	}

	nch := len(channels)
	pos := 0

	// Process DSD bytes in pairs: each pair → one DoP frame per channel.
	for srcIdx := 0; srcIdx < chLen; srcIdx += dsdBytesPerDoPFrame {
		for ch := range nch {
			dsdHigh := channels[ch][srcIdx]
			dsdLow := channels[ch][srcIdx+1]

			if e.lsbFirst {
				dsdHigh = bitReverse[dsdHigh]
				dsdLow = bitReverse[dsdLow]
			}

			// 24-bit little-endian: [low byte] [high byte] [marker]
			dst[pos] = dsdLow
			dst[pos+1] = dsdHigh
			dst[pos+2] = e.marker

			pos += bytesPerDoPFrame
		}

		// Alternate marker after each frame position (all channels share the same marker per frame).
		if e.marker == dopMarkerA {
			e.marker = dopMarkerB
		} else {
			e.marker = dopMarkerA
		}
	}

	return pos, nil
}
