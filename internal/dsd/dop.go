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

// dsdSilenceByte is the DSD idle pattern (alternating 1/0 bits) in MSB-first order.
// Used to pad the final DoP frame when the input has an odd number of DSD bytes.
const dsdSilenceByte = 0x69

// ErrDoPUnsupportedRate indicates a DSD rate that cannot be transported via DoP.
var ErrDoPUnsupportedRate = errors.New("dop: DSD rate not supported (DSD512+ exceeds DoP transport capacity)")

// ErrDoPInvalidInput indicates malformed input to the DoP encoder.
var ErrDoPInvalidInput = errors.New("dop: invalid input")

// DoPRate returns the PCM sample rate required to transport the given DSD rate
// via DoP. Returns [ErrDoPUnsupportedRate] for DSD512 and above.
func DoPRate(dsdRate InputRate) (int, error) {
	// DoP PCM rate = DSD rate / 16 (each 24-bit frame carries 16 DSD bits).
	//
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
// from every channel and writes interleaved little-endian PCM frames.
//
// When containerBytes > 3 (e.g. 4 for 32-bit), Encode writes padded frames
// directly with high-aligned DoP data: [0x00 padding..., dsdLow, dsdHigh, marker].
// This eliminates the need for a separate padding pass.
type DoPEncoder struct {
	marker         byte
	lsbFirst       bool
	containerBytes int // output bytes per sample; 0 or 3 = 24-bit, 4 = 32-bit, etc.
}

// NewDoPEncoder creates a DoP encoder for the given bit ordering and container size.
// LSB-first data (DSF) is automatically bit-reversed to MSB-first for DoP packing.
//
// containerBytes sets the output sample width. Use 0 or 3 for standard 24-bit DoP
// frames. Use 4 for 32-bit containers (the encoder writes high-aligned padded frames).
func NewDoPEncoder(bitOrder format.BitOrder, containerBytes int) *DoPEncoder {
	initTables() // ensure bitReverse table is populated

	return &DoPEncoder{
		marker:         dopMarkerA,
		lsbFirst:       bitOrder == format.LSBFirst,
		containerBytes: containerBytes,
	}
}

// Reset restores the encoder to its initial state (marker = 0x05).
func (e *DoPEncoder) Reset() {
	e.marker = dopMarkerA
}

// OutputSize returns the number of output bytes that Encode will produce
// for the given number of DSD bytes per channel, channel count, and container size.
//
// Each pair of DSD bytes produces one DoP frame per channel. containerBytes sets
// the output sample width (0 or 3 = 24-bit, 4 = 32-bit, etc.).
// When dsdBytes is odd, the final byte is paired with a DSD silence pad byte (0x69).
func OutputSize(dsdBytes, channels, containerBytes int) int {
	frames := (dsdBytes + 1) / dsdBytesPerDoPFrame
	bpf := effectiveContainerBytes(containerBytes)

	return frames * bpf * channels
}

// Consumed converts output bytes written by Encode back to DSD bytes
// consumed per channel. This is the inverse of OutputSize.
func Consumed(outputBytes, channels, containerBytes int) int {
	bpf := effectiveContainerBytes(containerBytes)
	frames := outputBytes / (bpf * channels)

	return frames * dsdBytesPerDoPFrame
}

// effectiveContainerBytes returns the actual bytes per sample, defaulting
// to bytesPerDoPFrame (3) when containerBytes is 0 or less than 3.
func effectiveContainerBytes(containerBytes int) int {
	if containerBytes > bytesPerDoPFrame {
		return containerBytes
	}

	return bytesPerDoPFrame
}

// Encode packs per-channel DSD data into interleaved DoP frames.
//
// All channel slices must have equal length. When the length is odd, the final
// byte is paired with a DSD silence pad byte (0x69) to complete the last frame.
// The marker alternation state is preserved across calls for streaming.
//
// When the encoder's containerBytes > 3, output frames are padded to the container
// size with high-aligned DoP data: [0x00 padding..., dsdLow, dsdHigh, marker].
//
// dst must be at least [OutputSize](len(channels[0]), len(channels), containerBytes) bytes.
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

	bpf := effectiveContainerBytes(e.containerBytes)
	needed := OutputSize(chLen, len(channels), e.containerBytes)

	if len(dst) < needed {
		return 0, fmt.Errorf("%w: dst too small (need %d, got %d)", ErrDoPInvalidInput, needed, len(dst))
	}

	nch := len(channels)
	pad := bpf - bytesPerDoPFrame // 0 for 24-bit, 1 for 32-bit
	pos := 0

	// Process DSD bytes in pairs: each pair → one DoP frame per channel.
	// When chLen is odd, the last byte is paired with DSD silence (0x69).
	for srcIdx := 0; srcIdx < chLen; srcIdx += dsdBytesPerDoPFrame {
		padLow := srcIdx+1 >= chLen // odd length: last byte needs silence pad

		for chIdx := range nch {
			dsdHigh := channels[chIdx][srcIdx]

			if e.lsbFirst {
				dsdHigh = bitReverse[dsdHigh]
			}

			var dsdLow byte
			if padLow {
				// Pad with DSD silence in MSB-first (DoP native order), skip bit-reversal.
				dsdLow = dsdSilenceByte
			} else {
				dsdLow = channels[chIdx][srcIdx+1]
				if e.lsbFirst {
					dsdLow = bitReverse[dsdLow]
				}
			}

			// Write padding bytes (zero) in LSBs for containers wider than 24 bits.
			for j := range pad {
				dst[pos+j] = 0
			}

			// DoP frame (high-aligned in container): [dsdLow, dsdHigh, marker]
			dst[pos+pad] = dsdLow
			dst[pos+pad+1] = dsdHigh
			dst[pos+pad+2] = e.marker

			pos += bpf
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
