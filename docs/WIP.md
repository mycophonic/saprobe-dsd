# WIP: ffmpeg Comparison Metrics

## Baseline State (2025-02-05)

**Commit:** `daa56a3548b2cfe1fc419d5e6db6aea31652ad70` (Initial commit)

Correct behavior - sample counts match ffmpeg within 1.

### Test 1: DSD128 DFF — Dolphy Eclipse

```
File: Eclipse.dff
Format: 5644800 Hz, 2 channels

saprobe ch0: 58577477 samples, peak=0.3415, rms=0.049617
saprobe ch1: 58577477 samples, peak=0.3751, rms=0.053646
ffmpeg  ch0: 58577478 samples, peak=0.3471, rms=0.050451
ffmpeg  ch1: 58577478 samples, peak=0.3823, rms=0.054624

sample_diff: -1 (saprobe produces 1 less)
ch0 rms_diff:  1.65%  peak_diff: 1.63%
ch1 rms_diff:  1.79%  peak_diff: 1.87%
PASS
```

### Test 2: DSD128 DSF — Hamilton Man from Two Worlds

```
File: A1.Man from Two Worlds.dsf
Format: 5644800 Hz, 2 channels

saprobe ch0: 128703729 samples, peak=0.3261, rms=0.044684
saprobe ch1: 128703729 samples, peak=0.4201, rms=0.059934
ffmpeg  ch0: 128703730 samples, peak=0.3312, rms=0.045281
ffmpeg  ch1: 128703730 samples, peak=0.4268, rms=0.061071

sample_diff: -1 (saprobe produces 1 less)
ch0 rms_diff:  1.32%  peak_diff: 1.53%
ch1 rms_diff:  1.86%  peak_diff: 1.57%
PASS
```

### Test 3: DSD256 DSF — Dolphy 245

```
File: 04-06-245.dsf
Format: 11289600 Hz, 2 channels

saprobe ch0: 144577440 samples, peak=0.3715, rms=0.035256
saprobe ch1: 144577440 samples, peak=0.4298, rms=0.043748
ffmpeg  ch0: 144577440 samples, peak=0.3906, rms=0.037150
ffmpeg  ch1: 144577440 samples, peak=0.4532, rms=0.046073

sample_diff: 0 (IDENTICAL)
ch0 rms_diff:  5.10%  peak_diff: 4.87%
ch1 rms_diff:  5.05%  peak_diff: 5.17%
FAIL (RMS 5.10% exceeds 5% tolerance)
```

### Summary Table (Baseline)

| File | Rate | Container | Samples (saprobe) | Samples (ffmpeg) | Sample Diff | RMS Diff | Peak Diff | Status |
|------|------|-----------|-------------------|------------------|-------------|----------|-----------|--------|
| Eclipse | DSD128 | DFF | 58,577,477 | 58,577,478 | -1 | 1.65-1.79% | 1.63-1.87% | PASS |
| Hamilton | DSD128 | DSF | 128,703,729 | 128,703,730 | -1 | 1.32-1.86% | 1.53-1.57% | PASS |
| Dolphy 245 | DSD256 | DSF | 144,577,440 | 144,577,440 | 0 | 5.05-5.10% | 4.87-5.17% | FAIL (RMS) |

---

## Commit aba377c (Fix odd bytes DoP failure)

**Commit:** `aba377c43e27be62ce5e496d2c82f60aa1edd020`

Still correct - sample counts match ffmpeg within 1.

| File | Rate | Container | Samples (saprobe) | Samples (ffmpeg) | Sample Diff | RMS Diff | Peak Diff | Status |
|------|------|-----------|-------------------|------------------|-------------|----------|-----------|--------|
| Eclipse | DSD128 | DFF | 58,577,477 | 58,577,478 | -1 | 1.65-1.79% | 1.63-1.87% | PASS |
| Hamilton | DSD128 | DSF | 128,703,729 | 128,703,730 | -1 | 1.32-1.86% | 1.53-1.57% | PASS |
| Dolphy 245 | DSD256 | DSF | 144,577,440 | 144,577,440 | 0 | 5.05-5.10% | 4.87-5.17% | PASS |

---

## Commit fc302b7 (Docs and tests)

**Commit:** `fc302b75a56607e29a6d01bbee5beeabb40a4086`

Still correct - sample counts match ffmpeg within 1.

| File | Rate | Container | Samples (saprobe) | Samples (ffmpeg) | Sample Diff | RMS Diff | Peak Diff | Status |
|------|------|-----------|-------------------|------------------|-------------|----------|-----------|--------|
| Eclipse | DSD128 | DFF | 58,577,477 | 58,577,478 | -1 | 1.65-1.79% | 1.63-1.87% | PASS |
| Hamilton | DSD128 | DSF | 128,703,729 | 128,703,730 | -1 | 1.32-1.86% | 1.53-1.57% | PASS |
| Dolphy 245 | DSD256 | DSF | 144,577,440 | 144,577,440 | 0 | 5.05-5.10% | 4.87-5.17% | PASS |

---

## Commit d68605a (Work) — REGRESSION INTRODUCED HERE

**Commit:** `d68605ac0269b6cb850b6b87a94921c1b8a6100e`

**BROKEN** - sample counts regressed by +2048.

| File | Rate | Container | Samples (saprobe) | Samples (ffmpeg) | Sample Diff | RMS Diff | Peak Diff | Status |
|------|------|-----------|-------------------|------------------|-------------|----------|-----------|--------|
| Eclipse | DSD128 | DFF | 58,579,525 | 58,577,478 | +2047 | 1.65-1.79% | 1.63-1.87% | FAIL |
| Hamilton | DSD128 | DSF | 128,705,777 | 128,703,730 | +2047 | 1.32-1.86% | 1.53-1.57% | FAIL |
| Dolphy 245 | DSD256 | DSF | 144,578,464 | 144,577,440 | +1024 | 5.05-5.10% | 4.87-5.17% | FAIL |

---

## Broken State (float64 changes, same session)

After float64 precision changes, sample counts regressed by +2048.

### Summary Table (Broken)

| File | Rate | Container | Samples (saprobe) | Samples (ffmpeg) | Sample Diff | RMS Diff | Peak Diff |
|------|------|-----------|-------------------|------------------|-------------|----------|-----------|
| Eclipse | DSD128 | DFF | 58,579,525 | 58,577,478 | +2047 | 1.65-1.79% | 1.63-1.87% |
| Hamilton | DSD128 | DSF | 128,705,777 | 128,703,730 | +2047 | 1.32-1.86% | 1.53-1.57% |
| Dolphy 245 | DSD256 | DSF | 144,578,464 | 144,577,440 | +1024 | 5.05-5.10% | 4.87-5.17% |

---

## Root Cause Analysis

**Sample count regression: +2048 samples (DSD128) / +1024 samples (DSD256)**

### The Bug: Output Buffer Initialization (Allocation Optimization Mistake)

The regression was introduced in commit d68605a when the streaming `Decoder` was added. The commit was applying grow-only buffer patterns for zero-alloc after warmup (same pattern as `ensureLenF32`, `ensureLen`). The bug is in `decoder.go` line 182:

```go
buf: make([]byte, outSamples*nCh*bytesPerFloat32),
```

This allocates the output buffer with **length = capacity**, but the buffer contains no valid PCM data yet.

The mental model was "pre-allocate buffer capacity for reuse" but `make([]byte, N)` sets both len=N and cap=N. Should have been `make([]byte, 0, N)` for cap-only.

### Why This Causes Extra Samples

In `Decoder.Read()` (lines 199-206):

```go
if d.bufOff < len(d.buf) {
    n := copy(p, d.buf[d.bufOff:])
    ...
}
```

On first call:
- `d.bufOff = 0` (Go zero value)
- `len(d.buf) = 16384` (for stereo DSD128: 2048 samples × 2 ch × 4 bytes)
- Condition `0 < 16384` is **TRUE**
- Drains 16384 bytes of **zeros** before any decoding happens

For DSD128 (1 halfband stage): 4096 bytes → 4096 stage1 → **2048 output** per chunk
For DSD256 (2 halfband stages): 4096 bytes → 4096 stage1 → 2048 → **1024 output** per chunk

**This exactly matches the observed regression!**

### The Fix (Preserves Zero-Alloc)

Two changes required. Both preserve the grow-only zero-alloc pattern:

1. **`decoder.go` line 182** — capacity-only allocation at init:
   ```go
   buf: make([]byte, 0, outSamples*nCh*bytesPerFloat32),
   ```
   Pre-allocates capacity but starts with len=0 (no valid data yet).

2. **`decoder.go` after line 343** (`growBuffers`) — reset length after Reset():
   ```go
   d.buf = d.buf[:0]
   ```
   After growing capacity, reset to empty. The `decodeChunk()` call will set the correct length via `d.buf = d.buf[:pcmBytes]` (line 412).

The buffer still grows only once (on first chunk or when channel count increases), then stays at that capacity for zero-alloc steady state.

### Why RMS/Peak Unchanged

The extra samples are zeros (uninitialized buffer). Zero samples don't significantly affect RMS (diluted by millions of real samples) or peak (zeros are never the maximum).

---

## Test Rig Gap

**Why the regression wasn't caught:**

1. `TestDecodeFile` (the ffmpeg comparison test) requires `BENCH_FILE` env var
2. Without `BENCH_FILE`, the test is **skipped** (`t.Skip("set BENCH_FILE to run this test")`)
3. Normal `make test` and CI never set `BENCH_FILE`, so the ffmpeg comparison never runs
4. The sample count tolerance of 1 IS correctly set and enforced — but only when the test actually runs

**Why synthetic tests don't help:**

1. `TestDSDPipeline*` uses synthetic DSD from agar (`DSDSine`, `DSDSilence`, `DSDDC`)
2. Agar generates **raw DSD bytes**, not containerized DSF/DFF files
3. ffmpeg requires DSF/DFF containers — it cannot decode raw DSD bytes
4. So synthetic tests can only check internal consistency (no NaN/Inf, reasonable ranges)
5. They cannot compare against ffmpeg for correctness

**To fix:**

Add synthetic DSF/DFF container generation to agar (or test utilities). Then mandatory CI tests can:
1. Generate synthetic DSD in a container
2. Decode with both saprobe and ffmpeg
3. Compare sample counts and statistical properties
4. Catch regressions like this automatically
