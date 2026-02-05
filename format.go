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
	intdsd "github.com/mycophonic/saprobe-dsd/internal/dsd"
	"github.com/mycophonic/saprobe-dsd/internal/format"
)

// BitOrder represents the bit ordering within each DSD octet.
type BitOrder = format.BitOrder

const (
	// MSBFirst means the most-significant bit is the earliest sample (DFF/DSDIFF convention).
	MSBFirst BitOrder = format.MSBFirst
	// LSBFirst means the least-significant bit is the earliest sample (DSF convention).
	LSBFirst BitOrder = format.LSBFirst
)

// PCMFormat describes the format of decoded PCM audio output.
type PCMFormat struct {
	SampleRate int
	BitDepth   int
	Channels   int
}

// InputRate represents a DSD sample rate in Hz.
type InputRate = intdsd.InputRate

// Supported DSD rates.
const (
	DSD64   InputRate = intdsd.DSD64
	DSD128  InputRate = intdsd.DSD128
	DSD256  InputRate = intdsd.DSD256
	DSD512  InputRate = intdsd.DSD512
	DSD1024 InputRate = intdsd.DSD1024
	DSD2048 InputRate = intdsd.DSD2048
)

// OutputRate represents a PCM output sample rate in Hz.
type OutputRate = intdsd.OutputRate

// Supported output rates (44.1 kHz family only).
const (
	Rate44100  OutputRate = intdsd.Rate44100
	Rate88200  OutputRate = intdsd.Rate88200
	Rate176400 OutputRate = intdsd.Rate176400
	Rate352800 OutputRate = intdsd.Rate352800
	Rate705600 OutputRate = intdsd.Rate705600
)
