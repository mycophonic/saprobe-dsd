# saprobe-dsd

Pure Go DSD (Direct Stream Digital) decoder supporting DSF and DSDIFF (DFF) container formats
with DSD-to-PCM conversion at all standard rate combinations.

## Containers

| Format | Extension | Bit order | Status |
|--------|-----------|-----------|--------|
| DSF    | `.dsf`    | LSB-first | Supported |
| DSDIFF | `.dff`    | MSB-first | Supported (uncompressed only) |

DST-compressed DSDIFF files are not supported.

## Rate matrix

All 25 combinations of DSD input rate and PCM output rate are supported:

| Input \ Output | 44.1 kHz | 88.2 kHz | 176.4 kHz | 352.8 kHz | 705.6 kHz |
|----------------|----------|----------|-----------|-----------|-----------|
| DSD64 (2.8 MHz)   | 64:1 | 32:1 | 16:1 | 8:1 | 4:1 |
| DSD128 (5.6 MHz)  | 128:1 | 64:1 | 32:1 | 16:1 | 8:1 |
| DSD256 (11.3 MHz) | 256:1 | 128:1 | 64:1 | 32:1 | 16:1 |
| DSD512 (22.6 MHz) | 512:1 | 256:1 | 128:1 | 64:1 | 32:1 |
| DSD1024 (45.2 MHz)| 1024:1 | 512:1 | 256:1 | 128:1 | 64:1 |

## Architecture

- **Stage 1:** Byte-level 8:1 decimation using Gesemann's 96-tap symmetric FIR filter
  with precomputed lookup tables (ported from dsd2pcm, BSD-2-Clause).
  A 4:1 bit-level variant handles the DSD64→705.6 kHz edge case.
- **Stage 2:** Cascaded halfband 2:1 decimation (Parks-McClellan FIR).
- **Noise shaping:** Optional second-order IIR for 16-bit quantization (135 dB dynamic range below 30 kHz).

## Usage

```go
import dsd "github.com/mycophonic/saprobe-dsd"

// Streaming DSD-to-PCM decode (DSF or DFF auto-detected).
dec, err := dsd.NewDecoder(reader, dsd.Rate352800)
pcm, err := io.ReadAll(dec)
fmt := dec.Format() // PCMFormat{SampleRate, BitDepth, Channels}

// Reuse decoder for multiple files (zero-alloc when format matches).
err = dec.Reset(nextReader, dsd.Rate352800)
```

## Dependencies

Zero. Pure Go, no CGO, no third-party dependencies.

## Documentation

- [Testing](docs/QA.md)
- [DAC format support research](docs/RESEARCH.md)
