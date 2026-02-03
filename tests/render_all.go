//go:build ignore

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

package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	dsd "github.com/mycophonic/saprobe-dsd"
)

const (
	baseDir   = "/Users/dmp/Projects/go/farcloser/audio/saprobe-dsd/tests/_"
	sourceDir = "/Volumes/Anisotope"
)

func main() {
	saprobeDir := filepath.Join(baseDir, "saprobe")
	ffmpegDir := filepath.Join(baseDir, "ffmpeg")
	os.MkdirAll(saprobeDir, 0o755)
	os.MkdirAll(ffmpegDir, 0o755)

	var files []string
	filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".dsf" || ext == ".dff" {
			files = append(files, path)
		}
		return nil
	})

	fmt.Printf("Found %d DSD files\n\n", len(files))

	for i, path := range files {
		name := safeName(path)
		ext := strings.ToLower(filepath.Ext(path))
		fmt.Printf("[%d/%d] %s\n", i+1, len(files), name)

		file, err := os.Open(path)
		if err != nil {
			fmt.Printf("  SKIP (open: %v)\n\n", err)
			continue
		}

		var channels [][]byte
		var dsdFmt dsd.DSDFormat
		switch ext {
		case ".dsf":
			channels, dsdFmt, err = dsd.DecodeDSF(file)
		case ".dff":
			channels, dsdFmt, err = dsd.DecodeDFF(file)
		}
		file.Close()
		if err != nil {
			fmt.Printf("  SKIP (decode: %v)\n\n", err)
			continue
		}

		inputRate := dsd.InputRate(dsdFmt.SampleRate)
		lsbFirst := dsdFmt.BitOrder == dsd.LSBFirst
		numCh := len(channels)

		fmt.Printf("  %d Hz, %d ch, %d bytes/ch\n", dsdFmt.SampleRate, numCh, len(channels[0]))

		wavPath := filepath.Join(saprobeDir, name+".wav")
		if !exists(wavPath) {
			renderWAV(wavPath, channels, inputRate, lsbFirst, numCh)
		} else {
			fmt.Printf("  WAV: exists, skipped\n")
		}

		dopPath := filepath.Join(saprobeDir, name+".dop.wav")
		if !exists(dopPath) {
			renderDoP(dopPath, channels, inputRate, dsdFmt.BitOrder, numCh)
		} else {
			fmt.Printf("  DoP: exists, skipped\n")
		}

		ffPath := filepath.Join(ffmpegDir, name+".wav")
		if !exists(ffPath) {
			renderFFmpeg(ffPath, path)
		} else {
			fmt.Printf("  ffmpeg: exists, skipped\n")
		}

		fmt.Println()
	}
}

func safeName(path string) string {
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	r := strings.NewReplacer(" ", "-", ",", "", "'", "", "(", "", ")", "", "[", "", "]", "")
	name := r.Replace(base)
	name = strings.ToLower(name)
	for strings.Contains(name, "--") {
		name = strings.ReplaceAll(name, "--", "-")
	}
	return strings.Trim(name, "-")
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func renderWAV(outPath string, channels [][]byte, inputRate dsd.InputRate, lsbFirst bool, numCh int) {
	pipe, err := dsd.NewPipeline(inputRate, dsd.Rate352800)
	if err != nil {
		fmt.Printf("  WAV: ERROR pipeline: %v\n", err)
		return
	}

	pcmChannels := make([][]float32, numCh)
	for ch := range numCh {
		n := pipe.OutputSamples(len(channels[ch]))
		pcmChannels[ch] = make([]float32, n)
		got := pipe.Process(channels[ch], lsbFirst, pcmChannels[ch])
		pcmChannels[ch] = pcmChannels[ch][:got]
		pipe.Reset()
	}

	samplesPerCh := len(pcmChannels[0])
	dataSize := samplesPerCh * numCh * 4
	data := make([]byte, dataSize)
	for i := range samplesPerCh {
		for ch := range numCh {
			offset := (i*numCh + ch) * 4
			binary.LittleEndian.PutUint32(data[offset:], math.Float32bits(pcmChannels[ch][i]))
		}
	}

	writeWAV(outPath, data, 352800, numCh, 32, true)
	fmt.Printf("  WAV: %s (%d samples/ch)\n", filepath.Base(outPath), samplesPerCh)
}

func renderDoP(outPath string, channels [][]byte, inputRate dsd.InputRate, bitOrder dsd.BitOrder, numCh int) {
	dopRate, err := dsd.DoPRate(inputRate)
	if err != nil {
		fmt.Printf("  DoP: skipped (%v)\n", err)
		return
	}

	enc := dsd.NewDoPEncoder(bitOrder)
	outSize := dsd.DoPOutputSize(len(channels[0]), numCh)
	dst := make([]byte, outSize)

	n, err := enc.Encode(channels, dst)
	if err != nil {
		fmt.Printf("  DoP: ERROR encode: %v\n", err)
		return
	}

	writeWAV(outPath, dst[:n], dopRate, numCh, 24, false)
	fmt.Printf("  DoP: %s (%d Hz)\n", filepath.Base(outPath), dopRate)
}

func renderFFmpeg(outPath, srcPath string) {
	cmd := exec.Command("ffmpeg", "-hide_banner", "-loglevel", "error",
		"-i", srcPath,
		"-ar", "352800", "-c:a", "pcm_f32le",
		outPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("  ffmpeg: ERROR %v: %s\n", err, string(out))
		return
	}
	fmt.Printf("  ffmpeg: %s\n", filepath.Base(outPath))
}

func writeWAV(path string, data []byte, sampleRate, channels, bitDepth int, isFloat bool) {
	bytesPerSample := bitDepth / 8
	blockAlign := channels * bytesPerSample
	byteRate := sampleRate * blockAlign
	dataSize := len(data)

	var audioFormat uint16 = 1
	if isFloat {
		audioFormat = 3
	}

	buf := make([]byte, 44+dataSize)
	copy(buf[0:], "RIFF")
	binary.LittleEndian.PutUint32(buf[4:], uint32(36+dataSize))
	copy(buf[8:], "WAVE")
	copy(buf[12:], "fmt ")
	binary.LittleEndian.PutUint32(buf[16:], 16)
	binary.LittleEndian.PutUint16(buf[20:], audioFormat)
	binary.LittleEndian.PutUint16(buf[22:], uint16(channels))
	binary.LittleEndian.PutUint32(buf[24:], uint32(sampleRate))
	binary.LittleEndian.PutUint32(buf[28:], uint32(byteRate))
	binary.LittleEndian.PutUint16(buf[32:], uint16(blockAlign))
	binary.LittleEndian.PutUint16(buf[34:], uint16(bitDepth))
	copy(buf[36:], "data")
	binary.LittleEndian.PutUint32(buf[40:], uint32(dataSize))
	copy(buf[44:], data)

	os.WriteFile(path, buf, 0o644)
}
