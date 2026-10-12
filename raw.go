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
	"io"
)

// RawReader reads a DSF or DFF file and delivers raw per-channel DSD data
// without any DSD-to-PCM conversion. This is intended for DoP (DSD over PCM)
// or other transport paths that need the original DSD bitstream.
//
// The container format is auto-detected from the file magic bytes.
type RawReader struct {
	reader io.ReadSeeker

	dsdRate  InputRate
	bitOrder BitOrder
	channels int

	chunkSize        int
	dataRemain       int64
	dsdBytesPerCh    int64
	dsdBytesConsumed int64

	readBuf        []byte
	chBufs         [][]byte
	deinterleaveFn func([]byte, [][]byte, int, int)
}

// NewRawReader creates a raw DSD reader from a DSF or DFF container.
//
// The container format is auto-detected from magic bytes. On return the
// reader is positioned at the start of the DSD data payload.
func NewRawReader(reader io.ReadSeeker) (*RawReader, error) {
	dsdFmt, chunkSize, dataSize, deinterleaveFn, err := parseContainer(reader)
	if err != nil {
		return nil, err
	}

	inputRate, ok := rateFromHz(dsdFmt.SampleRate)
	if !ok {
		return nil, fmt.Errorf("%w: unsupported DSD sample rate %d", ErrUnsupportedRate, dsdFmt.SampleRate)
	}

	nCh := int(dsdFmt.Channels)

	chBufs := make([][]byte, nCh)
	for ch := range nCh {
		chBufs[ch] = make([]byte, chunkSize)
	}

	return &RawReader{
		reader:         reader,
		dsdRate:        inputRate,
		bitOrder:       dsdFmt.BitOrder,
		channels:       nCh,
		chunkSize:      chunkSize,
		dataRemain:     dataSize,
		dsdBytesPerCh:  (dsdFmt.SamplesPerChannel + 7) / 8,
		readBuf:        make([]byte, chunkSize*nCh),
		chBufs:         chBufs,
		deinterleaveFn: deinterleaveFn,
	}, nil
}

// DSDRate returns the DSD sample rate in Hz (e.g. 2822400 for DSD64).
func (r *RawReader) DSDRate() InputRate {
	return r.dsdRate
}

// BitOrder returns the bit ordering of the DSD data (MSBFirst for DFF, LSBFirst for DSF).
func (r *RawReader) BitOrder() BitOrder {
	return r.bitOrder
}

// Channels returns the number of audio channels.
func (r *RawReader) Channels() int {
	return r.channels
}

// ReadChunk reads the next chunk of raw DSD data from the container and
// returns it as per-channel byte slices.
//
// n is the number of valid DSD bytes per channel (padding trimmed).
// The returned slices are valid until the next call to ReadChunk.
// Returns [io.EOF] when all DSD data has been consumed.
func (r *RawReader) ReadChunk() (channels [][]byte, n int, err error) {
	if r.dataRemain <= 0 {
		return nil, 0, io.EOF
	}

	// Calculate interleaved read size, rounded down to channel multiple.
	readSize := min(int64(r.chunkSize*r.channels), r.dataRemain)
	readSize -= readSize % int64(r.channels)

	buf := r.readBuf[:readSize]
	if _, err := io.ReadFull(r.reader, buf); err != nil {
		return nil, 0, fmt.Errorf("%w: reading DSD data: %w", ErrDecode, err)
	}

	r.dataRemain -= readSize

	// De-interleave into per-channel buffers.
	blockBytes := int(readSize) / r.channels
	r.deinterleaveFn(buf, r.chBufs, r.channels, blockBytes)

	// Trim to true DSD byte count (DSF blocks may have padding).
	chBytes := blockBytes
	remaining := int(r.dsdBytesPerCh - r.dsdBytesConsumed)

	if chBytes > remaining {
		chBytes = remaining
	}

	r.dsdBytesConsumed += int64(chBytes)

	return r.chBufs, chBytes, nil
}
