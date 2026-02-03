# DSD Decoders: Prevalence and Technical Specifications

## Overview

"Decoding" DSD means one or more of:

1. **Container parsing**: Reading DSF, DFF (DSDIFF), SACD ISO, or WavPack DSD files and extracting the raw 1-bit bitstream.
2. **DSD → PCM conversion**: Decimating/filtering the 1-bit bitstream into multi-bit PCM for playback on standard audio hardware. This is the most common operation. Quality depends heavily on the decimation filter used.
3. **Native DSD passthrough**: Sending the raw DSD bitstream to a DSD-capable DAC, either natively (ASIO DSD, ALSA DSD) or packed into PCM frames (DoP — DSD over PCM).

Unlike PCM codecs where decoding is deterministic (bit-perfect), DSD → PCM conversion is **lossy and filter-dependent**. Different decoders produce different PCM output from the same DSD source, because they use different decimation filters. There is no single "correct" PCM output from a DSD file.

For native DSD playback (passthrough to DAC), decoding is trivial — just read the container and pass the bits through. The complexity is in the plumbing (ASIO, DoP markers, ALSA DSD support), not the decoding.

---

## Decoder / Player Comparison Table

| Rank | Decoder | DSF | DFF | SACD ISO | WavPack DSD | DSD→PCM | Native DSD Out | Platform | Notes |
|------|---------|-----|-----|----------|-------------|---------|----------------|----------|-------|
| 1 | **FFmpeg (libavcodec)** | ✓ | ✓ | ✗ | ✗ | ✓ | ✗ | Cross-platform | Powers VLC, mpv, Kodi. The workhorse. |
| 2 | **foobar2000 + foo_input_sacd** | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ (ASIO/WASAPI/DoP) | Windows, macOS | Gold standard for DSD playback. |
| 3 | **MPD (Music Player Daemon)** | ✓ | ✓ | ✗ | ✓ | ✓ | ✓ (ALSA DSD/DoP) | Linux, BSD | Server-based. Native DSD via ALSA. |
| 4 | **HQPlayer (Signalyst)** | ✓ | ✓ | ✗ | ✗ | ✓ (14 algorithms) | ✓ (ASIO/DoP/ALSA) | Win/Mac/Linux | Best DSD→PCM filter selection. Also an encoder. |
| 5 | **Korg AudioGate** | ✓ | ✓ | ✗ | ✗ | ✓ | ✓ (ASIO/CoreAudio) | Win/Mac | Free (Lite). Full version needs Korg hardware. |
| 6 | **Audirvāna** | ✓ | ✓ | ✗ | ✗ | ✓ | ✓ (DoP/CoreAudio) | Mac/Win | Commercial player. Popular on macOS. |
| 7 | **Roon** | ✓ | ✓ | ✗ | ✓ | ✓ | ✓ (DoP/ASIO/ALSA) | Win/Mac/Linux | Subscription. Uses its own DSP engine. |
| 8 | **JRiver Media Center** | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ (ASIO/WASAPI/DoP) | Win/Mac/Linux | $70. Full-featured. |
| 9 | **dsd2pcm (Gesemann)** | ✗ (raw) | ✗ (raw) | ✗ | ✗ | ✓ | ✗ | Cross-platform | Foundational library. No container parsing. |
| 10 | **dsf2flac** | ✓ | ✓ | ✗ | ✗ | ✓ (→FLAC) | ✗ | Linux (primarily) | CLI tool. Uses dsd2pcm filters. |

---

## Detailed Notes

### FFmpeg (libavcodec) — Rank 1

- **Implementation**: Native DSD decoder in `dsddec.c`. Four codec IDs: `DSD_LSBF`, `DSD_MSBF`, `DSD_LSBF_PLANAR`, `DSD_MSBF_PLANAR` (least/most significant bit first, interleaved/planar).
- **Container support**: DSF and DFF demuxers in libavformat. No SACD ISO support. No WavPack DSD (FFmpeg has a WavPack decoder, but DSD-in-WavPack handling may not work correctly).
- **DSD→PCM conversion**: Uses a lookup-table-based approach via `ff_dsd2pcm_translate()`. Outputs 32-bit float PCM. The filter is based on Sebastian Gesemann's dsd2pcm library.
- **Default output**: When you `ffmpeg -i file.dsf -af lowpass=24000 -sample_fmt s24 -ar 88200 output.flac`, FFmpeg decodes DSD to high-rate PCM internally, then you downsample with filters. **Critical**: Without a lowpass filter, DSD→PCM conversion produces massive ultrasonic noise (this is inherent to DSD, not a bug). Always apply a lowpass when converting.
- **Bit depth issue**: FFmpeg defaults to 16-bit output for DSD, which is inadequate. Always specify `-sample_fmt s32` or `s24` explicitly. Some players (deadbeef, etc.) have been bitten by this default.
- **Volume offset**: DSD files typically have 6 dB of headroom. Converted PCM will sound quieter than expected. Use `-af volume=6dB` to compensate.
- **Who uses it**: VLC, mpv, Kodi, Plex, Jellyfin — any FFmpeg-based media player. This is how the vast majority of people encounter DSD playback, usually without realizing it. The conversion quality is adequate but not best-in-class.
- **No native DSD output**: FFmpeg always converts to PCM. Cannot pass DSD bitstream through to a DAC.

### foobar2000 + foo_input_sacd — Rank 2

- **Plugin developer**: Maxim V. Anisiutkin (SourceForge: sacddecoder). Active development — version 2.0.12 (October 2025).
- **Format support**: DSF, DFF (including DST-compressed), SACD ISO (including SACD-R/RW discs via DVD drive), WavPack DSD. This is the most comprehensive DSD input support available in any player.
- **DSD output modes**:
    - **PCM**: Converts DSD to PCM using selectable algorithms. Installable custom FIR filters supported.
    - **DSD (native)**: Outputs raw DSD via ASIO 2.1 DSD mode. Requires a DAC with ASIO DSD support.
    - **DSD+PCM**: DSD output with VU meters and DSP still functional.
    - **DoP**: DSD over PCM (0x05/0xFA and 0x06/0xF9 markers). Works with WASAPI, DirectSound, CoreAudio.
- **DSD→PCM filters**: Supports installable FIR filters for direct DSD-to-PCM decimation. Custom filter coefficients can be placed in a `filters` subfolder.
- **DSD processing**: Separate DSD Processor and DSD Converter plugins for extracting tracks from ISOs into DSF/DFF/DSD-encoded files.
- **Volume control**: Disabled in pure DSD mode (inherent limitation — you can't attenuate a 1-bit stream without converting to multi-bit). Works in DSD+PCM and PCM modes.
- **Channel support**: Stereo and multichannel. Downmix to stereo available for multichannel SACDs.
- **WavPack DSD**: Plays back as native DSD when SACD plugin is at top of decoder priority list; otherwise plays as 24/352.8 PCM.
- **Bottom line**: If you're on Windows and serious about DSD, this is the setup. Free, comprehensive, and actively maintained for 13+ years.

### MPD (Music Player Daemon) — Rank 3

- **DSD support**: Native DSF and DFF playback. WavPack DSD via libwavpack.
- **Output modes**:
    - ALSA DSD: Native DSD output on Linux via ALSA's DSD support (requires compatible kernel + DAC driver).
    - DoP: Packs DSD into PCM frames for DACs that support DoP but not native DSD.
    - PCM conversion: Decimates DSD to PCM at configurable sample rate.
- **Key advantage**: Server architecture. Runs headless on a Raspberry Pi, NAS, or dedicated music server. This is how many audiophile streamers (Volumio, Moode Audio, etc.) handle DSD.
- **No SACD ISO**: Must extract to DSF/DFF first using sacd_extract.
- **Gapless DSD**: Supported between tracks of the same DSD rate.

### HQPlayer (Signalyst) — Rank 4 (as decoder)

- **DSD→PCM**: 14 selectable decimation filter algorithms and 12 noise filter options. Far more filter options than any other decoder. If you care about the quality of DSD→PCM conversion specifically, this is the best tool.
- **DSD rate conversion**: 3 options for DSD→DSD rate conversion (e.g., DSD64→DSD256). Can adjust frequency aperture (narrow/wide) depending on source material characteristics.
- **6 dB gain**: Optional, to compensate for DSD's standard headroom.
- **Native DSD output**: ASIO DSD, DoP v1.1, ALSA DSD. Custom Linux kernel packages available for native DSD support.
- **Price**: $340. Overkill for simple playback. Justified if you're using it for its modulator/filter capabilities.

### dsd2pcm (Sebastian Gesemann) — Rank 9

- **Developer**: Sebastian Gesemann (Germany). Created 2009–2011.
- **What it is**: A small C library (not a player or file converter) that converts DSD samples to PCM using a lookup-table-based FIR filter approach. No container parsing — operates on raw DSD byte streams.
- **Historical importance**: This is the foundational DSD→PCM library. Its code and algorithm have been integrated into:
    - FFmpeg's DSD decoder (`dsddec.c` / `dsd.h`)
    - dsf2flac
    - dsd2dxd
    - MPD
    - foo_input_sacd
    - Numerous other projects
- **Filter**: Fixed 256-tap FIR filter. Good quality for a general-purpose filter, but not configurable.
- **License**: BSD-style.
- **Status**: No updates since 2011. The original Google Code repository is archived. But the code lives on in every major DSD decoder.

### dsf2flac — Rank 10

- **Developer**: Originally by Hank (GitHub: hank/dsf2flac). Multiple forks exist (denius, DrTon, julwong, dhalsimax).
- **What it does**: Command-line tool to convert DSF/DFF files to FLAC. Uses dsd2pcm-derived decimation filters.
- **Output**: FLAC at selectable PCM sample rates (must be multiples of 44.1 kHz). Default filters target 88.2 kHz output.
- **Performance**: ~14× realtime on a Core2Duo @ 2.4 GHz for DSD64 → 24/88.2 FLAC.
- **DoP output mode**: Can pack DSD into DoP-format FLAC files for passthrough to DoP-capable DACs.
- **Piped output**: Can output to stdout for piping to ffmpeg or ALSA, enabling DSD playback without intermediate files.
- **DSD256 support**: dhalsimax fork supports up to DSD256 → DXD conversion.
- **Use case**: Batch converting a DSD library to high-res FLAC for players that don't support DSD natively.

### dsd2dxd — Notable Fork

- **Developer**: clone206 (GitHub). Written in Rust.
- **What it is**: Enhanced DSD→PCM converter based on dsd2pcm. Supports DSF, DFF, and raw DSD bitstreams.
- **Custom filters**: Many decimation filters designed from scratch using listening tests, targeting audiophile-quality conversion. Supports equiripple, XLD-derived, and original dsd2pcm filters.
- **Recommended conversions**: DSD64→96 kHz, DSD128→176.4 kHz, DSD256→192 kHz.
- **Non-integer ratios**: Can convert to rates that aren't integer divisors of the DSD rate (e.g., DSD64 → 96 kHz), which most other tools don't support.

---

## DSD Playback Modes Explained

Since this is confusing even for experts, here's a definitive breakdown:

### 1. DSD → PCM Conversion (Software)
The decoder converts DSD to multi-bit PCM in software. The DAC receives normal PCM. This is what FFmpeg, VLC, mpv, and most players do. Quality depends on the decimation filter. The DAC never sees DSD.

### 2. DoP (DSD over PCM)
The DSD bitstream is packed into 24-bit PCM frames with special marker bytes (0x05/0xFA or 0x06/0xF9 in the 8 MSBs of each 24-bit sample). The DAC's firmware recognizes these markers and extracts the DSD bitstream for native DSD-to-analog conversion. From the OS audio stack's perspective, it looks like 176.4 kHz/24-bit PCM (for DSD64). Invented by dCS. Works over USB, S/PDIF, HDMI — anywhere PCM works.

**DoP bandwidth**: DSD64 requires 176.4 kHz PCM (so your DAC must support 176.4 kHz). DSD128 requires 352.8 kHz. DSD256 requires 705.6 kHz (very few interfaces support this). DSD512+ is impractical over DoP.

### 3. Native DSD (ASIO DSD / ALSA DSD)
The DSD bitstream is sent directly to the DAC driver without PCM framing. Requires explicit DSD support in the driver:
- **Windows**: ASIO 2.1+ with DSD mode (Steinberg extension). Bypasses Windows audio stack entirely.
- **Linux**: ALSA DSD support (in kernel). The ALSA driver exposes DSD-specific sample formats.
- **macOS**: No native DSD support in CoreAudio. Must use DoP, or vendor-specific ASIO-like drivers.

### 4. DSD-to-Analog (in DAC hardware)
The DAC's internal delta-sigma modulator directly converts the DSD bitstream to analog. This is the "purest" path — the DSD bitstream goes straight to the reconstruction filter and analog output stage without any digital intermediate conversion. This is what SACD players do.

---

## OS-Level DSD Support

| OS | Native DSD Container Support | Native DSD Playback | Notes |
|----|------------------------------|---------------------|-------|
| **macOS / iOS** | ✗ | ✗ | Zero DSD support in CoreAudio. No DSF/DFF file type IDs. No DSD codec. DoP is the only path (via third-party players). |
| **Windows** | ✗ | Via ASIO only | Windows audio stack (WASAPI, DirectSound) has no DSD awareness. ASIO 2.1+ can carry native DSD but bypasses Windows entirely. |
| **Linux** | ✗ | Via ALSA DSD | ALSA has native DSD sample format support in the kernel. MPD, HQPlayer Embedded, and others can use it. PulseAudio/PipeWire do NOT support DSD. |
| **Android** | ✗ | ✗ | No native DSD. Some third-party players (USB Audio Player PRO) implement their own USB audio stack for DSD via OTG. |

**The bottom line**: No operating system natively supports DSD. It's all third-party software and specialized audio drivers, everywhere.

---

## Summary: Which Decoder to Use?

| Use Case | Recommended | Why |
|----------|-------------|-----|
| Desktop playback (Windows) | **foobar2000 + foo_input_sacd** | Free, comprehensive format support, native DSD output, active development |
| Desktop playback (macOS) | **Audirvāna** or **HQPlayer** | DoP output. No free option with native DSD quality. |
| Desktop playback (Linux) | **MPD** + ALSA DSD | Native DSD output. Headless server architecture. |
| Networked/streamer | **MPD** or **Roon** + **HQPlayer** | MPD for simple, Roon+HQPlayer for maximum DSP |
| Batch DSD→FLAC conversion | **dsf2flac** or **FFmpeg** | dsf2flac for better defaults; FFmpeg for scripting (remember `-af lowpass`) |
| Best DSD→PCM filter quality | **HQPlayer** | 14 algorithms, configurable, measurably superior |
| General media player | **VLC / mpv** (via FFmpeg) | Just works. Adequate quality. No native DSD output. |
| SACD ISO extraction | **sacd_extract** (EuFlo fork) | The only tool. Extracts to DSF/DFF with DST decompression. |
| Maximum format compatibility | **foobar2000 + foo_input_sacd** | Only option supporting DSF + DFF + SACD ISO + WavPack DSD + DST |