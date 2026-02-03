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

// Package dff parses the DSDIFF (DSD Interchange File Format) container,
// extracting per-channel raw DSD byte data. DSDIFF files use big-endian
// IFF-style chunk headers and MSB-first bit ordering within each octet.
//
// DST-compressed DSDIFF files are not supported; [Decode] returns
// [ErrDSTCompressed] for these.
package dff
