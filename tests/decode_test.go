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
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mycophonic/agar/pkg/agar"

	dsd "github.com/mycophonic/saprobe-dsd"
)

// DSD-to-PCM comparison tolerances.
//
// Unlike lossless codecs (ALAC, FLAC) or even lossy codecs with a standard
// synthesis filterbank (MP3), DSD-to-PCM conversion has no standard algorithm.
// Different implementations use different FIR filters, decimation strategies,
// and group delays. Sample-level comparison is meaningless; instead we compare
// statistical properties of the decoded audio.
//
// Empirical measurements across DSF (DSD128, DSD256) and DFF (DSD128) files:
//   - DSD128: RMS diff 0.4-1.8%, peak diff 1.5-2.2%
//   - DSD256: RMS diff 5.1%, peak diff 4.9-5.2%
//
// Higher DSD rates require more decimation stages (halfband filters), causing
// greater divergence between saprobe's FIR+halfband cascade and ffmpeg's
// resampler. Tolerances accommodate the worst case with margin.
const (
	// dsdRMSTolerance is the maximum allowed RMS difference between saprobe and
	// ffmpeg, expressed as a fraction of the ffmpeg RMS value.
	dsdRMSTolerance = 0.10

	// dsdPeakTolerance is the maximum allowed peak difference between saprobe and
	// ffmpeg, expressed as a fraction of the ffmpeg peak value.
	dsdPeakTolerance = 0.10

	// dsdSampleCountTolerance is the maximum allowed difference in output sample
	// count between saprobe and ffmpeg. A difference of 1 sample occurs when the
	// DSD byte count per channel is odd: saprobe's halfband uses floor decimation
	// (discards the unpaired trailing sample) while ffmpeg's resampler uses ceiling
	// decimation (zero-pads to produce one extra sample). See docs/IMPLEMENTATION.md.
	dsdSampleCountTolerance = 1

	// dsdBytesPerFloat32 is the size of a float32 sample in bytes.
	dsdBytesPerFloat32 = 4
)

// TestDecodeFile decodes a real DSF or DFF file from disk using the streaming
// decoder and compares against ffmpeg.
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

	dec, err := dsd.NewDecoder(file, dsd.Rate352800)
	if err != nil {
		t.Fatalf("new decoder: %v", err)
	}

	pcmBytes, err := io.ReadAll(dec)
	if err != nil {
		t.Fatalf("read all: %v", err)
	}

	pcmFmt := dec.Format()
	channelCount := pcmFmt.Channels

	t.Logf("  Format: %d Hz, %d-bit, %d ch, %d bytes PCM",
		pcmFmt.SampleRate, pcmFmt.BitDepth, channelCount, len(pcmBytes))

	// Parse interleaved f32le bytes into per-channel float32 slices.
	saprobeChannels := parseInterleavedF32(pcmBytes, channelCount)

	for chIdx := range channelCount {
		peak := peakAbsF32(saprobeChannels[chIdx])
		rmsVal := rmsFloat32(saprobeChannels[chIdx])

		t.Logf("  saprobe ch%d: %d samples, peak=%.4f, rms=%.6f",
			chIdx, len(saprobeChannels[chIdx]), peak, rmsVal)

		if peak > 10.0 {
			t.Errorf("ch%d: peak %.4f exceeds reasonable range", chIdx, peak)
		}

		const minRMS = 0.0001
		if rmsVal < minRMS {
			t.Errorf("ch%d: RMS %.6f suspiciously low for real audio", chIdx, rmsVal)
		}

		if hasNaNOrInfF32(saprobeChannels[chIdx]) {
			t.Errorf("ch%d: output contains NaN or Inf", chIdx)
		}
	}

	// Cross-compare with ffmpeg.
	ffmpegChannels := decodeWithFFmpeg(t, filePath, channelCount)
	compareDSDOutputs(t, saprobeChannels, ffmpegChannels)
}

// parseInterleavedF32 splits interleaved f32le bytes into per-channel float32 slices.
func parseInterleavedF32(pcmBytes []byte, channelCount int) [][]float32 {
	samplesPerChannel := len(pcmBytes) / dsdBytesPerFloat32 / channelCount
	channels := make([][]float32, channelCount)

	for ch := range channelCount {
		channels[ch] = make([]float32, samplesPerChannel)
	}

	for idx := range samplesPerChannel {
		for ch := range channelCount {
			byteOffset := (idx*channelCount + ch) * dsdBytesPerFloat32
			bits := binary.LittleEndian.Uint32(pcmBytes[byteOffset:])
			channels[ch][idx] = math.Float32frombits(bits)
		}
	}

	return channels
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

	parsed := parseInterleavedF32(ffmpegResult.Stdout, channelCount)

	for chIdx := range channelCount {
		samplesPerChannel := len(parsed[chIdx])
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
