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

// dsd-example-decoder decodes a DSF or DFF file to WAV or raw PCM on stdout.
//
// Usage:
//
//	dsd-example-decoder [-format wav|pcm] [-rate 352800] <input.dsf|input.dff | ->
//
//nolint:gosec // Integer conversions bounded by audio format constraints; file paths from CLI args.
package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"math"
	"os"

	dsd "github.com/mycophonic/saprobe-dsd"
	"github.com/mycophonic/saprobe-dsd/version"
)

const formatWAV = "wav"

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	outputFormat := flag.String("format", formatWAV, "output format: wav or pcm")
	outputRate := flag.Int("rate", int(dsd.Rate352800), "PCM output sample rate in Hz")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-format wav|pcm] [-rate Hz] <input | ->\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if *showVersion {
		fmt.Fprintln(os.Stdout, version.String())
		os.Exit(0)
	}

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	if *outputFormat != formatWAV && *outputFormat != "pcm" {
		fmt.Fprintf(os.Stderr, "unknown format %q (use wav or pcm)\n", *outputFormat)
		os.Exit(1)
	}

	os.Exit(run(*outputFormat, dsd.OutputRate(*outputRate), flag.Arg(0)))
}

func run(outputFormat string, outRate dsd.OutputRate, inputPath string) int {
	reader, cleanup, err := openInput(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)

		return 1
	}

	defer cleanup()

	// Try DSF first, then DFF.
	channels, dsdFmt, err := dsd.DecodeDSF(reader)
	if err != nil {
		// Rewind and try DFF.
		if _, seekErr := reader.Seek(0, io.SeekStart); seekErr != nil {
			fmt.Fprintf(os.Stderr, "seek error: %v\n", seekErr)

			return 1
		}

		channels, dsdFmt, err = dsd.DecodeDFF(reader)
		if err != nil {
			fmt.Fprintf(os.Stderr, "decode: %v\n", err)

			return 1
		}
	}

	pcm, pcmFmt, err := convertToPCM(channels, dsdFmt, outRate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "convert: %v\n", err)

		return 1
	}

	fmt.Fprintf(os.Stderr, "%d Hz, %d-bit, %d ch, %d bytes PCM\n",
		pcmFmt.SampleRate, pcmFmt.BitDepth, pcmFmt.Channels, len(pcm))

	if outputFormat == formatWAV {
		err = writeWAV(os.Stdout, pcm, pcmFmt)
	} else {
		_, err = os.Stdout.Write(pcm)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "write: %v\n", err)

		return 1
	}

	return 0
}

//nolint:gosec // Integer conversions bounded by audio format constraints.
func convertToPCM(channels [][]byte, dsdFmt dsd.DSDFormat, outRate dsd.OutputRate) ([]byte, dsd.PCMFormat, error) {
	inputRate := dsd.InputRate(dsdFmt.SampleRate)
	lsbFirst := dsdFmt.BitOrder == dsd.LSBFirst
	channelCount := int(dsdFmt.Channels)

	pipe, err := dsd.NewPipeline(inputRate, outRate)
	if err != nil {
		return nil, dsd.PCMFormat{}, fmt.Errorf("creating pipeline: %w", err)
	}

	// Process each channel.
	outSamples := pipe.OutputSamples(len(channels[0]))
	pcmChannels := make([][]float32, channelCount)

	for channel := range channelCount {
		pcmChannels[channel] = make([]float32, outSamples)

		chPipe := pipe
		if channel > 0 {
			chPipe = pipe.Clone()
		}

		chPipe.Process(channels[channel], lsbFirst, pcmChannels[channel])
	}

	// Interleave to 32-bit float LE PCM bytes.
	totalSamples := outSamples * channelCount
	pcm := make([]byte, totalSamples*4) //nolint:mnd // 4 bytes per float32

	for sample := range outSamples {
		for channel := range channelCount {
			offset := (sample*channelCount + channel) * 4
			binary.LittleEndian.PutUint32(pcm[offset:], math.Float32bits(pcmChannels[channel][sample]))
		}
	}

	pcmFmt := dsd.PCMFormat{
		SampleRate: int(outRate),
		BitDepth:   32,
		Channels:   channelCount,
	}

	return pcm, pcmFmt, nil
}

// openInput returns a ReadSeeker for the given path, or buffers stdin when path is "-".
func openInput(path string) (io.ReadSeeker, func(), error) {
	if path == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, func() {}, fmt.Errorf("reading stdin: %w", err)
		}

		return bytes.NewReader(data), func() {}, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, func() {}, fmt.Errorf("opening %s: %w", path, err)
	}

	return f, func() { _ = f.Close() }, nil
}

// writeWAV writes a standard PCM WAV (32-bit float) to writer.
func writeWAV(writer io.Writer, pcm []byte, pcmFmt dsd.PCMFormat) error {
	bytesPerSample := pcmFmt.BitDepth / 8
	blockAlign := pcmFmt.Channels * bytesPerSample
	byteRate := pcmFmt.SampleRate * blockAlign
	dataSize := len(pcm)

	var hdr [44]byte

	copy(hdr[0:4], "RIFF")
	binary.LittleEndian.PutUint32(hdr[4:8], uint32(36+dataSize))
	copy(hdr[8:12], "WAVE")

	copy(hdr[12:16], "fmt ")
	binary.LittleEndian.PutUint32(hdr[16:20], 16)
	binary.LittleEndian.PutUint16(hdr[20:22], 3) // IEEE float
	binary.LittleEndian.PutUint16(hdr[22:24], uint16(pcmFmt.Channels))
	binary.LittleEndian.PutUint32(hdr[24:28], uint32(pcmFmt.SampleRate))
	binary.LittleEndian.PutUint32(hdr[28:32], uint32(byteRate))
	binary.LittleEndian.PutUint16(hdr[32:34], uint16(blockAlign))
	binary.LittleEndian.PutUint16(hdr[34:36], uint16(pcmFmt.BitDepth))

	copy(hdr[36:40], "data")
	binary.LittleEndian.PutUint32(hdr[40:44], uint32(dataSize))

	if _, err := writer.Write(hdr[:]); err != nil {
		return fmt.Errorf("writing WAV header: %w", err)
	}

	if _, err := writer.Write(pcm); err != nil {
		return fmt.Errorf("writing WAV data: %w", err)
	}

	return nil
}
