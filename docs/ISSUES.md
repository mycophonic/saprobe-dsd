# Known Issues

## Float WAV output not played by all DACs

The example decoder (`dsd-example-decoder`) writes IEEE float32 WAV files (format tag 3).
Some DACs do not output sound when receiving float format, particularly in exclusive/bitperfect mode.

### Reproduction

```bash
# Decode at 352800 Hz (default) — plays fine:
./bin/dsd-example-decoder "file.dsf" > test.wav
# mpv output: AO: [coreaudio_exclusive] 352800Hz stereo 2ch s32

# Decode at 44100 Hz — no audible output through DAC:
./bin/dsd-example-decoder --rate 44100 "file.dsf" > test2.wav
# mpv output:
#   AO: [coreaudio_exclusive] 44100Hz stereo 2ch float
#   [ao/coreaudio_exclusive] Stream format changed! Reloading.
#   AO: [coreaudio_exclusive] 44100Hz stereo 2ch float
```

The file contains valid audio data (confirmed via ffmpeg volumedetect: mean -28.1 dB, peak -0.9 dB),
but the DAC produces no sound when CoreAudio exclusive mode delivers float format.

### Workaround

Convert to integer PCM before playback:

```bash
ffmpeg -i test2.wav -c:a pcm_s24le test2_int.wav
mpv --audio-exclusive=yes test2_int.wav
```

### Fix

The example decoder's WAV writer should quantize float32 samples to integer PCM (s32 or s24)
and use format tag 1 (PCM integer) instead of tag 3 (IEEE float). The library's `Decoder.Read`
should continue returning float32 bytes — the conversion belongs at the WAV output boundary.
