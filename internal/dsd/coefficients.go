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

// DefaultNoiseShapeCoeffs contains noise shaping coefficients for 16-bit
// quantization of DSD-to-PCM output. Two second-order sections (SOS) that
// can preserve a dynamic range of 135 dB below 30 kHz.
//
// Layout: b1, b2, a1, a2 per section.
//
//nolint:gochecknoglobals // the default noise-shaping coefficients, read-only
var DefaultNoiseShapeCoeffs = []float32{
	// Section 1
	-1.62666423, 0.79410094, 0.61367127, 0.23311013,
	// Section 2
	-1.44870017, 0.54196219, 0.03373857, 0.70316556,
}
