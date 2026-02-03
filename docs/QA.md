# Testing

## DSD pipeline (`TestDSDPipeline`)

Tests the DSD-to-PCM conversion pipeline across all DSD rate / output rate combinations (DSD64-DSD1024 x 5 output rates = 25 combinations).

Sub-tests verify:
- **SampleCount** — correct output sample count, no NaN/Inf, reasonable peak levels
- **Silence** — silence input produces near-zero output (within quantization noise)
- **DC** — positive DC input produces positive mean output
- **SineEnergy** — 1 kHz sine produces measurable RMS energy
- **8to1Equivalence** — Pipeline matches direct Converter for the 8:1 DSD64→352.8k path

```bash
cd tests && go test -run TestDSDPipeline -count=1 -v
```

Test signals are generated synthetically by [agar](https://github.com/mycophonic/agar) (second-order sigma-delta modulator producing DSD sine, silence, DC).

## DoP encoder (`TestDoP`)

Tests the DoP (DSD over PCM) encoder covering rate mapping, marker alternation, bit reversal, streaming continuity, and error cases.

```bash
cd tests && go test -run TestDoP -count=1 -v
```

## Real files (`TestDecodeFile`)

Container decode and DSD-to-PCM pipeline test using real DSF/DFF files. Auto-detects container format by extension. Decodes the container, validates format metadata, then runs each channel through the pipeline at 352.8 kHz.

```bash
BENCH_FILE='/path/to/file.dsf' go test ./tests/ -run TestDecodeFile -count=1 -v
```

Verifies:
- Container parses without error
- Format metadata: valid DSD rate, channel count > 0, samples per channel > 0
- Channel data: all channels equal length, non-empty
- Pipeline output: correct sample count, no NaN/Inf, peak within reasonable range
- Real audio energy: RMS above minimum threshold

Reference files selected for variety in container format, DSD rate, and source material.

| File | Container | DSD Rate | Duration | Source | Character |
|------|-----------|----------|----------|--------|-----------|
| Eric Dolphy — 245 | DSF | DSD256 | 6:55 | 2017 Analogue Productions vinyl | Hard bop quintet |
| Eric Dolphy — Far Cry | DSF | DSD256 | ~7:00 | 2017 Analogue Productions vinyl | Hard bop quintet |
| Eric Dolphy — Outward Bound (6 tracks) | DSF | DSD256 | ~40:00 | 2017 Analogue Productions vinyl | Hard bop quintet |
| Eric Dolphy — Out There (7 tracks) | DFF | DSD128 | ~35:00 | 1982 OJC vinyl, KORG AudioGate | Chamber jazz quartet |
| Tom Waits — Swordfishtrombones (15 tracks) | DSF | DSD128 | ~40:00 | 2009 Island vinyl | Experimental, sparse to dense |
| Chico Hamilton — Man From Two Worlds (7 tracks) | DSF | DSD128 | ~35:00 | R2R DSD transfer (Stig) | Jazz quintet, reel-to-reel source |

Full paths:

```
/Volumes/Anisotope/Music/Processor/Dolphy/0-done/Dolphy, Eric/1960-Outward Bound/[2017-US-12_ Vinyl-Analogue Productions-APRJ 8236-753088823633]/04-06-245.dsf
/Volumes/Anisotope/Music/Processor/Dolphy/1-miss_some/Dolphy, Eric with Little, Booker/1962-Far Cry/[2017-US-12_ Vinyl-Analogue Productions; New Jazz-APRJ 8270; NJ 8270; NJLP 8270-]/03-07-Far Cry.dsf
/Volumes/Anisotope/Music/Processor/Dolphy/1-miss_some/Dolphy, Eric/1961-Out There/[1982-US-12_ Vinyl-New Jazz; Prestige; Original Jazz Classics-NJ-8252; OJC-023-]/04-07-Eclipse.dff
/Volumes/Anisotope/Music/Processor/Other Music/Waits, Tom/1983-Swordfishtrombones/[2009-XE-12_ Vinyl-Island-0042284246910-042284246910]/04-15-Johnsburg, Illinois.dsf
/Volumes/Anisotope/Music/Processor/Incoming/[NOT IN MB]/Chico Hamilton - Man From Two Worlds_R2R_DSD_Stig/A1.Man from Two Worlds.dsf
```

## Benchmarks

### Synthetic Decode (`TestBenchmarkDecode`)

Benchmarks the DSD-to-PCM pipeline with synthetic data across DSD64, DSD128, and DSD256 rates. Skipped in `-short` mode.

```bash
cd tests && go test -run TestBenchmarkDecode -count=1 -v
```

- **Rates:** DSD64 (2.8 MHz), DSD128 (5.6 MHz), DSD256 (11.2 MHz)
- **Duration:** 10 seconds per rate (stereo)
- **Decoder:** saprobe (pipeline only, no container overhead)
- **Iterations:** 10 per configuration
- **Statistics:** median, mean, stddev, min, max

| Rate | DSD Data | saprobe | Realtime Factor |
|------|----------|---------|-----------------|
| DSD64 | 6.7 MB | 61ms | 164x |
| DSD128 | 13.5 MB | 605ms | 16.5x |
| DSD256 | 26.9 MB | 1.69s | 5.9x |

DSD64 is comfortably fast. DSD128 and DSD256 are progressively slower due to the multi-stage halfband decimation chain (DSD128 adds one stage, DSD256 adds two).

### Real Files (`TestBenchmarkDecodeFile`)

Benchmarks full DSD-to-PCM decoding (container decode + pipeline) against ffmpeg. 10 iterations, median times. Skipped in `-short` mode.

```bash
BENCH_FILE='/path/to/file.dsf' go test ./tests/ -timeout 20m -run TestBenchmarkDecodeFile -count=1 -v
```

Reference files selected for variety in container format and duration. All files are DSD128 (5,644,800 Hz), stereo.
Decode times are median over 10 iterations.
Ratio column shows saprobe time relative to ffmpeg (>1x = saprobe slower).

| File | Container | Duration | DSD Data | saprobe | ffmpeg | vs ffmpeg | Character |
|------|-----------|----------|----------|---------|--------|-----------|-----------|
| Dolphy — Eclipse | DFF | 2:46 | 223.5 MB | 10.622s | 1.828s | 5.8x | Chamber jazz quartet, 1982 OJC vinyl |
| Dolphy — The Baron | DFF | 2:57 | 238.8 MB | 11.24s | 1.956s | 5.7x | Chamber jazz quartet, 1982 OJC vinyl |
| Hamilton — Man from Two Worlds | DSF | 6:04 | 491.0 MB | 23.036s | 4.766s | 4.8x | Jazz quintet, reel-to-reel DSD |
| Dolphy — Out There | DFF | 6:57 | 561.7 MB | 26.743s | 4.951s | 5.4x | Chamber jazz quartet, 1982 OJC vinyl |

Saprobe is ~4.8-5.8x slower than ffmpeg. ffmpeg uses SIMD-optimized C for DSD-to-PCM conversion. The ratio is consistent across both container formats (DSF/DFF) and file durations.

Full paths:

```
/Volumes/Anisotope/Music/Processor/Dolphy/1-miss_some/Dolphy, Eric/1961-Out There/[1982-US-12_ Vinyl-New Jazz; Prestige; Original Jazz Classics-NJ-8252; OJC-023-]/04-07-Eclipse.dff
/Volumes/Anisotope/Music/Processor/Dolphy/1-miss_some/Dolphy, Eric/1961-Out There/[1982-US-12_ Vinyl-New Jazz; Prestige; Original Jazz Classics-NJ-8252; OJC-023-]/03-07-The Baron.dff
/Volumes/Anisotope/Music/Processor/Incoming/[NOT IN MB]/Chico Hamilton - Man From Two Worlds_R2R_DSD_Stig/A1.Man from Two Worlds.dsf
/Volumes/Anisotope/Music/Processor/Dolphy/1-miss_some/Dolphy, Eric/1961-Out There/[1982-US-12_ Vinyl-New Jazz; Prestige; Original Jazz Classics-NJ-8252; OJC-023-]/01-07-Out There.dff
```
