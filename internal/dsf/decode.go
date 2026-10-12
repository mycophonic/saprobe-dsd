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

package dsf

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	dfmt "github.com/mycophonic/saprobe-dsd/internal/format"
)

// Sentinel errors for DSF parsing failures.
var (
	ErrNotDSF             = errors.New("dsf: not a DSF file")
	ErrInvalidChunk       = errors.New("dsf: invalid chunk structure")
	ErrMissingFmt         = errors.New("dsf: missing fmt chunk")
	ErrMissingData        = errors.New("dsf: missing data chunk")
	ErrUnsupportedFormat  = errors.New("dsf: unsupported format")
	ErrInvalidBlockSize   = errors.New("dsf: invalid block size")
	ErrInvalidSampleCount = errors.New("dsf: invalid sample count")
)

// DSF chunk magic identifiers.
const (
	dsdChunkID  = "DSD "
	fmtChunkID  = "fmt "
	dataChunkID = "data"
)

// Fixed sizes from the DSF specification.
const (
	dsdChunkSize    = 28 // DSD header chunk is always 28 bytes.
	fmtMinSize      = 52 // fmt chunk minimum size.
	chunkHeaderSize = 12 // 4-byte ID + 8-byte size.

	bitsPerByte = 8
)

// dsfFmt holds the parsed format chunk fields.
type dsfFmt struct {
	channelCount int
	sampleRate   int
	sampleCount  int64 // DSD samples (bits) per channel.
	blockSize    int   // bytes per channel per block.
}

// Header holds parsed DSF metadata. After ParseHeader returns, the reader is
// positioned at the start of the DSD data payload.
type Header struct {
	Format    dfmt.DSDFormat
	BlockSize int   // bytes per channel per block
	DataSize  int64 // total DSD data bytes (all channels interleaved)
}

// ParseHeader reads DSF headers and positions the reader at the start of the
// DSD data payload. Use this for streaming decoding without reading all data.
func ParseHeader(reader io.ReadSeeker) (Header, error) {
	if err := parseDSDChunk(reader); err != nil {
		return Header{}, err
	}

	fmtInfo, err := parseFmtChunk(reader)
	if err != nil {
		return Header{}, err
	}

	dataSize, err := parseDataChunkHeader(reader)
	if err != nil {
		return Header{}, err
	}

	return Header{
		Format: dfmt.DSDFormat{
			SampleRate:        fmtInfo.sampleRate,
			Channels:          uint(fmtInfo.channelCount),
			SamplesPerChannel: fmtInfo.sampleCount,
			BitOrder:          dfmt.LSBFirst,
		},
		BlockSize: fmtInfo.blockSize,
		DataSize:  dataSize,
	}, nil
}

// Decode reads a DSF file and returns per-channel raw DSD bytes.
//
// Each element of the returned slice contains the DSD data for one channel.
// The bit ordering is LSB-first (as specified by the DSF format).
func Decode(reader io.ReadSeeker) ([][]byte, dfmt.DSDFormat, error) {
	hdr, err := ParseHeader(reader)
	if err != nil {
		return nil, dfmt.DSDFormat{}, err
	}

	fmtInfo := dsfFmt{
		channelCount: int(hdr.Format.Channels),
		sampleRate:   hdr.Format.SampleRate,
		sampleCount:  hdr.Format.SamplesPerChannel,
		blockSize:    hdr.BlockSize,
	}

	channels, err := readAndDeinterleave(reader, hdr.DataSize, fmtInfo)
	if err != nil {
		return nil, dfmt.DSDFormat{}, err
	}

	return channels, hdr.Format, nil
}

// parseDSDChunk reads and validates the 28-byte DSD header chunk.
func parseDSDChunk(reader io.ReadSeeker) error {
	var buf [dsdChunkSize]byte
	if _, err := io.ReadFull(reader, buf[:]); err != nil {
		return fmt.Errorf("%w: reading DSD chunk: %w", ErrNotDSF, err)
	}

	if string(buf[:4]) != dsdChunkID {
		return ErrNotDSF
	}

	chunkSize := binary.LittleEndian.Uint64(buf[4:12])
	if chunkSize != dsdChunkSize {
		return fmt.Errorf("%w: DSD chunk size %d, expected %d", ErrInvalidChunk, chunkSize, dsdChunkSize)
	}

	return nil
}

// parseFmtChunk reads and validates the fmt chunk.
func parseFmtChunk(reader io.ReadSeeker) (dsfFmt, error) {
	// Read chunk header: 4-byte ID + 8-byte size.
	var header [chunkHeaderSize]byte
	if _, err := io.ReadFull(reader, header[:]); err != nil {
		return dsfFmt{}, fmt.Errorf("%w: reading fmt header: %w", ErrMissingFmt, err)
	}

	if string(header[:4]) != fmtChunkID {
		return dsfFmt{}, fmt.Errorf("%w: expected \"fmt \", got %q", ErrMissingFmt, string(header[:4]))
	}

	chunkSize := binary.LittleEndian.Uint64(header[4:12])
	if chunkSize < fmtMinSize {
		return dsfFmt{}, fmt.Errorf("%w: fmt chunk too small (%d bytes)", ErrInvalidChunk, chunkSize)
	}

	// Read the fmt payload (already consumed the chunk header).
	payloadSize := chunkSize - chunkHeaderSize
	payload := make([]byte, payloadSize)

	if _, err := io.ReadFull(reader, payload); err != nil {
		return dsfFmt{}, fmt.Errorf("%w: reading fmt payload: %w", ErrInvalidChunk, err)
	}

	// Offsets within payload (relative to after the 12-byte header):
	// [0:4]   format version
	// [4:8]   format ID (0 = DSD raw)
	// [8:12]  channel type
	// [12:16] channel count
	// [16:20] sample rate
	// [20:24] bits per sample
	// [24:32] sample count per channel
	// [32:36] block size per channel
	formatID := binary.LittleEndian.Uint32(payload[4:8])
	if formatID != 0 {
		return dsfFmt{}, fmt.Errorf("%w: format ID %d (expected 0 = DSD raw)", ErrUnsupportedFormat, formatID)
	}

	channelCount := binary.LittleEndian.Uint32(payload[12:16])
	sampleRate := binary.LittleEndian.Uint32(payload[16:20])
	bitsPerSample := binary.LittleEndian.Uint32(payload[20:24])
	sampleCount := binary.LittleEndian.Uint64(payload[24:32])
	blockSize := binary.LittleEndian.Uint32(payload[32:36])

	if bitsPerSample != 1 && bitsPerSample != bitsPerByte {
		return dsfFmt{}, fmt.Errorf("%w: bits per sample %d (expected 1 or 8)", ErrUnsupportedFormat, bitsPerSample)
	}

	if blockSize == 0 {
		return dsfFmt{}, fmt.Errorf("%w: zero", ErrInvalidBlockSize)
	}

	if sampleCount == 0 {
		return dsfFmt{}, fmt.Errorf("%w: zero", ErrInvalidSampleCount)
	}

	if channelCount == 0 {
		return dsfFmt{}, fmt.Errorf("%w: zero channels", ErrUnsupportedFormat)
	}

	return dsfFmt{
		channelCount: int(channelCount),
		sampleRate:   int(sampleRate),
		sampleCount:  int64(sampleCount), // #nosec G115 -- DSD file sample counts fit in int64
		blockSize:    int(blockSize),
	}, nil
}

// parseDataChunkHeader reads the data chunk header and returns the payload size.
// The reader is positioned at the start of the DSD data payload on return.
func parseDataChunkHeader(reader io.ReadSeeker) (int64, error) {
	var header [chunkHeaderSize]byte
	if _, err := io.ReadFull(reader, header[:]); err != nil {
		return 0, fmt.Errorf("%w: reading data header: %w", ErrMissingData, err)
	}

	if string(header[:4]) != dataChunkID {
		return 0, fmt.Errorf("%w: expected \"data\", got %q", ErrMissingData, string(header[:4]))
	}

	chunkSize := binary.LittleEndian.Uint64(header[4:12])

	return int64(chunkSize - chunkHeaderSize), nil // #nosec G115 -- DSF data sizes fit in int64
}

// readAndDeinterleave reads the DSD data payload and de-interleaves per-channel blocks.
func readAndDeinterleave(reader io.ReadSeeker, dataSize int64, info dsfFmt) ([][]byte, error) {
	data := make([]byte, dataSize)
	if _, err := io.ReadFull(reader, data); err != nil {
		return nil, fmt.Errorf("%w: reading DSD data: %w", ErrInvalidChunk, err)
	}

	return deinterleaveBlocks(data, info)
}

// deinterleaveBlocks splits channel-blocked DSF data into per-channel slices.
//
// DSF stores data as repeating groups of per-channel blocks:
// [ch0 block][ch1 block]...[chN block][ch0 block][ch1 block]...
// Each block is blockSize bytes.
func deinterleaveBlocks(data []byte, info dsfFmt) ([][]byte, error) {
	blockSize := info.blockSize
	channelCount := info.channelCount
	groupSize := blockSize * channelCount

	// True byte count per channel, derived from the sample count.
	trueBytesPerChannel := (info.sampleCount + bitsPerByte - 1) / bitsPerByte

	channels := make([][]byte, channelCount)
	for ch := range channels {
		channels[ch] = make([]byte, 0, trueBytesPerChannel)
	}

	for offset := 0; offset+groupSize <= len(data); offset += groupSize {
		for ch := range channelCount {
			blockStart := offset + ch*blockSize
			channels[ch] = append(channels[ch], data[blockStart:blockStart+blockSize]...)
		}
	}

	// Trim each channel to the true byte count (last block may have padding).
	for ch := range channels {
		if int64(len(channels[ch])) > trueBytesPerChannel {
			channels[ch] = channels[ch][:trueBytesPerChannel]
		}
	}

	return channels, nil
}
