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
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"github.com/mycophonic/saprobe-dsd/internal/dff"
	intdsd "github.com/mycophonic/saprobe-dsd/internal/dsd"
	"github.com/mycophonic/saprobe-dsd/internal/dsf"
	"github.com/mycophonic/saprobe-dsd/internal/format"
)

// magicSize is the number of bytes needed to detect the container format.
const magicSize = 4

// dffChunkSize is the DSD bytes per channel per read for DFF files,
// which have no natural block boundary. Matches the common DSF block size.
const dffChunkSize = 4096

// bytesPerFloat32 is the size of a float32 sample in bytes.
const bytesPerFloat32 = 4

// Decoder reads a DSF or DFF file and decodes DSD to interleaved float32
// little-endian PCM bytes. It implements [io.Reader].
//
// The container format is auto-detected from the file magic bytes.
// Each Read call returns a chunk of decoded PCM; the caller controls
// throughput by varying the buffer size passed to Read.
type Decoder struct {
	reader io.ReadSeeker
	format PCMFormat

	// Per-channel pipelines (stateful, one per channel).
	pipelines []*intdsd.Pipeline
	lsbFirst  bool

	// Container reading state.
	channels         int
	chunkSize        int   // DSD bytes per channel per read
	dataRemain       int64 // remaining interleaved DSD data bytes
	dsdBytesPerCh    int64 // true DSD bytes per channel (excl. padding)
	dsdBytesConsumed int64 // DSD bytes per channel consumed so far

	// Pre-allocated working buffers (zero-alloc after construction).
	readBuf []byte      // interleaved DSD bytes from container
	chBufs  [][]byte    // de-interleaved per-channel DSD
	pcmBufs [][]float32 // per-channel pipeline output

	// Output PCM buffer (interleaved f32le bytes), drained by Read.
	buf    []byte
	bufOff int
	eof    bool

	// De-interleave strategy (DSF block-interleaved vs DFF sample-interleaved).
	deinterleaveFn func(src []byte, dst [][]byte, channels, chunkSize int)
}

// NewDecoder creates a streaming DSD-to-PCM decoder.
//
// The container format (DSF or DFF) is auto-detected from the file magic bytes.
// outputRate specifies the target PCM sample rate (e.g. [Rate352800]).
func NewDecoder(reader io.ReadSeeker, outputRate OutputRate) (*Decoder, error) {
	dsdFmt, chunkSize, dataSize, deinterleaveFn, err := parseContainer(reader)
	if err != nil {
		return nil, err
	}

	return buildDecoder(reader, outputRate, dsdFmt, chunkSize, dataSize, deinterleaveFn)
}

// parseContainer detects the container format from magic bytes, parses headers,
// and returns the DSD format metadata. The reader is positioned at the start of
// the DSD data payload on return.
//
//revive:disable-next-line:function-result-limit // container parsing requires all return values.
func parseContainer(
	reader io.ReadSeeker,
) (format.DSDFormat, int, int64, func([]byte, [][]byte, int, int), error) {
	var magic [magicSize]byte
	if _, err := io.ReadFull(reader, magic[:]); err != nil {
		return format.DSDFormat{}, 0, 0, nil, fmt.Errorf("%w: reading magic: %w", ErrContainer, err)
	}

	if _, err := reader.Seek(0, io.SeekStart); err != nil {
		return format.DSDFormat{}, 0, 0, nil, fmt.Errorf("%w: seeking to start: %w", ErrContainer, err)
	}

	switch string(magic[:]) {
	case "DSD ":
		hdr, err := dsf.ParseHeader(reader)
		if err != nil {
			return format.DSDFormat{}, 0, 0, nil, fmt.Errorf("%w: %w", ErrContainer, err)
		}

		return hdr.Format, hdr.BlockSize, hdr.DataSize, deinterleaveDSF, nil

	case "FRM8":
		hdr, err := dff.ParseHeader(reader)
		if err != nil {
			return format.DSDFormat{}, 0, 0, nil, fmt.Errorf("%w: %w", ErrContainer, err)
		}

		return hdr.Format, dffChunkSize, hdr.DataSize, deinterleaveDFF, nil

	default:
		return format.DSDFormat{}, 0, 0, nil, fmt.Errorf("%w: unrecognized magic %q", ErrContainer, string(magic[:]))
	}
}

//revive:disable-next-line:argument-limit // construction requires all parameters.
func buildDecoder(
	reader io.ReadSeeker,
	outputRate OutputRate,
	dsdFmt format.DSDFormat,
	chunkSize int,
	dataSize int64,
	deinterleaveFn func([]byte, [][]byte, int, int),
) (*Decoder, error) {
	inputRate, ok := rateFromHz(dsdFmt.SampleRate)
	if !ok {
		return nil, fmt.Errorf("%w: unsupported DSD sample rate %d", ErrUnsupportedRate, dsdFmt.SampleRate)
	}

	nCh := int(dsdFmt.Channels) //nolint:gosec // channel count validated by container parser

	// Create one pipeline per channel.
	pipelines := make([]*intdsd.Pipeline, nCh)

	for channel := range nCh {
		pipe, err := intdsd.NewPipeline(inputRate, outputRate)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrUnsupportedRate, err)
		}

		pipelines[channel] = pipe
	}

	// Pre-allocate buffers.
	outSamples := pipelines[0].OutputSamples(chunkSize)

	chBufs := make([][]byte, nCh)
	pcmBufs := make([][]float32, nCh)

	for channel := range nCh {
		chBufs[channel] = make([]byte, chunkSize)
		pcmBufs[channel] = make([]float32, outSamples)
	}

	// True DSD bytes per channel (excluding block padding in DSF).
	dsdBytesPerCh := (dsdFmt.SamplesPerChannel + 7) / 8

	return &Decoder{
		reader:         reader,
		format:         PCMFormat{SampleRate: int(outputRate), BitDepth: 32, Channels: nCh},
		pipelines:      pipelines,
		lsbFirst:       dsdFmt.BitOrder == LSBFirst,
		channels:       nCh,
		chunkSize:      chunkSize,
		dataRemain:     dataSize,
		dsdBytesPerCh:  dsdBytesPerCh,
		readBuf:        make([]byte, chunkSize*nCh),
		chBufs:         chBufs,
		pcmBufs:        pcmBufs,
		buf:            make([]byte, 0, outSamples*nCh*bytesPerFloat32),
		deinterleaveFn: deinterleaveFn,
	}, nil
}

// Format returns the PCM output format.
func (d *Decoder) Format() PCMFormat {
	return d.format
}

// Read decodes DSD data and writes interleaved float32 little-endian PCM bytes
// to p. It implements [io.Reader].
func (d *Decoder) Read(p []byte) (int, error) { //nolint:varnamelen // p is idiomatic for io.Reader.Read
	total := 0

	for len(p) > 0 {
		// Phase 1: drain buffered PCM data.
		if d.bufOff < len(d.buf) {
			n := copy(p, d.buf[d.bufOff:])
			d.bufOff += n
			total += n
			p = p[n:]

			continue
		}

		// Phase 2: check EOF.
		if d.eof {
			if total > 0 {
				return total, nil
			}

			return 0, io.EOF
		}

		// Phase 3: decode next chunk.
		if err := d.decodeChunk(); err != nil {
			if total > 0 {
				return total, nil
			}

			return 0, err
		}
	}

	return total, nil
}

// Reset reinitializes the decoder for a new file, reusing pipelines and
// buffers when the format (DSD rate, output rate, channel count) matches.
//
// When format matches: zero allocation — pipelines are reset, buffers reused.
// When format differs: pipelines are rebuilt, buffers grow-only.
//
// If Reset returns an error, the decoder is in an undefined state.
// The caller should discard it or call Reset again with a valid file.
func (d *Decoder) Reset(reader io.ReadSeeker, outputRate OutputRate) error {
	dsdFmt, chunkSize, dataSize, deinterleaveFn, err := parseContainer(reader)
	if err != nil {
		return err
	}

	inputRate, ok := rateFromHz(dsdFmt.SampleRate)
	if !ok {
		return fmt.Errorf("%w: unsupported DSD sample rate %d", ErrUnsupportedRate, dsdFmt.SampleRate)
	}

	newChannels := int(dsdFmt.Channels) //nolint:gosec // channel count validated by container parser

	if err := d.resetPipelines(inputRate, outputRate, newChannels); err != nil {
		return err
	}

	d.growBuffers(newChannels, chunkSize)

	// Reset decoder state.
	d.reader = reader
	d.format = PCMFormat{SampleRate: int(outputRate), BitDepth: 32, Channels: newChannels}
	d.lsbFirst = dsdFmt.BitOrder == LSBFirst
	d.channels = newChannels
	d.chunkSize = chunkSize
	d.dataRemain = dataSize
	d.dsdBytesPerCh = (dsdFmt.SamplesPerChannel + 7) / 8
	d.dsdBytesConsumed = 0
	d.bufOff = 0
	d.eof = false
	d.deinterleaveFn = deinterleaveFn

	return nil
}

// resetPipelines resets or rebuilds the per-channel pipelines as needed.
// When the DSD/output rate combination matches the existing topology,
// pipelines are reused (zero-alloc). Otherwise they are rebuilt.
func (d *Decoder) resetPipelines(
	inputRate InputRate,
	outputRate OutputRate,
	newChannels int,
) error {
	if d.pipelines[0].DSDRate() == inputRate && d.pipelines[0].PCMRate() == outputRate {
		// Reuse existing pipelines, just clear audio state.
		for channel := range min(len(d.pipelines), newChannels) {
			d.pipelines[channel].Reset()
		}

		// Grow pipeline slice if channels increased.
		for channel := len(d.pipelines); channel < newChannels; channel++ {
			cloned := d.pipelines[0].Clone()
			cloned.Reset()
			d.pipelines = append(d.pipelines, cloned)
		}
	} else {
		// Topology changed — rebuild all pipelines.
		for channel := range newChannels {
			pipe, err := intdsd.NewPipeline(inputRate, outputRate)
			if err != nil {
				return fmt.Errorf("%w: %w", ErrUnsupportedRate, err)
			}

			if channel < len(d.pipelines) {
				d.pipelines[channel] = pipe
			} else {
				d.pipelines = append(d.pipelines, pipe)
			}
		}
	}

	// Trim slice length (but keep capacity for future growth).
	d.pipelines = d.pipelines[:newChannels]

	return nil
}

// growBuffers ensures all working buffers are large enough for the given
// channel count and chunk size, growing them if needed (never shrinking).
func (d *Decoder) growBuffers(newChannels, chunkSize int) {
	outSamples := d.pipelines[0].OutputSamples(chunkSize)

	// Grow per-channel buffer slices.
	for len(d.chBufs) < newChannels {
		d.chBufs = append(d.chBufs, nil)
	}

	d.chBufs = d.chBufs[:newChannels]

	for len(d.pcmBufs) < newChannels {
		d.pcmBufs = append(d.pcmBufs, nil)
	}

	d.pcmBufs = d.pcmBufs[:newChannels]

	// Grow individual per-channel buffers.
	for channel := range newChannels {
		d.chBufs[channel] = ensureLenBytes(d.chBufs[channel], chunkSize)
		d.pcmBufs[channel] = ensureLenF32(d.pcmBufs[channel], outSamples)
	}

	// Grow interleaved read buffer.
	d.readBuf = ensureLenBytes(d.readBuf, chunkSize*newChannels)

	// Grow interleaved output buffer capacity, but start empty (no decoded data yet).
	d.buf = ensureLenBytes(d.buf, outSamples*newChannels*bytesPerFloat32)
	d.buf = d.buf[:0]
}

// ensureLenBytes returns buf[:n] if capacity is sufficient, otherwise a new slice.
func ensureLenBytes(buf []byte, length int) []byte {
	if cap(buf) >= length {
		return buf[:length]
	}

	return make([]byte, length)
}

// ensureLenF32 returns buf[:n] if capacity is sufficient, otherwise a new slice.
func ensureLenF32(buf []float32, length int) []float32 {
	if cap(buf) >= length {
		return buf[:length]
	}

	return make([]float32, length)
}

func (d *Decoder) decodeChunk() error {
	if d.dataRemain <= 0 {
		d.eof = true

		return io.EOF
	}

	// Calculate how much interleaved data to read.
	// Round down to a multiple of channels so de-interleaving produces
	// equal-length per-channel slices (last chunk may be shorter).
	readSize := min(int64(d.chunkSize*d.channels), d.dataRemain)
	readSize -= readSize % int64(d.channels)

	// Read interleaved DSD data.
	buf := d.readBuf[:readSize]
	if _, err := io.ReadFull(d.reader, buf); err != nil {
		return fmt.Errorf("%w: reading DSD data: %w", ErrDecode, err)
	}

	d.dataRemain -= readSize

	// Calculate per-channel byte count for this chunk (block stride).
	blockBytes := int(readSize) / d.channels

	// De-interleave into per-channel buffers using full block stride.
	// Must happen before trimming: DSF blocks are [ch0_block][ch1_block]
	// and the stride between channels is always blockSize, not the trimmed count.
	d.deinterleaveFn(buf, d.chBufs, d.channels, blockBytes)

	// Trim to true DSD byte count per channel (DSF blocks may have padding).
	chBytes := blockBytes
	remaining := int(d.dsdBytesPerCh - d.dsdBytesConsumed)

	if chBytes > remaining {
		chBytes = remaining
	}

	d.dsdBytesConsumed += int64(chBytes)

	// Process each channel through its pipeline (trimmed, excluding padding).
	outSamples := d.pipelines[0].OutputSamples(chBytes)

	for ch := range d.channels {
		d.pipelines[ch].Process(d.chBufs[ch][:chBytes], d.lsbFirst, d.pcmBufs[ch][:outSamples])
	}

	// Interleave float32 output as f32le bytes.
	pcmBytes := outSamples * d.channels * bytesPerFloat32
	d.buf = d.buf[:pcmBytes]

	for i := range outSamples {
		for ch := range d.channels {
			off := (i*d.channels + ch) * bytesPerFloat32
			binary.LittleEndian.PutUint32(d.buf[off:], math.Float32bits(d.pcmBufs[ch][i]))
		}
	}

	d.bufOff = 0

	return nil
}

// deinterleaveDSF splits block-interleaved DSF data into per-channel slices.
// DSF stores data as: [ch0_block][ch1_block]...[chN_block].
func deinterleaveDSF(src []byte, dst [][]byte, channels, chunkSize int) {
	for ch := range channels {
		copy(dst[ch][:chunkSize], src[ch*chunkSize:])
	}
}

// deinterleaveDFF splits sample-interleaved DFF data into per-channel slices.
// DFF stores data as: [ch0_byte, ch1_byte, ch0_byte, ch1_byte, ...].
func deinterleaveDFF(src []byte, dst [][]byte, channels, chunkSize int) {
	for i := range chunkSize {
		for ch := range channels {
			dst[ch][i] = src[i*channels+ch]
		}
	}
}

// rateFromHz converts a DSD sample rate in Hz to an InputRate constant.
func rateFromHz(hz int) (InputRate, bool) {
	switch InputRate(hz) {
	case intdsd.DSD64, intdsd.DSD128, intdsd.DSD256, intdsd.DSD512, intdsd.DSD1024, intdsd.DSD2048:
		return InputRate(hz), true
	default:
		return 0, false
	}
}
