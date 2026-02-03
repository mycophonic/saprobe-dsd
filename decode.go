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
	"fmt"
	"io"

	"github.com/mycophonic/saprobe-dsd/internal/dff"
	"github.com/mycophonic/saprobe-dsd/internal/dsf"
)

// DecodeDSF parses a DSF container and returns per-channel raw DSD bytes.
//
// Each element of the returned slice contains the DSD data for one channel.
// The bit ordering is LSB-first (as specified by the DSF format).
func DecodeDSF(reader io.ReadSeeker) ([][]byte, DSDFormat, error) {
	channels, dsdFmt, err := dsf.Decode(reader)
	if err != nil {
		return nil, DSDFormat{}, fmt.Errorf("%w: %w", ErrContainer, err)
	}

	return channels, dsdFmt, nil
}

// DecodeDFF parses a DSDIFF (DFF) container and returns per-channel raw DSD bytes.
//
// Each element of the returned slice contains the DSD data for one channel.
// The bit ordering is MSB-first (as specified by the DSDIFF format).
// DST-compressed files are not supported.
func DecodeDFF(reader io.ReadSeeker) ([][]byte, DSDFormat, error) {
	channels, dsdFmt, err := dff.Decode(reader)
	if err != nil {
		return nil, DSDFormat{}, fmt.Errorf("%w: %w", ErrContainer, err)
	}

	return channels, dsdFmt, nil
}
