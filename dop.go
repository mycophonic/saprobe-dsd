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

// DoPFormat describes the PCM container parameters for DoP output.
type DoPFormat = intdsd.DoPFormat

// DoPEncoder packs raw DSD data into DoP (DSD over PCM) frames.
//
// It is stateful: the marker alternation carries across successive Encode calls,
// enabling streaming use. Each Encode call consumes an equal number of DSD bytes
// from every channel and writes interleaved little-endian PCM frames.
//
// When containerBytes > 3, output frames are padded to the container size with
// high-aligned DoP data, eliminating the need for a separate padding pass.
type DoPEncoder struct {
	enc *intdsd.DoPEncoder
}

// NewDoPEncoder creates a DoP encoder for the given bit ordering and container size.
// LSB-first data (DSF) is automatically bit-reversed to MSB-first for DoP packing.
//
// containerBytes sets the output sample width. Use 0 or 3 for standard 24-bit DoP
// frames. Use 4 for 32-bit containers (the encoder writes high-aligned padded frames).
func NewDoPEncoder(bitOrder BitOrder, containerBytes int) *DoPEncoder {
	return &DoPEncoder{enc: intdsd.NewDoPEncoder(bitOrder, containerBytes)}
}

// Reset restores the encoder to its initial state (marker = 0x05).
func (e *DoPEncoder) Reset() {
	e.enc.Reset()
}

// Encode packs per-channel DSD data into interleaved DoP frames.
//
// All channel slices must have equal length. When the length is odd, the final
// byte is paired with a DSD silence pad byte (0x69) to complete the last frame.
// The marker alternation state is preserved across calls for streaming.
//
// dst must be at least [DoPOutputSize](len(channels[0]), len(channels), containerBytes) bytes.
// Returns the number of bytes written to dst.
func (e *DoPEncoder) Encode(channels [][]byte, dst []byte) (int, error) {
	n, err := e.enc.Encode(channels, dst)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrDoP, err)
	}

	return n, nil
}

// DoPRate returns the PCM sample rate required to transport the given DSD rate
// via DoP. Returns an error for DSD512 and above, which exceed DoP transport capacity.
func DoPRate(dsdRate InputRate) (int, error) {
	rate, err := intdsd.DoPRate(dsdRate)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrDoP, err)
	}

	return rate, nil
}

// DoPOutputSize returns the number of output bytes that [DoPEncoder.Encode]
// will produce for the given number of DSD bytes per channel, channel count,
// and container size.
//
// Each pair of DSD bytes produces one DoP frame per channel. containerBytes
// sets the output sample width (0 or 3 = 24-bit, 4 = 32-bit, etc.).
// When dsdBytes is odd, the final byte is paired with a DSD silence pad byte.
func DoPOutputSize(dsdBytes, channels, containerBytes int) int {
	return intdsd.OutputSize(dsdBytes, channels, containerBytes)
}

// DoPDSDConsumed converts output bytes written by [DoPEncoder.Encode] back
// to DSD bytes consumed per channel. This is the inverse of [DoPOutputSize].
func DoPDSDConsumed(outputBytes, channels, containerBytes int) int {
	return intdsd.Consumed(outputBytes, channels, containerBytes)
}
