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
// from every channel and writes interleaved 24-bit little-endian PCM frames.
type DoPEncoder struct {
	enc *intdsd.DoPEncoder
}

// NewDoPEncoder creates a DoP encoder for the given bit ordering.
// LSB-first data (DSF) is automatically bit-reversed to MSB-first for DoP packing.
func NewDoPEncoder(bitOrder BitOrder) *DoPEncoder {
	return &DoPEncoder{enc: intdsd.NewDoPEncoder(bitOrder)}
}

// Reset restores the encoder to its initial state (marker = 0x05).
func (e *DoPEncoder) Reset() {
	e.enc.Reset()
}

// Encode packs per-channel DSD data into interleaved 24-bit LE DoP frames.
//
// All channel slices must have equal length, and that length must be even
// (each DoP frame consumes 2 DSD bytes). The marker alternation state is
// preserved across calls for streaming.
//
// dst must be at least [DoPOutputSize](len(channels[0]), len(channels)) bytes.
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
// will produce for the given number of DSD bytes per channel and channel count.
//
// Each pair of DSD bytes produces one DoP frame (3 bytes) per channel.
// dsdBytes must be even.
func DoPOutputSize(dsdBytes, channels int) int {
	return intdsd.OutputSize(dsdBytes, channels)
}
