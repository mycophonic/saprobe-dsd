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

// Package dsd provides DSD-to-PCM conversion supporting all standard DSD rates
// (DSD64 through DSD1024) with output to any 44.1 kHz-family rate
// (44.1k, 88.2k, 176.4k, 352.8k, 705.6k).
//
// Architecture:
//
//   - Stage 1: Byte-level 8:1 decimation using Gesemann's 96-tap symmetric FIR
//     filter with precomputed lookup tables. Ported from Sebastian Gesemann's
//     dsd2pcm (BSD-2-Clause). A 4:1 bit-level variant handles the DSD64->705.6k
//     edge case.
//
//   - Stage 2: Cascaded halfband 2:1 decimation on float64 data. Parks-McClellan
//     designed FIR filter reused at each stage.
//
// Use [NewPipeline] for multi-rate conversion. The low-level [Converter] type
// provides direct access to the 8:1 byte-level decimation.
//
// The package also includes a [NoiseShaper] for high-quality quantization
// to 16-bit PCM while preserving dynamic range.
package dsd
