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

package dff

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	dfmt "github.com/mycophonic/saprobe-dsd/internal/format"
)

// Sentinel errors for DSDIFF parsing failures.
var (
	ErrNotDFF             = errors.New("dff: not a DSDIFF file")
	ErrInvalidChunk       = errors.New("dff: invalid chunk structure")
	ErrMissingPROP        = errors.New("dff: missing PROP chunk")
	ErrMissingSampleRate  = errors.New("dff: missing FS property")
	ErrMissingChannels    = errors.New("dff: missing CHNL property")
	ErrMissingCompression = errors.New("dff: missing CMPR property")
	ErrDSTCompressed      = errors.New("dff: DST compression not supported")
	ErrMissingDSDData     = errors.New("dff: missing DSD data chunk")
)

// DSDIFF chunk identifiers.
const (
	frm8ID = "FRM8"
	propID = "PROP"
	fsID   = "FS  "
	chnlID = "CHNL"
	cmprID = "CMPR"
	dsdID  = "DSD "
)

// Fixed sizes from the DSDIFF specification.
const (
	// chunkHeaderSize is the size of a DSDIFF chunk header (4-byte ID + 8-byte size).
	chunkHeaderSize = 12

	// frm8HeaderSize is the total FRM8 header: ID(4) + size(8) + formType(4).
	frm8HeaderSize = 16

	bitsPerByte = 8
)

// dffProps holds the parsed PROP sub-chunk fields.
type dffProps struct {
	sampleRate   int
	channelCount int
}

// Header holds parsed DFF metadata. After ParseHeader returns, the reader is
// positioned at the start of the DSD data payload.
type Header struct {
	Format   dfmt.DSDFormat
	DataSize int64 // total DSD data bytes (all channels interleaved)
}

// ParseHeader reads DFF headers and positions the reader at the start of the
// DSD data payload. Use this for streaming decoding without reading all data.
func ParseHeader(reader io.ReadSeeker) (Header, error) {
	formEnd, err := parseFRM8Header(reader)
	if err != nil {
		return Header{}, err
	}

	props, dsdDataOffset, dsdDataSize, err := walkTopChunks(reader, frm8HeaderSize, formEnd)
	if err != nil {
		return Header{}, err
	}

	if dsdDataSize == 0 {
		return Header{}, ErrMissingDSDData
	}

	// Seek to start of DSD data payload.
	if _, err := reader.Seek(dsdDataOffset, io.SeekStart); err != nil {
		return Header{}, fmt.Errorf("%w: seeking to DSD data: %w", ErrInvalidChunk, err)
	}

	return Header{
		Format: dfmt.DSDFormat{
			SampleRate:        props.sampleRate,
			Channels:          uint(props.channelCount),
			SamplesPerChannel: dsdDataSize / int64(props.channelCount) * bitsPerByte,
			BitOrder:          dfmt.MSBFirst,
		},
		DataSize: dsdDataSize,
	}, nil
}

// Decode reads a DSDIFF file and returns per-channel raw DSD bytes.
//
// Each element of the returned slice contains the DSD data for one channel.
// The bit ordering is MSB-first (as specified by the DSDIFF format).
// DST-compressed files are not supported and return [ErrDSTCompressed].
func Decode(reader io.ReadSeeker) ([][]byte, dfmt.DSDFormat, error) {
	hdr, err := ParseHeader(reader)
	if err != nil {
		return nil, dfmt.DSDFormat{}, err
	}

	// Reader is already positioned at DSD data start from ParseHeader.
	data := make([]byte, hdr.DataSize)
	if _, err := io.ReadFull(reader, data); err != nil {
		return nil, dfmt.DSDFormat{}, fmt.Errorf("%w: reading DSD data: %w", ErrInvalidChunk, err)
	}

	channels := deinterleave(data, int(hdr.Format.Channels))

	return channels, hdr.Format, nil
}

// parseFRM8Header reads the 16-byte FRM8 header and returns the end offset of
// the form data (formDataStart + dataSize).
func parseFRM8Header(reader io.ReadSeeker) (int64, error) {
	var buf [frm8HeaderSize]byte
	if _, err := io.ReadFull(reader, buf[:]); err != nil {
		return 0, fmt.Errorf("%w: reading FRM8 header: %w", ErrNotDFF, err)
	}

	if string(buf[:4]) != frm8ID {
		return 0, ErrNotDFF
	}

	dataSize := binary.BigEndian.Uint64(buf[4:12])

	if string(buf[12:16]) != dsdID {
		return 0, fmt.Errorf("%w: form type %q, expected %q", ErrNotDFF, string(buf[12:16]), dsdID)
	}

	// dataSize covers everything after the ID+size fields (bytes 0-11),
	// so the form data starts at offset 12 and ends at 12 + dataSize.
	return int64(dataSize) + chunkHeaderSize, nil // #nosec G115 -- DSDIFF sizes fit in int64
}

// walkTopChunks iterates over the top-level chunks inside the FRM8 form,
// parsing FVER, PROP, and locating the DSD data chunk.
//
//revive:disable-next-line:function-result-limit // chunked format parsing requires offset+size+error returns.
func walkTopChunks(
	reader io.ReadSeeker,
	pos, end int64,
) (props dffProps, dsdOffset, dsdSize int64, err error) {
	var propFound bool

	for pos < end {
		chunkID, chunkDataSize, err := readChunkHeader(reader, pos)
		if err != nil {
			break
		}

		contentStart := pos + chunkHeaderSize

		switch chunkID {
		case propID:
			props, err = parsePROPChunk(reader, contentStart, chunkDataSize)
			if err != nil {
				return dffProps{}, 0, 0, err
			}

			propFound = true

		case dsdID:
			dsdOffset = contentStart
			dsdSize = chunkDataSize

		default:
			// Skip unknown and no-op chunks (e.g. FVER — we accept any version).
		}

		// Advance past chunk data, respecting 2-byte alignment padding.
		pos = contentStart + chunkDataSize
		if chunkDataSize%2 != 0 {
			pos++
		}
	}

	if !propFound {
		return dffProps{}, 0, 0, ErrMissingPROP
	}

	return props, dsdOffset, dsdSize, nil
}

// parsePROPChunk parses the PROP chunk and its sub-chunks (FS, CHNL, CMPR).
//
//nolint:gocognit // the sub-chunk walk is the DSDIFF format's: one case per sub-chunk, each required one tracked
func parsePROPChunk(reader io.ReadSeeker, contentStart, dataSize int64) (dffProps, error) {
	// First 4 bytes of PROP content are the property type ("SND ").
	var propType [4]byte

	if _, err := reader.Seek(contentStart, io.SeekStart); err != nil {
		return dffProps{}, fmt.Errorf("%w: seeking to PROP content: %w", ErrInvalidChunk, err)
	}

	if _, err := io.ReadFull(reader, propType[:]); err != nil {
		return dffProps{}, fmt.Errorf("%w: reading PROP type: %w", ErrInvalidChunk, err)
	}

	// Walk sub-chunks starting after the 4-byte property type.
	subStart := contentStart + 4
	subEnd := contentStart + dataSize

	var (
		props     dffProps
		fsFound   bool
		chnlFound bool
		cmprFound bool
	)

	pos := subStart
	for pos < subEnd {
		chunkID, chunkDataSize, err := readChunkHeader(reader, pos)
		if err != nil {
			break
		}

		chunkContentStart := pos + chunkHeaderSize

		switch chunkID {
		case fsID:
			props.sampleRate, err = parseFSChunk(reader, chunkContentStart)
			if err != nil {
				return dffProps{}, err
			}

			fsFound = true

		case chnlID:
			props.channelCount, err = parseCHNLChunk(reader, chunkContentStart)
			if err != nil {
				return dffProps{}, err
			}

			chnlFound = true

		case cmprID:
			if err := parseCMPRChunk(reader, chunkContentStart); err != nil {
				return dffProps{}, err
			}

			cmprFound = true

		default:
			// Skip unknown sub-chunks.
		}

		pos = chunkContentStart + chunkDataSize
		if chunkDataSize%2 != 0 {
			pos++
		}
	}

	if !fsFound {
		return dffProps{}, ErrMissingSampleRate
	}

	if !chnlFound {
		return dffProps{}, ErrMissingChannels
	}

	if !cmprFound {
		return dffProps{}, ErrMissingCompression
	}

	return props, nil
}

// parseFSChunk reads the FS (sample rate) sub-chunk: 4-byte uint32 BE.
func parseFSChunk(reader io.ReadSeeker, contentStart int64) (int, error) {
	if _, err := reader.Seek(contentStart, io.SeekStart); err != nil {
		return 0, fmt.Errorf("%w: seeking to FS: %w", ErrInvalidChunk, err)
	}

	var buf [4]byte
	if _, err := io.ReadFull(reader, buf[:]); err != nil {
		return 0, fmt.Errorf("%w: reading FS: %w", ErrInvalidChunk, err)
	}

	return int(binary.BigEndian.Uint32(buf[:])), nil
}

// parseCHNLChunk reads the CHNL sub-chunk: 2-byte uint16 BE channel count.
func parseCHNLChunk(reader io.ReadSeeker, contentStart int64) (int, error) {
	if _, err := reader.Seek(contentStart, io.SeekStart); err != nil {
		return 0, fmt.Errorf("%w: seeking to CHNL: %w", ErrInvalidChunk, err)
	}

	var buf [2]byte
	if _, err := io.ReadFull(reader, buf[:]); err != nil {
		return 0, fmt.Errorf("%w: reading CHNL: %w", ErrInvalidChunk, err)
	}

	return int(binary.BigEndian.Uint16(buf[:])), nil
}

// parseCMPRChunk reads the CMPR sub-chunk and validates the compression type.
// Only uncompressed DSD ("DSD ") is supported; DST returns ErrDSTCompressed.
func parseCMPRChunk(reader io.ReadSeeker, contentStart int64) error {
	if _, err := reader.Seek(contentStart, io.SeekStart); err != nil {
		return fmt.Errorf("%w: seeking to CMPR: %w", ErrInvalidChunk, err)
	}

	var buf [4]byte
	if _, err := io.ReadFull(reader, buf[:]); err != nil {
		return fmt.Errorf("%w: reading CMPR: %w", ErrInvalidChunk, err)
	}

	compressionType := string(buf[:])

	if compressionType == "DST " {
		return ErrDSTCompressed
	}

	if compressionType != dsdID {
		return fmt.Errorf("%w: compression type %q", ErrInvalidChunk, compressionType)
	}

	return nil
}

// deinterleave splits sample-interleaved DSD data into per-channel slices.
func deinterleave(data []byte, channelCount int) [][]byte {
	bytesPerChannel := len(data) / channelCount
	channels := make([][]byte, channelCount)

	for ch := range channels {
		channels[ch] = make([]byte, bytesPerChannel)
	}

	for idx := 0; idx < len(data)-channelCount+1; idx += channelCount {
		group := idx / channelCount
		for ch := range channelCount {
			channels[ch][group] = data[idx+ch]
		}
	}

	return channels
}

// readChunkHeader reads a DSDIFF chunk header at the given position.
// Returns the 4-byte chunk ID and the data size (excluding the header).
func readChunkHeader(reader io.ReadSeeker, pos int64) (string, int64, error) {
	if _, err := reader.Seek(pos, io.SeekStart); err != nil {
		return "", 0, fmt.Errorf("%w: seeking to chunk at %d: %w", ErrInvalidChunk, pos, err)
	}

	var buf [chunkHeaderSize]byte
	if _, err := io.ReadFull(reader, buf[:]); err != nil {
		return "", 0, fmt.Errorf("%w: reading chunk header at %d: %w", ErrInvalidChunk, pos, err)
	}

	chunkID := string(buf[:4])
	dataSize := int64(binary.BigEndian.Uint64(buf[4:12])) // #nosec G115 -- DSDIFF sizes fit in int64

	return chunkID, dataSize, nil
}
