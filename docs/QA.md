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

## ffmpeg Comparison

`TestDecodeFile` cross-compares saprobe output against ffmpeg using statistical properties (RMS energy, peak level) rather than per-sample matching. DSD-to-PCM conversion has no standard algorithm — different implementations use different FIR filters, decimation strategies, and group delays.

**Tolerances:**
- **RMS/Peak difference:** ≤10% (relative to ffmpeg)
- **Sample count difference:** ≤4096 samples

**Measured Results (2025-02, float64 precision):**

| File | Rate | Container | Sample Diff | RMS Diff | Peak Diff |
|------|------|-----------|-------------|----------|-----------|
| Dolphy — Eclipse | DSD128 | DFF | +2047 | 1.65-1.79% | 1.63-1.87% |
| Hamilton — Man from Two Worlds | DSD128 | DSF | +2047 | 1.32-1.86% | 1.53-1.57% |
| Dolphy — 245 | DSD256 | DSF | +1024 | 5.05-5.10% | 4.87-5.17% |

**Observations:**
- DSD128: RMS diff 1.3-1.9%, peak diff 1.5-1.9% — well within tolerance
- DSD256: RMS diff ~5.1%, peak diff ~5.0% — within tolerance but higher due to additional halfband stages
- Sample count: saprobe produces ~2048 more samples (DSD128) or ~1024 more (DSD256) due to different group delay handling at stream boundaries

Higher DSD rates require more halfband decimation stages, causing greater divergence between saprobe's FIR+halfband cascade and ffmpeg's resampler.

## Benchmarks

### Synthetic Decode (`TestBenchmarkDecode`)

Benchmarks the DSD-to-PCM pipeline with synthetic data across DSD64, DSD128, and DSD256 rates. Skipped in `-short` mode.

```bash
cd tests && go test -run TestBenchmarkDecode -count=1 -v
```

**Methodology:**
- **Rates:** DSD64 (2.8 MHz), DSD128 (5.6 MHz), DSD256 (11.2 MHz)
- **Duration:** 1 second per rate (stereo)
- **Container:** Synthetic DSF with DSD silence (0x69)
- **Output:** 352.8 kHz PCM
- **Iterations:** 10 per configuration
- **Metric:** median decode time

**Baseline (2025-02, M1 Max):**

| Rate | Container | Median | Realtime |
|------|-----------|--------|----------|
| DSD64 | 0.7 MB | 8ms | 125x |
| DSD128 | 1.4 MB | 62ms | 16x |
| DSD256 | 2.7 MB | 171ms | 5.8x |

Higher rates are progressively slower due to the multi-stage halfband decimation chain (DSD128 adds one stage, DSD256 adds two). DSD256 at 5.8x realtime is adequate for streaming playback but leaves less headroom than lower rates.

### Profiling (`TestProfileDecode`)

CPU and allocation profiling of the DSD-to-PCM pipeline using synthetic DSD128 stereo data (10 seconds). Uses longer duration than benchmark to collect enough pprof samples.

```bash
./hack/bench.sh TestProfileDecode
```

Generates `docs/decode_cpu.png` and `docs/decode_alloc.png`.

**CPU profile breakdown:**

| Function | Flat | Cumulative |
|----------|------|------------|
| Halfband.Process | 72.15% | 76.96% |
| Converter.Translate | 14.66% | 15.32% |
| asyncPreempt (runtime) | 5.66% | 5.66% |
| Pipeline.Process | 1.60% | 94.02% |

The halfband filter dominates CPU time — this is expected since it performs the heavy lifting of decimation across multiple stages. The converter (8:1 DSD-to-PCM translation with noise shaping) accounts for ~15%.

**Memory profile:** Test data generation dominates allocations (~79%). Runtime allocations during decoding are minimal — the pipeline reuses buffers and produces zero allocations per chunk after warmup (verified by `BenchmarkPipelineAllocs`).

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
