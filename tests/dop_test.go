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
	"errors"
	"testing"

	dsd "github.com/mycophonic/saprobe-dsd"
)

// TestDoPRate verifies the DSD-to-DoP rate mapping.
func TestDoPRate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input   dsd.InputRate
		want    int
		wantErr bool
	}{
		{dsd.DSD64, 176_400, false},
		{dsd.DSD128, 352_800, false},
		{dsd.DSD256, 705_600, false},
		{dsd.DSD512, 0, true},
		{dsd.DSD1024, 0, true},
	}

	for _, tt := range tests {
		rate, err := dsd.DoPRate(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("DoPRate(%s): expected error, got %d", tt.input, rate)
			}

			if !errors.Is(err, dsd.ErrDoP) {
				t.Errorf("DoPRate(%s): error %v does not wrap ErrDoP", tt.input, err)
			}

			continue
		}

		if err != nil {
			t.Errorf("DoPRate(%s): unexpected error: %v", tt.input, err)

			continue
		}

		if rate != tt.want {
			t.Errorf("DoPRate(%s) = %d, want %d", tt.input, rate, tt.want)
		}
	}
}

// TestDoPMarkerAlternation verifies that DoP markers alternate 0x05/0xFA
// correctly across frames and across multiple Encode calls.
func TestDoPMarkerAlternation(t *testing.T) {
	t.Parallel()

	enc := dsd.NewDoPEncoder(dsd.MSBFirst)

	// 8 DSD bytes → 4 DoP frames (mono).
	src := [][]byte{{0xAA, 0x55, 0xAA, 0x55, 0xAA, 0x55, 0xAA, 0x55}}
	dst := make([]byte, dsd.DoPOutputSize(len(src[0]), 1))

	n, err := enc.Encode(src, dst)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	if n != len(dst) {
		t.Fatalf("Encode returned %d bytes, want %d", n, len(dst))
	}

	// Marker is the 3rd byte of each 3-byte frame.
	expectedMarkers := []byte{0x05, 0xFA, 0x05, 0xFA}
	for i, want := range expectedMarkers {
		got := dst[i*3+2]
		if got != want {
			t.Errorf("frame %d: marker = 0x%02X, want 0x%02X", i, got, want)
		}
	}

	// Second Encode call: marker should continue alternating.
	src2 := [][]byte{{0xFF, 0x00}}
	dst2 := make([]byte, dsd.DoPOutputSize(len(src2[0]), 1))

	_, err = enc.Encode(src2, dst2)
	if err != nil {
		t.Fatalf("second Encode: %v", err)
	}

	// After 4 frames (even), next marker should be 0x05.
	if got := dst2[2]; got != 0x05 {
		t.Errorf("streaming: marker = 0x%02X, want 0x05", got)
	}
}

// TestDoPBitReversal verifies that LSB-first input produces
// the same output as MSB-first input with pre-reversed bytes.
func TestDoPBitReversal(t *testing.T) {
	t.Parallel()

	// 0xB4 bit-reversed = 0x2D.
	// 0x69 (DSD silence) bit-reversed = 0x96.
	msbData := [][]byte{{0x2D, 0x96}}
	lsbData := [][]byte{{0xB4, 0x69}}

	msbEnc := dsd.NewDoPEncoder(dsd.MSBFirst)
	lsbEnc := dsd.NewDoPEncoder(dsd.LSBFirst)

	msbOut := make([]byte, dsd.DoPOutputSize(2, 1))
	lsbOut := make([]byte, dsd.DoPOutputSize(2, 1))

	_, err := msbEnc.Encode(msbData, msbOut)
	if err != nil {
		t.Fatalf("MSB Encode: %v", err)
	}

	_, err = lsbEnc.Encode(lsbData, lsbOut)
	if err != nil {
		t.Fatalf("LSB Encode: %v", err)
	}

	for i := range msbOut {
		if msbOut[i] != lsbOut[i] {
			t.Errorf("byte %d: MSB=0x%02X LSB=0x%02X", i, msbOut[i], lsbOut[i])
		}
	}
}

// TestDoPSilence verifies that DSD silence (0x69) encodes correctly.
func TestDoPSilence(t *testing.T) {
	t.Parallel()

	enc := dsd.NewDoPEncoder(dsd.MSBFirst)

	src := [][]byte{{0x69, 0x69, 0x69, 0x69}}
	dst := make([]byte, dsd.DoPOutputSize(len(src[0]), 1))

	n, err := enc.Encode(src, dst)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	if n != len(dst) {
		t.Fatalf("Encode returned %d, want %d", n, len(dst))
	}

	// Each DoP frame should carry 0x69 in both data bytes.
	for frame := range 2 {
		lo := dst[frame*3]
		hi := dst[frame*3+1]

		if lo != 0x69 || hi != 0x69 {
			t.Errorf("frame %d: data = [0x%02X, 0x%02X], want [0x69, 0x69]", frame, hi, lo)
		}
	}
}

// TestDoPRoundTrip verifies that the original DSD bytes can be extracted
// from DoP frames by stripping the marker byte.
func TestDoPRoundTrip(t *testing.T) {
	t.Parallel()

	original := []byte{0xDE, 0xAD, 0xBE, 0xEF, 0xCA, 0xFE}
	enc := dsd.NewDoPEncoder(dsd.MSBFirst)

	src := [][]byte{original}
	dst := make([]byte, dsd.DoPOutputSize(len(original), 1))

	_, err := enc.Encode(src, dst)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	// Extract DSD bytes back: each 3-byte frame has [lo, hi, marker].
	frames := len(original) / 2
	recovered := make([]byte, len(original))

	for i := range frames {
		recovered[i*2] = dst[i*3+1]   // hi byte
		recovered[i*2+1] = dst[i*3+0] // lo byte
	}

	for i := range original {
		if recovered[i] != original[i] {
			t.Errorf("byte %d: recovered 0x%02X, want 0x%02X", i, recovered[i], original[i])
		}
	}
}

// TestDoPInterleaving verifies that stereo DoP output is correctly interleaved.
func TestDoPInterleaving(t *testing.T) {
	t.Parallel()

	enc := dsd.NewDoPEncoder(dsd.MSBFirst)

	left := []byte{0xAA, 0x55}
	right := []byte{0x11, 0x22}
	src := [][]byte{left, right}

	dst := make([]byte, dsd.DoPOutputSize(2, 2))

	n, err := enc.Encode(src, dst)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	// 2 bytes per channel / 2 = 1 frame per channel. 2 channels → 6 bytes total.
	if n != 6 {
		t.Fatalf("Encode returned %d bytes, want 6", n)
	}

	// Frame 0, channel 0 (left): [0x55, 0xAA, marker]
	if dst[0] != 0x55 || dst[1] != 0xAA {
		t.Errorf("left frame: data = [0x%02X, 0x%02X], want [0x55, 0xAA]", dst[0], dst[1])
	}

	// Frame 0, channel 1 (right): [0x22, 0x11, marker]
	if dst[3] != 0x22 || dst[4] != 0x11 {
		t.Errorf("right frame: data = [0x%02X, 0x%02X], want [0x22, 0x11]", dst[3], dst[4])
	}

	// Both channels in the same frame position share the same marker.
	if dst[2] != dst[5] {
		t.Errorf("marker mismatch: left=0x%02X right=0x%02X", dst[2], dst[5])
	}
}

// TestDoPOddLength verifies that an odd number of DSD bytes per channel
// is padded with DSD silence (0x69) in the final DoP frame.
func TestDoPOddLength(t *testing.T) {
	t.Parallel()

	enc := dsd.NewDoPEncoder(dsd.MSBFirst)

	// 3 DSD bytes → 2 frames: [0xAA,0x55] and [0xBB,pad].
	src := [][]byte{{0xAA, 0x55, 0xBB}}
	dst := make([]byte, dsd.DoPOutputSize(3, 1))

	n, err := enc.Encode(src, dst)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	// OutputSize for 3 bytes mono = 2 frames × 3 bytes = 6.
	if n != 6 {
		t.Fatalf("Encode returned %d bytes, want 6", n)
	}

	// Frame 0: [0x55, 0xAA, marker] — normal pair.
	if dst[0] != 0x55 || dst[1] != 0xAA {
		t.Errorf("frame 0: data = [0x%02X, 0x%02X], want [0x55, 0xAA]", dst[0], dst[1])
	}

	// Frame 1: [0x69, 0xBB, marker] — 0xBB is the real byte, 0x69 is DSD silence pad.
	if dst[3] != 0x69 {
		t.Errorf("frame 1 low byte: got 0x%02X, want 0x69 (DSD silence pad)", dst[3])
	}

	if dst[4] != 0xBB {
		t.Errorf("frame 1 high byte: got 0x%02X, want 0xBB", dst[4])
	}
}

// TestDoPOddLengthLSBFirst verifies that the DSD silence pad byte is written
// in MSB-first (DoP native order) even when input is LSB-first.
func TestDoPOddLengthLSBFirst(t *testing.T) {
	t.Parallel()

	enc := dsd.NewDoPEncoder(dsd.LSBFirst)

	// 1 DSD byte → 1 frame: [0xAA, pad].
	src := [][]byte{{0xAA}}
	dst := make([]byte, dsd.DoPOutputSize(1, 1))

	n, err := enc.Encode(src, dst)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	if n != 3 {
		t.Fatalf("Encode returned %d bytes, want 3", n)
	}

	// Pad byte should be 0x69 (DSD silence in MSB-first, no reversal applied).
	if dst[0] != 0x69 {
		t.Errorf("pad byte: got 0x%02X, want 0x69", dst[0])
	}
}

// TestDoPNoChannels verifies that encoding with zero channels returns an error.
func TestDoPNoChannels(t *testing.T) {
	t.Parallel()

	enc := dsd.NewDoPEncoder(dsd.MSBFirst)

	_, err := enc.Encode(nil, nil)
	if err == nil {
		t.Fatal("expected error for nil channels")
	}
}

// TestDoPChannelLengthMismatch verifies that mismatched channel lengths
// produce an error.
func TestDoPChannelLengthMismatch(t *testing.T) {
	t.Parallel()

	enc := dsd.NewDoPEncoder(dsd.MSBFirst)

	src := [][]byte{{0xAA, 0x55}, {0xBB}}
	dst := make([]byte, 100)

	_, err := enc.Encode(src, dst)
	if err == nil {
		t.Fatal("expected error for mismatched channel lengths")
	}
}

// TestDoPDstTooSmall verifies that an undersized dst buffer produces an error.
func TestDoPDstTooSmall(t *testing.T) {
	t.Parallel()

	enc := dsd.NewDoPEncoder(dsd.MSBFirst)

	src := [][]byte{{0xAA, 0x55}}
	dst := make([]byte, 1) // needs 3

	_, err := enc.Encode(src, dst)
	if err == nil {
		t.Fatal("expected error for undersized dst")
	}
}

// TestDoPStreamingContinuity encodes data across many small calls and verifies
// that marker state is continuous throughout.
func TestDoPStreamingContinuity(t *testing.T) {
	t.Parallel()

	enc := dsd.NewDoPEncoder(dsd.MSBFirst)

	// 10 calls of 2 bytes each → 10 frames total.
	chunk := [][]byte{{0x69, 0x69}}
	frame := make([]byte, 3)

	for call := range 10 {
		_, err := enc.Encode(chunk, frame)
		if err != nil {
			t.Fatalf("call %d: %v", call, err)
		}

		var wantMarker byte
		if call%2 == 0 {
			wantMarker = 0x05
		} else {
			wantMarker = 0xFA
		}

		if frame[2] != wantMarker {
			t.Errorf("call %d: marker = 0x%02X, want 0x%02X", call, frame[2], wantMarker)
		}
	}
}

// TestDoPOutputSize verifies the output size calculation.
func TestDoPOutputSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		dsdBytes int
		channels int
		want     int
	}{
		{2, 1, 3},     // 1 frame, mono
		{2, 2, 6},     // 1 frame, stereo
		{4, 1, 6},     // 2 frames, mono
		{4, 2, 12},    // 2 frames, stereo
		{100, 6, 900}, // 50 frames, 6ch
		{0, 1, 0},     // no input
		{1, 1, 3},     // 1 byte odd → 1 frame (padded), mono
		{3, 1, 6},     // 3 bytes odd → 2 frames (last padded), mono
		{3, 2, 12},    // 3 bytes odd → 2 frames (last padded), stereo
		{5, 1, 9},     // 5 bytes odd → 3 frames (last padded), mono
	}

	for _, tt := range tests {
		got := dsd.DoPOutputSize(tt.dsdBytes, tt.channels)
		if got != tt.want {
			t.Errorf("DoPOutputSize(%d, %d) = %d, want %d", tt.dsdBytes, tt.channels, got, tt.want)
		}
	}
}

// TestDoPReset verifies that Reset restores the marker to its initial state.
func TestDoPReset(t *testing.T) {
	t.Parallel()

	enc := dsd.NewDoPEncoder(dsd.MSBFirst)

	// Encode one frame to advance marker to 0xFA.
	src := [][]byte{{0x69, 0x69}}
	dst := make([]byte, 3)

	_, err := enc.Encode(src, dst)
	if err != nil {
		t.Fatal(err)
	}

	if dst[2] != 0x05 {
		t.Fatalf("first frame marker = 0x%02X, want 0x05", dst[2])
	}

	// Reset and encode again — should start at 0x05.
	enc.Reset()

	_, err = enc.Encode(src, dst)
	if err != nil {
		t.Fatal(err)
	}

	if dst[2] != 0x05 {
		t.Errorf("after Reset: marker = 0x%02X, want 0x05", dst[2])
	}
}
