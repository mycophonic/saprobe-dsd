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
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	dsd "github.com/mycophonic/saprobe-dsd"
	"github.com/mycophonic/saprobe-dsd/internal/dff"
	intdsd "github.com/mycophonic/saprobe-dsd/internal/dsd"
	"github.com/mycophonic/saprobe-dsd/internal/dsf"
	"github.com/mycophonic/saprobe-dsd/internal/format"
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
		fmt.Printf("[%d/%d] %s\n", i+1, len(files), name)

		file, err := os.Open(path)
		if err != nil {
			fmt.Printf("  SKIP (open: %v)\n\n", err)
			continue
		}

		// WAV rendering: use streaming decoder.
		wavPath := filepath.Join(saprobeDir, name+".wav")
		if !exists(wavPath) {
			renderWAV(wavPath, file)
		} else {
			fmt.Printf("  WAV: exists, skipped\n")
		}

		// DoP rendering: needs raw DSD data from container.
		if _, seekErr := file.Seek(0, io.SeekStart); seekErr != nil {
			fmt.Printf("  SKIP (seek: %v)\n\n", seekErr)
			file.Close()
			continue
		}

		dopPath := filepath.Join(saprobeDir, name+".dop.wav")
		if !exists(dopPath) {
			renderDoP(dopPath, file, path)
		} else {
			fmt.Printf("  DoP: exists, skipped\n")
		}

		file.Close()

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

func renderWAV(outPath string, file *os.File) {
	dec, err := dsd.NewDecoder(file, dsd.Rate352800)
	if err != nil {
		fmt.Printf("  WAV: ERROR decoder: %v\n", err)
		return
	}

	pcm, err := io.ReadAll(dec)
	if err != nil {
		fmt.Printf("  WAV: ERROR read: %v\n", err)
		return
	}

	pcmFmt := dec.Format()
	samplesPerCh := len(pcm) / (pcmFmt.Channels * 4)

	writeWAV(outPath, pcm, 352800, pcmFmt.Channels, 32, true)
	fmt.Printf("  WAV: %s (%d samples/ch)\n", filepath.Base(outPath), samplesPerCh)
}

func renderDoP(dopPath string, file *os.File, path string) {
	ext := strings.ToLower(filepath.Ext(path))

	var channels [][]byte
	var dsdFmt format.DSDFormat
	var err error

	switch ext {
	case ".dsf":
		channels, dsdFmt, err = dsf.Decode(file)
	case ".dff":
		channels, dsdFmt, err = dff.Decode(file)
	default:
		fmt.Printf("  DoP: SKIP (unknown ext %s)\n", ext)
		return
	}

	if err != nil {
		fmt.Printf("  DoP: SKIP (decode: %v)\n", err)
		return
	}

	inputRate := intdsd.InputRate(dsdFmt.SampleRate)
	numCh := len(channels)

	dopRate, err := dsd.DoPRate(dsd.InputRate(inputRate))
	if err != nil {
		fmt.Printf("  DoP: skipped (%v)\n", err)
		return
	}

	enc := dsd.NewDoPEncoder(dsd.BitOrder(dsdFmt.BitOrder))
	outSize := dsd.DoPOutputSize(len(channels[0]), numCh)
	dst := make([]byte, outSize)

	n, err := enc.Encode(channels, dst)
	if err != nil {
		fmt.Printf("  DoP: ERROR encode: %v\n", err)
		return
	}

	writeWAV(dopPath, dst[:n], dopRate, numCh, 24, false)
	fmt.Printf("  DoP: %s (%d Hz)\n", filepath.Base(dopPath), dopRate)
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
