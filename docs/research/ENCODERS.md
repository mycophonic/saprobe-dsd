# DSD Encoders: Prevalence and Technical Specifications

## Overview

DSD (Direct Stream Digital) is **not a codec in the traditional sense**. It is a raw 1-bit delta-sigma modulated (PDM) bitstream, developed by Sony and Philips for SACD. Unlike PCM codecs (FLAC, ALAC, MP3) where "encoding" means compressing audio, DSD "encoding" means one of two fundamentally different things:

1. **PCM → DSD modulation**: Converting multi-bit PCM audio into a 1-bit delta-sigma bitstream via sigma-delta modulation. This is a **lossy** operation — the modulator must shape quantization noise out of the audible band, and the quality depends entirely on the modulator's order, topology, and noise-shaping filters.

2. **Lossless DSD compression**: Compressing an existing DSD bitstream to reduce file size without altering any bits. Only two options exist: DST (Philips) and WavPack 5.x.

DSD comes in several rates, all multiples of the base 2.8224 MHz (64× the CD sample rate of 44.1 kHz):

| Name | Sample Rate | Multiplier | Bitrate (stereo) | ~Equivalent PCM |
|------|-------------|------------|------------------|-----------------|
| DSD64 | 2.8224 MHz | 64× | 5.6 Mbps | ~20-bit/96 kHz |
| DSD128 | 5.6448 MHz | 128× | 11.3 Mbps | ~24-bit/176 kHz |
| DSD256 | 11.2896 MHz | 256× | 22.6 Mbps | ~24-bit/352 kHz |
| DSD512 | 22.5792 MHz | 512× | 45.2 Mbps | theoretical |
| DSD1024 | 45.1584 MHz | 1024× | 90.3 Mbps | theoretical |

**Container formats**: DSF (Sony, supports ID3v2 tags — use this one), DFF/DSDIFF (Philips, no metadata — avoid unless forced), SACD ISO (disc image), WavPack .wv (lossless compressed).

DSD is patented/trademarked by Sony. The DSDIFF spec is by Philips (freely available). The DSF spec is by Sony (freely available). There are no royalties for creating DSF/DFF files.

---

## Part 1: PCM → DSD Modulators

This is where DSD "encoding" actually happens. The sigma-delta modulator converts multi-bit PCM into a 1-bit bitstream. Quality varies **enormously** between implementations — unlike lossless codecs where the output is deterministic, different modulators produce audibly different results because the noise shaping is different.

### Modulator Comparison Table

| Rank | Encoder | Max DSD Rate | Modulator Options | Platform | Price | Notes |
|------|---------|-------------|-------------------|----------|-------|-------|
| 1 | **HQPlayer (Signalyst)** | DSD1024+ (98.3 MHz) | 36 modulators | Win/Mac/Linux | $340 | Undisputed king. Absurd depth of options. CUDA GPU acceleration for filters. |
| 2 | **Weiss Saracon-DSD** | DSD128 | Selectable (proprietary) | Win/Mac/Linux | ~$2,000+ | Professional mastering standard. Limited to DSD128 max. |
| 3 | **Korg AudioGate 4** | DSD256 (11.2 MHz) | Proprietary (Korg) | Win/Mac | Free* | *Full version requires Korg DS-DAC or MR hardware for activation. Lite version converts DSD to 44.1/48 kHz PCM only. |
| 4 | **JRiver Media Center** | DSD128 | Built-in (limited) | Win/Mac/Linux | $70 | Real-time PCM→DSD upsampling during playback. Basic modulator. |
| 5 | **AuI ConverteR** | DSD1024 | Proprietary | Win/Mac | $50–$250 | Batch converter. Free version mutes 2 seconds in middle of output. |
| 6 | **PCM-DSD_Converter** | DSD2048 | 8-stage elliptic noise shaper | Windows | Free | Open source (FFTW3). Unstable above DSD512. Single-developer Japanese project. |
| 7 | **TASCAM Hi-Res Editor** | DSD256 (11.2 MHz) | Proprietary | Win/Mac | Free | Basic editing and conversion. Stereo only. Originally bundled with DA-3000. |
| 8 | **dBpoweramp** | DSD128 | Built-in | Win/Mac | $48 | Batch converter with DSD support. Mature, well-maintained. |

---

### Detailed Notes

### HQPlayer (Signalyst) — Rank 1

- **Developer**: Jussi Laako (Finland). Solo developer, 25+ years of development.
- **What it is**: An audiophile player + real-time DSP engine + PCM→DSD modulator. Not just a file converter — it processes audio on-the-fly during playback.
- **Modulators**: 36 delta-sigma modulators available, ranging from lightweight 5th-order IIR types to brutally CPU-intensive 7th-order adaptive types (ASDM7ECv2). Each has different noise-shaping characteristics.
- **Filters**: 71 upsampling algorithms, 78 PCM oversampling options, 3 DSD rate converters, 10 dithering/noise-shaping options. The combinatorial space is enormous.
- **CUDA acceleration**: Filter computation can be offloaded to NVIDIA GPUs (Compute Capability 7.5+). Some filter+modulator combos at DSD1024 require both a top-tier CPU (i9-class single-core) AND a workstation GPU.
- **The "holy grail"**: sinc-L filter + ASDMECv2 modulator @ DSD1024. Requires absurd hardware. Most users settle for DSD256 or DSD512 with less demanding combos.
- **Volume control**: One of the few tools that can do software volume on DSD (by operating in the multi-bit domain before final modulation).
- **Integration**: Works with Roon as an endpoint. Native DSD output via ASIO, DoP, or ALSA DSD.
- **Bottom line**: If you're serious about PCM→DSD conversion, this is it. Nothing else comes close in terms of modulator quality and options.

### Weiss Saracon-DSD — Rank 2

- **Developer**: Weiss Engineering (Switzerland). Daniel Weiss — decades of professional digital audio experience.
- **What it is**: A professional-grade batch sample rate converter with DSD support. Not a player.
- **DSD limitation**: DSD64 and DSD128 only. No DSD256+. This is a deliberate choice — Weiss targets mastering houses, and DSD128 is the practical ceiling for SACD mastering.
- **Modulators**: Multiple selectable, each with gain, dithering, and filter options. Proprietary algorithms.
- **Filter controversy**: Saracon applies a very steep filter starting around 22 kHz when converting DSD→PCM, killing everything above ~40 kHz. This means DSD→PCM at 88.2 kHz and 176.4 kHz produce identical frequency content. Professional mastering engineers (notably Bruce Brown of Puget Sound Studios) have publicly criticized this. Pyramix reportedly produces better DSD→PCM conversions.
- **Channels**: Unlimited for DSD→PCM. Limited to 6 channels for PCM→DSD.
- **File formats**: wav, aiff, au, bwf, sd2, caf, paf, mat, raw, dsf, dsdiff, FLAC, Ogg/Vorbis.
- **THD+N**: PCM→DSD better than -147 dB (32-bit float input, unweighted). PCM→PCM better than -180 dB.
- **Bottom line**: The professional mastering tool. Expensive, limited in DSD rate, but trusted by studios worldwide.

### Korg AudioGate 4 — Rank 3

- **Developer**: Korg (Japan). Company with deep DSD heritage from their MR-series recorders.
- **What it is**: Player, editor, and format converter. Originally bundled with Korg hardware.
- **DSD editing**: Can perform basic edits (split, join, fade, volume, normalize) entirely in the DSD domain without converting to PCM. This is rare and genuinely useful — most tools require a PCM round-trip.
- **Conversion quality**: Decent. Korg has years of DSD experience from professional recorders. Community consensus places it below HQPlayer and Pyramix but above most others.
- **Activation trap**: Free download, but the Lite version cripples DSD playback to 44.1/48 kHz PCM conversion. Full version requires connecting a Korg DS-DAC or MR-series device for license authentication. You can find used DS-DAC units on eBay for ~$100, making this the cheapest path to a decent DSD converter.
- **DSD disc creation**: Can burn DSD discs playable on Sony SCD-XA5400ES and PlayStation 3.
- **Recording**: Supports DSD recording when paired with DS-DAC-10R (stereo) or Nu I (8-channel, up to 11.2 MHz).

### PCM-DSD_Converter — Rank 6

- **Developer**: sunacchi/serieril (Japan). Solo developer.
- **What it is**: Bare-bones PCM→DSD converter. Open source (uses FFTW3, GPL).
- **Approach**: Upsamples PCM to target DSD rate in 2× stages, then applies 8-stage elliptic noise shaping.
- **Two modes**: High precision (FIR anti-aliasing filter) and Low precision (IIR). Custom filter coefficients can be loaded from files.
- **Limitations**: Windows only. Mono/stereo only. Unstable above DSD512. No metadata handling.
- **Historical significance**: Based on Koon's pioneering "Wav2DSFconverter" (2011), one of the first open-source PCM→DSD tools.

---

## Part 2: Lossless DSD Compression

DSD files are enormous. A stereo DSD64 album runs ~2–3 GB. DSD256 is 4× that. Only two lossless compression formats exist for DSD.

### Compression Comparison Table

| Rank | Compressor | Compression Ratio | DSD Rates | Containers | Platform | License | Notes |
|------|-----------|------------------|-----------|------------|----------|---------|-------|
| 1 | **WavPack 5.x** | ~40–60% | DSD64–DSD256+ | .wv | Cross-platform | BSD-3-Clause | The only practical end-user option. Free, open source. |
| 2 | **DST (Direct Stream Transfer)** | ~50–65% | DSD64 (SACD) | .dff (DSDIFF) | N/A (pro tools) | Philips proprietary | Used on SACDs. Not practically available to end users. |

### WavPack 5.x — The Only Game in Town

- **Developer**: David Bryant. Solo developer since 1998. DSD support added in version 5.0 (December 2016).
- **How it works**: Treats DSD as 8-bit samples (8 consecutive DSD bits packed into a byte, MSB-first temporally). Uses WavPack's lossless compression on these byte streams. No floating-point math — integer-only data path to guarantee bit-perfect round-trips.
- **Compression**: Typically 40–60% reduction depending on source material and mode. Two DSD compression modes available (default is faster with less compression; high mode is slower with better compression).
- **Input**: DSF and DFF files. Rejects DST-compressed DFF (already compressed, and usually better than WavPack would achieve).
- **Output**: .wv files. Can be decompressed back to DSF or DFF with `wvunpack`. Can also convert between DSF↔DFF during decompression.
- **Playback support**: foobar2000 (with SACD plugin 1.04+), JRiver, MPD, Roon, dBpoweramp. FFmpeg can decode WavPack but may not handle DSD WavPack correctly in all cases.
- **DSD-as-PCM mode**: `OPEN_DSD_AS_PCM` flag decimates DSD 8× and presents it as 24-bit PCM at 352.8 kHz (for DSD64). Still very high sample rate with lots of quantization noise — needs further downsampling.
- **No hybrid mode for DSD**: WavPack's unique hybrid lossy+correction mode only works for PCM. There is no lossy algorithm for DSD.
- **CRC verification**: 32-bit error detection codes included. DSF and DFF containers do NOT provide CRC verification — WavPack is the only DSD format that does.
- **Bottom line**: If you have a DSD library, compress it with WavPack. Period. It's the only option, and it's excellent — free, open source, well-maintained, ~50% size reduction with bit-perfect lossless restoration.

### DST (Direct Stream Transfer) — Historical/SACD Only

- **Developer**: Philips. Part of the DSDIFF specification.
- **What it is**: Lossless compression of DSD, used on commercial SACDs to fit more audio on the disc.
- **Not available to end users**: The DST encoder is professional/proprietary Philips software. You cannot create DST-compressed files yourself without pro mastering tools (e.g., Sonic Studio/Pyramix).
- **Decoding**: sacd_extract can decompress DST→raw DSD during SACD ISO extraction. The foobar2000 SACD plugin can decode DST in real-time. Very CPU-intensive compared to raw DSD playback.
- **Compression**: Generally slightly better than WavPack (~50–65%), but extremely heavy on CPU for both encoding and decoding.

---

## Part 3: SACD ISO Extraction

Not technically "encoding," but an essential part of the DSD workflow: extracting DSD tracks from SACD disc images.

### sacd_extract (sacd-ripper)

- **Original developer**: "Mr Wicked" (anonymous). Originally required a jailbroken PS3 to rip SACDs.
- **Active fork**: EuFlo/sacd-ripper (GitHub). Cross-platform (Windows, Linux, macOS including Apple Silicon). Actively maintained as of 2024.
- **What it does**: Extracts individual tracks from SACD ISO images (or live from a networked Blu-ray player running sacd-ripper daemon) to DSF, DFF, or raw ISO.
- **DST handling**: Can decompress DST to raw DSD during extraction (`-c` flag). Essential for DSF output since DSF doesn't support DST.
- **Output formats**: DSF (recommended — has ID3v2 metadata), DFF/DSDIFF (no metadata), DSDIFF Edit Master, raw ISO.
- **Stereo + multichannel**: Can extract both. SACD discs may contain stereo, multichannel (5.0/5.1), or both. Can also include a CD-quality Redbook layer.
- **CUE sheet export**: Supported.
- **Usage**: `sacd_extract -2 -s -c -i input.iso -o ./output/` (stereo, DSF, decompress DST).

### dsdunpack

- **What it does**: Converts between DSF↔DFF containers, including DST decompression. Based on sacd-ripper's libsacd and libdstdec.
- **Use case**: If you have DFF files and want DSF (for metadata), or vice versa.

---

## Summary: Which Encoder/Tool to Use?

| Use Case | Recommended | Why |
|----------|-------------|-----|
| PCM→DSD (audiophile, real-time) | **HQPlayer** | 36 modulators, GPU acceleration, unmatched quality |
| PCM→DSD (professional mastering) | **Weiss Saracon-DSD** | Studio standard, trusted algorithms, but DSD128 max |
| PCM→DSD (free/budget) | **Korg AudioGate** + cheap used DS-DAC | Decent quality, DSD-domain editing, free with hardware |
| PCM→DSD (batch, free) | **PCM-DSD_Converter** | Free, open source, but Windows-only and basic |
| Compress DSD library | **WavPack 5.x** | The only option. ~50% lossless compression. Free, open source. |
| Extract SACD ISOs | **sacd_extract** (EuFlo fork) | The standard. Cross-platform, actively maintained. |
| Convert DSF↔DFF | **dsdunpack** or **WavPack** (`wvunpack`) | Both work. WavPack round-trip also verifies integrity via CRC. |

### Important Reality Check

PCM→DSD conversion is **fundamentally lossy**. You are running audio through a sigma-delta modulator, which introduces shaped quantization noise. Converting a 24-bit/96 kHz FLAC to DSD64 does not make it "higher quality" — it makes it a DSD64 file with the theoretical dynamic range and frequency response of DSD64 (which is roughly comparable to 20-bit/96 kHz PCM in the audible band). The only DSD that is genuinely "native" is audio that was recorded directly to DSD through an analog-to-DSD converter, or mastered from DXD (352.8 kHz PCM) to DSD by a mastering engineer.

Whether converted DSD sounds "better" through your specific DAC is a function of your DAC's DSD path vs. its PCM path — not an inherent quality improvement. Some DACs do have a cleaner DSD signal path, which is why people convert. But the conversion itself adds nothing that wasn't in the source.