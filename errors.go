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

import "errors"

// Public sentinel errors for consumer error matching.
var (
	// ErrContainer indicates an invalid or unrecognized DSD container
	// (not a valid DSF or DFF file, missing required chunks).
	ErrContainer = errors.New("container error")

	// ErrUnsupportedRate indicates an invalid DSD/output rate combination.
	ErrUnsupportedRate = errors.New("unsupported rate combination")

	// ErrDecode indicates a failure during DSD-to-PCM conversion.
	ErrDecode = errors.New("decode failed")

	// ErrDoP indicates a failure in DoP (DSD over PCM) encoding.
	ErrDoP = errors.New("dop error")
)
