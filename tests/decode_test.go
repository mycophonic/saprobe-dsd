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

package tests_test

import (
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mycophonic/agar/pkg/agar"

	dsd "github.com/mycophonic/saprobe-dsd"
)

// Supported DSD sample rates for validation.
//
//nolint:gochecknoglobals
var validDSDRates = map[int]dsd.InputRate{
	int(dsd.DSD64):   dsd.DSD64,
	int(dsd.DSD128):  dsd.DSD128,
	int(dsd.DSD256):  dsd.DSD256,
	int(dsd.DSD512):  dsd.DSD512,
	int(dsd.DSD1024): dsd.DSD1024,
}

// DSD-to-PCM comparison tolerances.
//
// Unlike lossless codecs (ALAC, FLAC) or even lossy codecs with a standard
// synthesis filterbank (MP3), DSD-to-PCM conversion has no standard algorithm.
// Different implementations use different FIR filters, decimation strategies,
// and group delays. Sample-level comparison is meaningless; instead we compare
// statistical properties of the decoded audio.
//
// Empirical measurements across DSF (DSD128, DSD256) and DFF (DSD128) files
// show RMS differences of 0.4-1.8% and peak differences of 1.5-2.2% between
// saprobe and ffmpeg. Tolerances are set with margin above observed values.
const (
	// dsdRMSTolerance is the maximum allowed RMS difference between saprobe and
	// ffmpeg, expressed as a fraction of the ffmpeg RMS value.
	dsdRMSTolerance = 0.05

	// dsdPeakTolerance is the maximum allowed peak difference between saprobe and
	// ffmpeg, expressed as a fraction of the ffmpeg peak value.
	dsdPeakTolerance = 0.10

	// dsdSampleCountTolerance is the maximum allowed difference in output sample
	// count between saprobe and ffmpeg. A difference of 1 sample can occur from
	// rounding in decimation chains with non-power-of-2 ratios.
	dsdSampleCountTolerance = 2

	// dsdBytesPerFloat32 is the size of a float32 sample in bytes.
	dsdBytesPerFloat32 = 4
)

// TestDecodeFile decodes a real DSF or DFF file from disk, validates the
// container parse and DSD-to-PCM pipeline output, and compares against ffmpeg.
//
// Set BENCH_FILE to a .dsf or .dff file path to run.
//
//nolint:paralleltest // File I/O benchmark, run sequentially.
func TestDecodeFile(t *testing.T) {
	filePath := os.Getenv("BENCH_FILE")
	if filePath == "" {
		t.Skip("set BENCH_FILE to run this test")
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".dsf" && ext != ".dff" {
		t.Fatalf("unsupported extension %q: expected .dsf or .dff", ext)
	}

	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	defer file.Close()

	t.Logf("File: %s", filePath)

	channels, dsdFmt, err := decodeContainer(file, ext)
	if err != nil {
		t.Fatalf("decode container: %v", err)
	}

	validateFormat(t, dsdFmt, len(channels))
	validateChannelData(t, channels)

	inputRate := validDSDRates[dsdFmt.SampleRate]
	lsbFirst := dsdFmt.BitOrder == dsd.LSBFirst

	saprobeChannels := decodePipeline(t, channels, inputRate, lsbFirst)

	// Cross-compare with ffmpeg.
	ffmpegChannels := decodeWithFFmpeg(t, filePath, int(dsdFmt.Channels))
	compareDSDOutputs(t, saprobeChannels, ffmpegChannels)
}

func decodeContainer(file *os.File, ext string) ([][]byte, dsd.DSDFormat, error) {
	switch ext {
	case ".dsf":
		return dsd.DecodeDSF(file)
	default:
		return dsd.DecodeDFF(file)
	}
}

func validateFormat(t *testing.T, dsdFmt dsd.DSDFormat, channelCount int) {
	t.Helper()

	t.Logf("  Format: %d Hz, %d channels, %d samples/ch, bitOrder=%d",
		dsdFmt.SampleRate, dsdFmt.Channels, dsdFmt.SamplesPerChannel, dsdFmt.BitOrder)

	if _, valid := validDSDRates[dsdFmt.SampleRate]; !valid {
		t.Fatalf("unexpected DSD sample rate: %d", dsdFmt.SampleRate)
	}

	if dsdFmt.Channels == 0 {
		t.Fatal("channel count is 0")
	}

	if dsdFmt.SamplesPerChannel <= 0 {
		t.Fatal("samples per channel is 0")
	}

	if int(dsdFmt.Channels) != channelCount {
		t.Fatalf("format says %d channels, got %d channel slices", dsdFmt.Channels, channelCount)
	}
}

func validateChannelData(t *testing.T, channels [][]byte) {
	t.Helper()

	channelLen := len(channels[0])
	if channelLen == 0 {
		t.Fatal("channel 0 has no data")
	}

	for chIdx, chData := range channels {
		if len(chData) != channelLen {
			t.Fatalf("channel %d length %d != channel 0 length %d", chIdx, len(chData), channelLen)
		}
	}

	t.Logf("  Channels: %d, bytes/channel: %d", len(channels), channelLen)
}

// decodePipeline runs the DSD-to-PCM pipeline on each channel and validates output.
func decodePipeline(
	t *testing.T,
	channels [][]byte,
	inputRate dsd.InputRate,
	lsbFirst bool,
) [][]float32 {
	t.Helper()

	pipe, err := dsd.NewPipeline(inputRate, dsd.Rate352800)
	if err != nil {
		t.Fatalf("NewPipeline(%s, 352.8kHz): %v", inputRate, err)
	}

	result := make([][]float32, len(channels))

	for chIdx, chData := range channels {
		expectedOut := pipe.OutputSamples(len(chData))
		pcmOut := make([]float32, expectedOut)

		got := pipe.Process(chData, lsbFirst, pcmOut)
		if got != expectedOut {
			t.Errorf("ch%d: pipeline produced %d samples, expected %d", chIdx, got, expectedOut)
		}

		if hasNaNOrInfF32(pcmOut[:got]) {
			t.Errorf("ch%d: output contains NaN or Inf", chIdx)
		}

		peak := peakAbsF32(pcmOut[:got])
		rmsVal := rmsFloat32(pcmOut[:got])

		t.Logf("  saprobe ch%d: %d samples, peak=%.4f, rms=%.6f", chIdx, got, peak, rmsVal)

		if peak > 10.0 {
			t.Errorf("ch%d: peak %.4f exceeds reasonable range", chIdx, peak)
		}

		const minRMS = 0.0001
		if rmsVal < minRMS {
			t.Errorf("ch%d: RMS %.6f suspiciously low for real audio", chIdx, rmsVal)
		}

		result[chIdx] = pcmOut[:got]
		pipe.Reset()
	}

	return result
}

// decodeWithFFmpeg decodes a DSD file to float32 PCM at 352.8 kHz using ffmpeg.
//
// Uses agar.FFmpeg directly (not FFmpegDecode) because the DSD-to-PCM decode
// requires float32 output format which differs from the integer PCM formats
// that FFmpegDecode supports.
func decodeWithFFmpeg(t *testing.T, filePath string, channelCount int) [][]float32 {
	t.Helper()

	ffmpegResult := agar.FFmpeg(t, agar.FFmpegOptions{
		Args: []string{
			"-i", filePath,
			"-ar", "352800",
			"-ac", "2",
			"-f", "f32le",
			"-acodec", "pcm_f32le",
			"-",
		},
	})

	ffmpegPCM := ffmpegResult.Stdout

	samplesPerChannel := len(ffmpegPCM) / dsdBytesPerFloat32 / channelCount
	parsed := make([][]float32, channelCount)

	for ch := range channelCount {
		parsed[ch] = make([]float32, samplesPerChannel)
	}

	for idx := range samplesPerChannel {
		for ch := range channelCount {
			byteOffset := (idx*channelCount + ch) * dsdBytesPerFloat32
			bits := binary.LittleEndian.Uint32(ffmpegPCM[byteOffset:])
			parsed[ch][idx] = math.Float32frombits(bits)
		}
	}

	for chIdx := range channelCount {
		peak := peakAbsF32(parsed[chIdx])
		rmsVal := rmsFloat32(parsed[chIdx])

		t.Logf("  ffmpeg  ch%d: %d samples, peak=%.4f, rms=%.6f", chIdx, samplesPerChannel, peak, rmsVal)
	}

	return parsed
}

// compareDSDOutputs compares saprobe and ffmpeg DSD-to-PCM output using
// statistical properties (RMS energy, peak level) rather than per-sample matching.
//
// DSD-to-PCM conversion has no standard algorithm. Different implementations use
// different FIR filters, decimation strategies, and group delays. The decoded PCM
// represents the same audio content but will differ at the sample level.
func compareDSDOutputs(t *testing.T, saprobe, ffmpeg [][]float32) {
	t.Helper()

	if len(saprobe) != len(ffmpeg) {
		t.Fatalf("channel count mismatch: saprobe=%d, ffmpeg=%d", len(saprobe), len(ffmpeg))
	}

	for ch := range saprobe {
		sLen := len(saprobe[ch])
		fLen := len(ffmpeg[ch])

		lenDiff := sLen - fLen
		if lenDiff < 0 {
			lenDiff = -lenDiff
		}

		if lenDiff > dsdSampleCountTolerance {
			t.Errorf("ch%d: sample count diff %d exceeds tolerance %d (saprobe=%d, ffmpeg=%d)",
				ch, lenDiff, dsdSampleCountTolerance, sLen, fLen)
		}

		compareLen := min(sLen, fLen)
		sRMS := float64(rmsFloat32(saprobe[ch][:compareLen]))
		fRMS := float64(rmsFloat32(ffmpeg[ch][:compareLen]))
		sPeak := float64(peakAbsF32(saprobe[ch][:compareLen]))
		fPeak := float64(peakAbsF32(ffmpeg[ch][:compareLen]))

		rmsDiff := math.Abs(sRMS-fRMS) / fRMS
		peakDiff := math.Abs(sPeak-fPeak) / fPeak

		t.Logf("  compare ch%d: rms_diff=%.2f%%, peak_diff=%.2f%%",
			ch, rmsDiff*100, peakDiff*100)

		if rmsDiff > dsdRMSTolerance {
			t.Errorf("ch%d: RMS diff %.2f%% exceeds tolerance %.0f%% (saprobe=%.6f, ffmpeg=%.6f)",
				ch, rmsDiff*100, dsdRMSTolerance*100, sRMS, fRMS)
		}

		if peakDiff > dsdPeakTolerance {
			t.Errorf("ch%d: peak diff %.2f%% exceeds tolerance %.0f%% (saprobe=%.4f, ffmpeg=%.4f)",
				ch, peakDiff*100, dsdPeakTolerance*100, sPeak, fPeak)
		}
	}
}

func hasNaNOrInfF32(samples []float32) bool {
	for _, sample := range samples {
		if math.IsNaN(float64(sample)) || math.IsInf(float64(sample), 0) {
			return true
		}
	}

	return false
}

func peakAbsF32(samples []float32) float32 {
	var peak float32

	for _, sample := range samples {
		absVal := float32(math.Abs(float64(sample)))
		if absVal > peak {
			peak = absVal
		}
	}

	return peak
}

func rmsFloat32(samples []float32) float32 {
	if len(samples) == 0 {
		return 0
	}

	var sumSq float64

	for _, sample := range samples {
		sumSq += float64(sample) * float64(sample)
	}

	return float32(math.Sqrt(sumSq / float64(len(samples))))
}
