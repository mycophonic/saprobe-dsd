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

// Package format defines shared DSD format types used by both the public API
// and internal container parsers.
package format

// BitOrder represents the bit ordering within each DSD octet.
type BitOrder uint8

const (
	// MSBFirst means the most-significant bit is the earliest sample (DFF/DSDIFF convention).
	MSBFirst BitOrder = iota
	// LSBFirst means the least-significant bit is the earliest sample (DSF convention).
	LSBFirst
)

// DSDFormat describes the format of raw DSD audio data extracted from a container.
type DSDFormat struct {
	// SampleRate is the DSD sample rate in Hz (e.g. 2822400 for DSD64).
	SampleRate int
	// Channels is the number of audio channels.
	Channels uint
	// SamplesPerChannel is the total number of DSD samples (bits) per channel.
	SamplesPerChannel int64
	// BitOrder indicates MSB-first (DFF) or LSB-first (DSF) bit ordering.
	BitOrder BitOrder
}
