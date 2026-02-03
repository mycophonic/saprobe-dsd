# High-End Audiophile DAC Format Support Research

## Purpose

This document surveys the top 20 high-end audiophile DACs currently on the market (2024-2026) to
understand what input formats and sample rates a DSD-to-PCM converter should target for maximum
compatibility with the best DACs available.

---

## DAC Survey

### 1. dCS Vivaldi APEX DAC

| Attribute | Value |
|---|---|
| **Manufacturer** | dCS (Data Conversion Systems) |
| **Price** | ~$46,500 |
| **Max PCM** | 24-bit / 384 kHz (USB, Dual AES); 192 kHz (single AES/S/PDIF) |
| **Max DSD** | DSD128 (via USB DoP, Dual AES DoP, or dCS-encrypted DSD) |
| **DoP Support** | Yes |
| **Native DSD** | Yes (dCS-encrypted DSD via Dual AES) |
| **DAC Chip** | Proprietary dCS Ring DAC (FPGA-based) |
| **Notes** | MQA decoding. Internal upsampling to DXD or DSD. |

**Sources:** [Paragon Sight & Sound](https://www.paragonsns.com/products/dcs-vivaldi-apex-dac), [Stereophile](https://www.stereophile.com/content/dcs-vivaldi-apex-da-processor)

---

### 2. dCS Rossini APEX DAC

| Attribute | Value |
|---|---|
| **Manufacturer** | dCS (Data Conversion Systems) |
| **Price** | ~$32,800 |
| **Max PCM** | 24-bit / 384 kHz (USB, Dual AES, Ethernet); 192 kHz (single AES/S/PDIF); 96 kHz (Toslink) |
| **Max DSD** | DSD128 (USB DoP, Ethernet, Dual AES DoP); DSD64 (single AES/S/PDIF DoP) |
| **DoP Support** | Yes |
| **Native DSD** | Yes (dCS-encrypted DSD via Dual AES); DSD via Ethernet in DFF/DSF |
| **DAC Chip** | Proprietary dCS Ring DAC APEX (FPGA-based) |
| **Notes** | MQA full decoder. Internal upsampling to DXD, DSD, or DSDx2. |

**Sources:** [Paragon Sight & Sound](https://www.paragonsns.com/products/dcs-rossini-apex-dac), [dCS](https://dcsaudio.com/product/rossiniapex), [Stereophile](https://www.stereophile.com/content/dcs-rossini-apex-da-processor)

---

### 3. MSB Select DAC II

| Attribute | Value |
|---|---|
| **Manufacturer** | MSB Technology |
| **Price** | ~$84,500 (base); >$100,000 fully configured |
| **Max PCM** | 32-bit / 768 kHz (Pro USB, Renderer); internal DAC modules capable up to 3,072 kHz (3.072 MHz) |
| **Max DSD** | DSD512 (8xDSD) via Pro USB; DSD256 (4xDSD) via Renderer |
| **DoP Support** | Yes (all inputs) |
| **Native DSD** | Yes (massive parallel single-bit discrete DSD mode) |
| **DAC Chip** | Proprietary discrete R-2R ladder (Hybrid DAC modules, 8 modules / 16 channels) |
| **Notes** | Separate PCM (multibit) and DSD (single-bit) conversion modes. Femto 33 clock. |

**Sources:** [MSB Technology](https://msbtechnology.com/dacs/select-features/), [SoundStage Ultra](https://www.soundstageultra.com/index.php/equipment-menu/1048-msb-technology-premier-digital-to-analog-converter)

---

### 4. MSB Premier DAC

| Attribute | Value |
|---|---|
| **Manufacturer** | MSB Technology |
| **Price** | ~$24,950 (base, without input modules) |
| **Max PCM** | 32-bit / 768 kHz (Pro USB, Renderer); internal modules capable up to 3,072 kHz |
| **Max DSD** | DSD512 (8xDSD) via Pro USB; DSD256 (4xDSD) via Renderer |
| **DoP Support** | Yes (all inputs) |
| **Native DSD** | Yes |
| **DAC Chip** | Proprietary discrete R-2R ladder (4x Prime DAC modules, fully balanced) |
| **Notes** | Same DAC module architecture as Select, with 4 modules instead of 8. Femto 93 clock. |

**Sources:** [MSB Technology](https://msbtechnology.com/dacs/premier-dac-features/), [SoundStage Ultra](https://www.soundstageultra.com/index.php/equipment-menu/1048-msb-technology-premier-digital-to-analog-converter)

---

### 5. Chord Electronics DAVE

| Attribute | Value |
|---|---|
| **Manufacturer** | Chord Electronics |
| **Price** | ~$13,000-$15,000 |
| **Max PCM** | 32-bit / 768 kHz (USB only); 384 kHz (BNC coax); 192 kHz (Toslink); 96 kHz (AES) |
| **Max DSD** | DSD512 (USB); lower rates on other inputs |
| **DoP Support** | Yes |
| **Native DSD** | Yes (USB) |
| **DAC Chip** | Proprietary FPGA (Spartan 6), 164,000-tap filter |
| **Notes** | Separate FPGA programs for PCM and DSD. Also a preamplifier and headphone amp. Compatible with M Scaler for upsampling. |

**Sources:** [Chord Electronics](https://chordelectronics.co.uk/product/dave), [Stereophile](https://www.stereophile.com/content/chord-electronics-dave-da-processor), [Moon Audio](https://www.moon-audio.com/products/chord-dave-dac)

---

### 6. T+A SDV 3100 HV (G2)

| Attribute | Value |
|---|---|
| **Manufacturer** | T+A (Theory + Application) |
| **Price** | ~$38,000-$42,500 (EUR 29,000) |
| **Max PCM** | 32-bit / 768 kHz (USB); 192 kHz (S/PDIF) |
| **Max DSD** | DSD1024 (49.2 MHz) via USB (requires Windows driver or Linux kernel 4.4+) |
| **DoP Support** | Yes |
| **Native DSD** | Yes (True 1-Bit Analogue Converter with native bitstream path) |
| **DAC Chip** | Dual architecture: 4x 32-bit Sigma-Delta D/A converters per channel (PCM path) + T+A True 1-Bit converter (DSD path) |
| **Notes** | Two entirely different digital pathways for PCM vs DSD. Four selectable digital filters. DSD512/DSD1024 require specific OS drivers. |

**Sources:** [T+A Official](https://www.ta-hifi.de/en/audiosystems/hv-series/sdv-3100-reference-streaming-dac-preamp/), [The Absolute Sound](https://www.theabsolutesound.com/articles/theory-application-elektroakustik-sdv-3100-hv-dacpreamplifier-and-pdt-3100-hv-cdsacd-transport/)

---

### 7. Nagra HD DAC X

| Attribute | Value |
|---|---|
| **Manufacturer** | Nagra (Audio Technology Switzerland) |
| **Price** | ~$65,000-$70,000+ |
| **Max PCM** | 24-bit / 384 kHz (USB, Nagra Link ST fiber); 192 kHz (AES/EBU, S/PDIF, Toslink) |
| **Max DSD** | DSD256 (USB, Nagra Link); some sources claim DSD512 |
| **DoP Support** | Yes |
| **Native DSD** | Yes (USB) |
| **DAC Chip** | FPGA-based; internal conversion of all signals to DSD256 (11.2 MHz) |
| **Notes** | Dual-mono topology. 37 regulated ultra-low-noise power supplies. JAN5963 tube output stage. 8-core X-Mos USB chip. |

**Sources:** [Nagra Official](https://www.nagraaudio.com/product/hd-dac-x/), [Alma Music and Audio](https://almaaudio.com/products/nagra-hd-dac-x)

---

### 8. Esoteric N-01XD SE

| Attribute | Value |
|---|---|
| **Manufacturer** | Esoteric (TEAC) |
| **Price** | ~$22,000-$26,000 |
| **Max PCM** | 48-bit / 768 kHz (ES-LINK 5 / HDMI); 32-bit / 384 kHz (USB, Ethernet); 24-bit / 192 kHz (AES, S/PDIF, optical) |
| **Max DSD** | DSD512 (22.5 MHz) via USB, Ethernet, ES-LINK 5; DSD64 (DoP) on AES/S/PDIF |
| **DoP Support** | Yes (AES, S/PDIF, optical) |
| **Native DSD** | Yes (USB, Ethernet, ES-LINK 5) |
| **DAC Chip** | Proprietary Master Sound Discrete DAC (FPGA-based, 64-bit/512Fs Delta-Sigma modulator) |
| **Notes** | Separate FPGA algorithms for DSD and PCM. ES-Link 5 proprietary interface for highest rates. Network streaming built-in (Roon Ready). |

**Sources:** [Playback Distribution](https://playbackdistribution.com/products/esoteric-n-01xd-se-network-dac), [Esoteric Official](https://www.esoteric.jp/en/product/n-01xd/top)

---

### 9. Holo Audio May KTE

| Attribute | Value |
|---|---|
| **Manufacturer** | Holo Audio (Kitsune) |
| **Price** | ~$5,600 (KTE); $3,798 (Level 1); $4,298 (Level 2) |
| **Max PCM** | Up to 1,536 kHz (1.536 MHz) via USB/I2S; theoretically 3,072 kHz |
| **Max DSD** | DSD1024 native via USB/I2S; theoretically DSD2048 (untested) |
| **DoP Support** | Yes (all digital inputs) |
| **Native DSD** | Yes (world's first native DSD on an R-2R ladder DAC) |
| **DAC Chip** | Proprietary discrete R-2R ladder (dual mono, separate DSD and PCM ladder networks) |
| **Notes** | Four operating modes: NOS, OS, OS/PCM, OS/DSD. Dual independent R-2R networks: one optimized for PCM with linear compensation, one optimized for DSD. 0.00005% tolerance accuracy claimed. |

**Sources:** [Kitsune HiFi](https://kitsunehifi.com/products/holoaudio-may-dac-us), [Stereophile](https://www.stereophile.com/content/holoaudio-may-level-3-da-processor), [Magna HiFi](https://www.magnahifi.com/shop/holo-may-kte-22)

---

### 10. Meitner Audio MA3

| Attribute | Value |
|---|---|
| **Manufacturer** | Meitner Audio / EMM Labs |
| **Price** | ~$10,500 |
| **Max PCM** | 32-bit / 384 kHz (USB); 24-bit / 192 kHz (AES/EBU, S/PDIF, Toslink, Ethernet) |
| **Max DSD** | DSD128 (USB); DSD64 (all other inputs via DoP) |
| **DoP Support** | Yes |
| **Native DSD** | Yes (USB) |
| **DAC Chip** | Proprietary MDAC2 fully discrete single-bit converters; internal upsampling to DSD1024 (16xDSD) via MDAT2 DSP |
| **Notes** | Everything internally upsampled to DSD1024. Real-time transient detection in DSP. MQA full unfolding via USB and Ethernet. Roon Ready. |

**Sources:** [EMM Labs / Meitner](https://emmlabs-meitner.com/products/ma3), [Stereophile](https://www.stereophile.com/content/meitner-ma3-integrated-da-processor), [SoundStage Ultra](https://www.soundstageultra.com/index.php/equipment-menu/1060-meitner-audio-ma3-dac-preamplifier)

---

### 11. Weiss DAC502 MK2

| Attribute | Value |
|---|---|
| **Manufacturer** | Weiss Engineering |
| **Price** | ~$12,995 (DAC502 MK2); $11,995 (DAC501 MK2) |
| **Max PCM** | 32-bit / 384 kHz (USB, Ethernet); 24-bit / 192 kHz (AES/EBU, S/PDIF, optical) |
| **Max DSD** | DSD128 (USB, Ethernet only) |
| **DoP Support** | Yes |
| **Native DSD** | Not specified (DSD via USB/Ethernet) |
| **DAC Chip** | 2x 32-bit ESS converter chips (dual channel for improved SNR) |
| **Notes** | Internal fixed resampling to ~195 kHz. DSP capabilities. Future format support via software updates. |

**Sources:** [Weiss Engineering](https://weiss.ch/products/highend-hifi/dac501/), [Stereophile](https://www.stereophile.com/content/weiss-engineering-dac502-da-processor)

---

### 12. Mola Mola Tambaqui

| Attribute | Value |
|---|---|
| **Manufacturer** | Mola Mola |
| **Price** | ~$13,400-$15,325 |
| **Max PCM** | 32-bit / 384 kHz (USB, Roon/Ethernet); 24-bit / 192 kHz (all other inputs) |
| **Max DSD** | DSD256 (Quad-DSD, 11.2896 MHz) via USB and Roon/Ethernet |
| **DoP Support** | Yes |
| **Native DSD** | Yes (USB) |
| **DAC Chip** | Proprietary discrete 32-stage PWM DAC with analog FIR output stage |
| **Notes** | Internal upsampling to 3.125 MHz / 32-bit. 130 dB SNR. Designed by Bruno Putzeys. Roon Ready, Bluetooth (SBC/AAC/aptX/LDAC). |

**Sources:** [Mola Mola Official](https://www.mola-mola.nl/tambaqui.php), [SoundStage Ultra](https://www.soundstageultra.com/index.php/equipment-menu/988-mola-mola-tambaqui-digital-to-analog-converter)

---

### 13. Bricasti M21 Platinum

| Attribute | Value |
|---|---|
| **Manufacturer** | Bricasti Design |
| **Price** | ~$17,000 (Classic); $19,000 (Platinum) |
| **Max PCM** | 24-bit / 384 kHz |
| **Max DSD** | DSD128 (5.6 MHz) |
| **DoP Support** | Yes |
| **Native DSD** | Yes (true 1-bit analog DSD converter with 10 GHz analog switch) |
| **DAC Chip** | Three independent paths: ADI 1955 (sigma-delta PCM), 20-bit ADI (ladder PCM), true 1-bit analog (DSD) |
| **Notes** | Three selectable DAC conversion paths. The DSD path is fully analog with no DSP -- uses a 10 GHz analog switch followed by analog filtering. No up/down sampling for DSD. |

**Sources:** [Bricasti Official](https://www.bricasti.com/en/consumer/m21_specs.php), [Stereophile](https://www.stereophile.com/content/bricasti-design-m21-da-processor)

---

### 14. PS Audio DirectStream MK2

| Attribute | Value |
|---|---|
| **Manufacturer** | PS Audio |
| **Price** | ~$7,999 |
| **Max PCM** | 24-bit / 705.6 kHz (USB, I2S/HDMI); 192 kHz (AES, S/PDIF); 96 kHz (Toslink) |
| **Max DSD** | DSD256 (USB, I2S/HDMI); DSD128 (Dual AES DoP); DSD64 (single AES/S/PDIF DoP) |
| **DoP Support** | Yes |
| **Native DSD** | Yes (I2S/HDMI, USB) |
| **DAC Chip** | Proprietary twin FPGA (no off-the-shelf DAC chip) |
| **Notes** | All inputs internally upsampled to 50-bit at 11.2896 MHz, then converted to 1-bit DSD at 11.2896 MHz. Pure DSD conversion architecture. |

**Sources:** [PS Audio](https://www.psaudio.com/products/directstream-dac-mk2), [Stereophile](https://www.stereophile.com/content/ps-audio-perfectwave-directstream-mk2-da-processor)

---

### 15. LampizatOr Pacific 2

| Attribute | Value |
|---|---|
| **Manufacturer** | LampizatOr |
| **Price** | ~$27,000-$35,700 (configuration dependent) |
| **Max PCM** | 32-bit / 768 kHz (USB); up to 705 kHz DXD (USB); 192 kHz (S/PDIF) |
| **Max DSD** | DSD512 (USB, LAN auto-sensing) |
| **DoP Support** | Yes |
| **Native DSD** | Yes |
| **DAC Chip** | Proprietary "Engine Eleven-P" chipset (Pacific 2); tube-based analog output stage |
| **Notes** | Vacuum tube DAC with rectifier tubes (5U4G/274B/GZ34 etc.) and power tubes (PX4/300B/AD-1 etc.). Fully balanced DSD/PCM digital engine. Roon Bridge via RJ45. |

**Sources:** [LampizatOr North America](https://www.lampizator.com/the-pacific-dac-2), [LampizatOr Poland](https://www.lampizatorpoland.com/pacific-dac)

---

### 16. Luxman DA-07X

| Attribute | Value |
|---|---|
| **Manufacturer** | Luxman |
| **Price** | ~$7,000-$8,000 |
| **Max PCM** | 32-bit / 768 kHz (USB); 24-bit / 192 kHz (S/PDIF, AES/EBU) |
| **Max DSD** | DSD512 (22.5 MHz, 1-bit) via USB; DSD64 (DoP) via S/PDIF |
| **DoP Support** | Yes |
| **Native DSD** | Yes (USB) |
| **DAC Chip** | ROHM BD34301EKV ("MUS-IC" series), dual-mono configuration |
| **Notes** | MQA support. Two digital filter modes for PCM/MQA, two analog filter modes for DSD. Inherited technology from D-10X flagship SACD player. |

**Sources:** [Luxman Official](https://www.luxman.com/product/detail.php?id=50), [Music Direct](https://www.musicdirect.com/equipment/dac/luxman-da-07x-dac/)

---

### 17. Accuphase DC-1000

| Attribute | Value |
|---|---|
| **Manufacturer** | Accuphase |
| **Price** | ~$20,000-$22,000 (EUR 22,600) |
| **Max PCM** | 32-bit / 384 kHz |
| **Max DSD** | DSD256 (11.2896 MHz) via USB, HS-LINK |
| **DoP Support** | Yes |
| **Native DSD** | Yes (MDSD processing: 8 phase-shifted DSD signals into 8-channel MDS++ converter) |
| **DAC Chip** | ESS ES9038 PRO (8-channel parallel, MDSD/MDS++ topology) |
| **Notes** | 50th anniversary flagship. 8-circuit parallel-driven MDS++ converter. MDSD converts DSD directly without PCM intermediate step. HS-LINK proprietary interface. |

**Sources:** [Accuphase Official](https://www.accuphase.com/model/dc-1000.html), [Choice Audio](https://choiceaudio.com/products/accuphase-dc-1000-precision-digital-processor)

---

### 18. Aqua Acoustic Quality Formula xHD (Rev. 2)

| Attribute | Value |
|---|---|
| **Manufacturer** | Aqua Acoustic Quality (Italy) |
| **Price** | ~$17,000-$17,500 |
| **Max PCM** | 24-bit / 768 kHz (USB, AQlink/I2S); 192 kHz (BNC, RCA S/PDIF, AES/EBU) |
| **Max DSD** | DSD512 (USB native); DSD256 (AQlink/I2S via DoP); DSD64 (S/PDIF/AES DoP) |
| **DoP Support** | Yes |
| **Native DSD** | Yes (USB, via hybrid XMOS xCore 200 + FPGA) |
| **DAC Chip** | Proprietary "Optologic" FPGA-based pure R-2R ladder (no digital filter, NOS) |
| **Notes** | No oversampling, no digital filter. Modular design. Fully balanced. Hybrid XMOS/FPGA USB implementation. |

**Sources:** [Aqua HiFi](https://www.aquahifi.com/formula.html), [Stereophile](https://www.stereophile.com/content/aqua-acoustic-quality-aqua-formula-xhd-da-processor)

---

### 19. Playback Designs MPD-8 Dream DAC

| Attribute | Value |
|---|---|
| **Manufacturer** | Playback Designs |
| **Price** | ~$26,000 |
| **Max PCM** | 24-bit / 384 kHz (USB, PLINK optical); 192 kHz (AES/EBU, S/PDIF); 96 kHz (Toslink) |
| **Max DSD** | DSD256 (11.2 MHz) via USB, PLINK |
| **DoP Support** | Yes (AES/EBU, S/PDIF) |
| **Native DSD** | Yes (USB via ASIO driver on Windows; DoP on macOS limits to DSD128; native on Linux) |
| **DAC Chip** | Proprietary FPGA-based single-bit DSD DAC |
| **Notes** | All signals internally converted to DSD128, then remodulated to DSD2048 (~50 MHz) before final analog conversion. PLINK proprietary fiber optical link for galvanic isolation. |

**Sources:** [Playback Designs Product Brief](https://playbackdesigns.com/docs/productsheets/PlaybackDesigns_Dream_MPD-8_ProductBrief.pdf), [Darko.Audio](https://darko.audio/2018/07/playback-designs-mpd-8-dream-dac-review/)

---

### 20. Rockna Wavedream Reference Signature

| Attribute | Value |
|---|---|
| **Manufacturer** | Rockna Audio (Romania) |
| **Price** | ~$22,000-$25,000 |
| **Max PCM** | 24-bit / 384 kHz (USB, I2S/HDMI); internally upsampled to 768 kHz / 705.6 kHz |
| **Max DSD** | DSD512 (22.5792 MHz) via USB, I2S/HDMI |
| **DoP Support** | Yes |
| **Native DSD** | DSD is converted to PCM internally and played on the PCM R-2R ladder |
| **DAC Chip** | Proprietary discrete R-2R (4x RD-0 MK2 modules, each with own 28-bit FPGA, sign-magnitude hybrid ladder) |
| **Notes** | 16x internal upsampling to 768/705.6 kHz with 35-bit precision. Parks-McClellan upsampling filter. Selectable linear/minimum/hybrid-phase filters. Roon Bridge, UPnP, AirPlay, HQPlayer NAA. |

**Sources:** [Rockna Audio](https://www.rockna-audio.com/products/wavedream-reference), [Audiophile Style](https://audiophilestyle.com/ca/reviews/review-rockna-wavedream-reference-signature-dac-r1343/)

---

## Berkeley Audio Design Alpha DAC Reference Series 3 (Honorable Mention)

| Attribute | Value |
|---|---|
| **Manufacturer** | Berkeley Audio Design |
| **Price** | ~$22,000 |
| **Max PCM** | 24-bit / 192 kHz (AES/EBU, S/PDIF BNC, Toslink only -- no USB) |
| **Max DSD** | None accepted directly -- Berkeley recommends converting DSD to 176.4 kHz / 24-bit PCM externally |
| **DoP Support** | No |
| **Native DSD** | No |
| **DAC Chip** | Proprietary (multibit, FPGA-based) |
| **Notes** | Deliberately omits USB and Ethernet. Berkeley's philosophy: DSD-to-multibit conversion is better done externally (in software) than inside the DAC. Recommends 176.4 kHz / 24-bit PCM conversion of DSD files. MQA rendering to 384 kHz+. |

**Sources:** [Berkeley Audio](https://www.berkeleyaudiodesign.com/alpha-dac-reference-series-3p), [SoundStage Australia](https://www.soundstageaustralia.com/index.php/reviews/483-berkeley-audio-design-alpha-dac-reference-series-3)

---

## Summary Table

| # | DAC | Price (USD) | Max PCM Rate | Max PCM Bits | Max DSD | DoP | Native DSD | DAC Architecture |
|---|-----|------------|-------------|-------------|---------|-----|-----------|-----------------|
| 1 | dCS Vivaldi APEX | $46,500 | 384 kHz | 24 | DSD128 | Yes | Yes | Proprietary Ring DAC |
| 2 | dCS Rossini APEX | $32,800 | 384 kHz | 24 | DSD128 | Yes | Yes | Proprietary Ring DAC |
| 3 | MSB Select DAC II | $84,500+ | 768 kHz | 32 | DSD512 | Yes | Yes | Discrete R-2R ladder |
| 4 | MSB Premier | $24,950+ | 768 kHz | 32 | DSD512 | Yes | Yes | Discrete R-2R ladder |
| 5 | Chord DAVE | $13,000-15,000 | 768 kHz | 32 | DSD512 | Yes | Yes | Proprietary FPGA |
| 6 | T+A SDV 3100 HV | $38,000-42,500 | 768 kHz | 32 | DSD1024 | Yes | Yes | Sigma-Delta + True 1-Bit |
| 7 | Nagra HD DAC X | $65,000-70,000 | 384 kHz | 24 | DSD256-512 | Yes | Yes | FPGA + tubes |
| 8 | Esoteric N-01XD SE | $22,000-26,000 | 768 kHz | 48 | DSD512 | Yes | Yes | Proprietary FPGA discrete |
| 9 | Holo Audio May KTE | $5,600 | 1,536 kHz | 32 | DSD1024 | Yes | Yes | Discrete R-2R ladder |
| 10 | Meitner MA3 | $10,500 | 384 kHz | 32 | DSD128 | Yes | Yes | Discrete single-bit |
| 11 | Weiss DAC502 MK2 | $12,995 | 384 kHz | 32 | DSD128 | Yes | Yes | ESS ES-series |
| 12 | Mola Mola Tambaqui | $13,400-15,325 | 384 kHz | 32 | DSD256 | Yes | Yes | Proprietary PWM |
| 13 | Bricasti M21 | $17,000-19,000 | 384 kHz | 24 | DSD128 | Yes | Yes | Multi-path (DS + ladder + 1-bit) |
| 14 | PS Audio DS MK2 | $7,999 | 705.6 kHz | 24 | DSD256 | Yes | Yes | Proprietary FPGA |
| 15 | LampizatOr Pacific 2 | $27,000-35,700 | 768 kHz | 32 | DSD512 | Yes | Yes | Proprietary + tubes |
| 16 | Luxman DA-07X | $7,000-8,000 | 768 kHz | 32 | DSD512 | Yes | Yes | ROHM BD34301EKV |
| 17 | Accuphase DC-1000 | $20,000-22,000 | 384 kHz | 32 | DSD256 | Yes | Yes | ESS ES9038 PRO (8x parallel) |
| 18 | Aqua Formula xHD | $17,000-17,500 | 768 kHz | 24 | DSD512 | Yes | Yes | Proprietary R-2R (NOS) |
| 19 | Playback Designs MPD-8 | $26,000 | 384 kHz | 24 | DSD256 | Yes | Yes | Proprietary FPGA single-bit |
| 20 | Rockna Wavedream Ref Sig | $22,000-25,000 | 384 kHz | 24 | DSD512 | Yes | DSD->PCM | Discrete R-2R |
| -- | Berkeley Alpha Ref S3 | $22,000 | 192 kHz | 24 | None | No | No | Proprietary FPGA |

---

## Analysis: Target Formats for a DSD-to-PCM Converter

### Key Findings

#### 1. PCM Sample Rate Support

The maximum PCM sample rates accepted by these DACs (via their highest-capability input, typically USB):

| Max PCM Rate | DAC Count | Percentage |
|---|---|---|
| 192 kHz | 1 (Berkeley) | 5% |
| 384 kHz | 10 | 48% |
| 705.6 kHz | 1 (PS Audio) | 5% |
| 768 kHz | 7 | 33% |
| 1,536 kHz | 1 (Holo Audio) | 5% |

**Conclusion:** 384 kHz is the universal safe target -- 100% of DACs surveyed (except Berkeley at 192 kHz) accept at least 384 kHz PCM via USB. Targeting 352.8 kHz (DXD) or 384 kHz guarantees compatibility with virtually every high-end DAC.

#### 2. PCM Bit Depth Support

| Max Bit Depth | DAC Count | Percentage |
|---|---|---|
| 24-bit | 8 | 38% |
| 32-bit | 11 | 52% |
| 48-bit | 1 (Esoteric ES-LINK) | 5% |

**Conclusion:** 24-bit is universally accepted. 32-bit is widely supported via USB but not required. A converter outputting 24-bit PCM will be compatible with all DACs. Outputting 32-bit would be ideal for the 52% that support it but offers no practical benefit beyond 24-bit for audio content.

#### 3. DSD Support Landscape

| Max DSD Rate | DAC Count | Percentage |
|---|---|---|
| None | 1 (Berkeley) | 5% |
| DSD128 | 5 | 24% |
| DSD256 | 5 | 24% |
| DSD512 | 8 | 38% |
| DSD1024 | 2 | 10% |

**Conclusion:** DSD support varies widely. A significant minority (24%) top out at DSD128. This means a DSD-to-PCM converter has clear value -- it allows DSD256/DSD512/DSD1024 content to be played on DACs that only accept DSD128 or less, or on DACs like Berkeley that accept no DSD at all.

#### 4. DoP vs Native DSD

- **100%** of DSD-capable DACs support DoP (DSD over PCM)
- **Most** also support native DSD via USB (ASIO on Windows, native on Linux)
- DoP is the universal transport mechanism for DSD over USB
- Some DACs (Rockna) accept DSD input but convert it to PCM internally anyway

#### 5. Internal Processing Insights

Many high-end DACs convert between DSD and PCM internally regardless of input format:

| DAC | Internal Processing |
|---|---|
| dCS Rossini/Vivaldi | Upsamples to DXD or DSD internally |
| Meitner MA3 | Upsamples everything to DSD1024 |
| PS Audio DS MK2 | Converts everything to 1-bit DSD at 11.2896 MHz |
| Nagra HD DAC X | Converts everything to DSD256 internally |
| Playback Designs MPD-8 | Converts everything to DSD2048 internally |
| Rockna Wavedream | Upsamples to 768/705.6 kHz PCM, converts DSD to PCM |
| Mola Mola Tambaqui | Upsamples everything to 3.125 MHz / 32-bit PCM |
| Weiss DAC502 | Resamples everything to ~195 kHz internally |

This means that for many DACs, the input format (DSD vs PCM) is less critical than the quality of the conversion, since the DAC will re-convert internally anyway.

### Recommended Target Formats for DSD-to-PCM Conversion

Based on the survey data, a DSD-to-PCM converter should support the following output formats to maximize compatibility:

#### Primary Targets (Maximum Compatibility)

| Output Format | Rationale |
|---|---|
| **352.8 kHz / 24-bit PCM** | DXD rate. Compatible with all surveyed DACs (except Berkeley at 192 kHz). Natural target for DSD64/DSD128 conversion (8x and 4x of 44.1 kHz base). Maintains the full audio bandwidth of DSD128. |
| **176.4 kHz / 24-bit PCM** | Compatible with 100% of surveyed DACs including Berkeley. Berkeley specifically recommends this rate for DSD conversion. Good target for DSD64. Exactly 4x CD sample rate. |

#### Secondary Targets (For DACs Supporting Higher Rates)

| Output Format | Rationale |
|---|---|
| **384 kHz / 24-bit PCM** | Accepted by 95%+ of DACs via USB. Useful for DSD content from 48 kHz base rates. |
| **705.6 kHz / 24-bit PCM** | For the ~40% of DACs supporting 768 kHz. Preserves more of the bandwidth from DSD256/DSD512 source material. |
| **768 kHz / 24-bit or 32-bit PCM** | Maximum rate supported by many flagship DACs. Optimal for DSD512+ source material on capable DACs. |

#### Conversion Rate Mapping

| DSD Source Rate | Recommended PCM Output | Alternative |
|---|---|---|
| DSD64 (2.8224 MHz) | 176.4 kHz / 24-bit | 352.8 kHz / 24-bit |
| DSD128 (5.6448 MHz) | 352.8 kHz / 24-bit | 176.4 kHz / 24-bit |
| DSD256 (11.2896 MHz) | 352.8 kHz / 24-bit | 705.6 kHz / 24-bit |
| DSD512 (22.5792 MHz) | 352.8 kHz / 24-bit | 705.6 kHz / 24-bit |
| DSD1024 (45.1584 MHz) | 705.6 kHz / 24-bit | 352.8 kHz / 24-bit |

### Key Takeaways

1. **352.8 kHz / 24-bit PCM (DXD)** is the single best universal target format for DSD-to-PCM conversion. It is accepted by virtually every high-end DAC via USB, preserves the full audio bandwidth of DSD128, and aligns with the 44.1 kHz sample rate family used by most music.

2. **24-bit depth is sufficient.** While some DACs accept 32-bit, the noise floor of DSD (even DSD512) does not contain meaningful information beyond 24 bits of dynamic range. Outputting 32-bit float internally and dithering to 24-bit integer for output is the recommended approach.

3. **176.4 kHz / 24-bit is the safest fallback** for DACs with limited sample rate support (like Berkeley's 192 kHz maximum). This rate is universally accepted.

4. **The 44.1 kHz family (176.4, 352.8, 705.6 kHz) should be preferred** over the 48 kHz family (192, 384, 768 kHz) for DSD conversion, since DSD's base rate of 2.8224 MHz is derived from 44.1 kHz (64x). Converting to 44.1 kHz multiples avoids unnecessary sample rate conversion artifacts.

5. **Every DAC surveyed supports DoP**, making it the universal DSD transport. However, many audiophile DACs internally convert DSD to PCM (or vice versa) regardless, suggesting that a high-quality external DSD-to-PCM conversion is not inherently inferior to "native" DSD playback.

---

*Research conducted January 2026. Prices and specifications may change. All specifications verified via manufacturer websites, authorized dealer listings, and reputable audio review publications.*
