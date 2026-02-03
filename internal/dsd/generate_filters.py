#!/usr/bin/env python3
"""Generate Parks-McClellan FIR filter coefficients for DSD-to-PCM conversion.

Outputs Go source file with:
  - Halfband 2:1 decimation filter (symmetric, right-half only)
  - 4:1 decimation filter for DSD64->705.6k edge case (symmetric, right-half only)

Usage:
    python3 generate_filters.py              # print Go source to stdout
    python3 generate_filters.py --verify     # print filter specs and verify Gesemann
"""

import argparse
import sys

import numpy as np
from scipy import signal


def design_halfband(num_taps=127, passband_edge=0.45, stopband_attn_db=140):
    """Design a halfband 2:1 decimation filter using Parks-McClellan.

    For halfband filters, the passband edge and stopband edge are symmetric
    around 0.25 (normalized to Nyquist=1.0). We use a slightly relaxed
    transition band for practical tap counts.

    Args:
        num_taps: Filter length (must be odd for type I symmetric).
        passband_edge: Passband edge as fraction of Nyquist (0.0 to 0.5).
        stopband_attn_db: Target stopband attenuation in dB.

    Returns:
        Right-half coefficients of the symmetric filter.
    """
    # Ensure odd tap count for type I (symmetric, odd length).
    if num_taps % 2 == 0:
        num_taps += 1

    stopband_edge = 1.0 - passband_edge  # symmetric around 0.5

    # Bands: [0, passband_edge], [stopband_edge, 1.0] (normalized to Nyquist=1)
    # For remez, frequencies are normalized to [0, 0.5] where 0.5 = Nyquist.
    bands = [0, passband_edge / 2, stopband_edge / 2, 0.5]
    desired = [1, 0]

    # Weight the stopband more heavily to achieve target attenuation.
    stopband_weight = 10 ** (stopband_attn_db / 20)
    weights = [1, stopband_weight]

    try:
        taps = signal.remez(num_taps, bands, desired, weight=weights, fs=1.0)
    except Exception:
        # If remez fails to converge, try with slightly relaxed params.
        num_taps += 16
        if num_taps % 2 == 0:
            num_taps += 1
        taps = signal.remez(num_taps, bands, desired, weight=weights, fs=1.0)

    # Extract right half (including center tap).
    center = len(taps) // 2
    right_half = taps[center:]

    return right_half, taps


def design_4to1_filter(num_taps=63):
    """Design a 4:1 decimation filter for DSD64 -> 705.6 kHz.

    Operating at DSD64 rate (2.8224 MHz). Output rate is 705.6 kHz,
    so output Nyquist is 352.8 kHz.

    DSD64 has useful audio content only up to ~50 kHz; above that is
    noise-shaped quantization noise. The filter prevents this noise from
    aliasing back into the audio band during 4:1 decimation.

    Requirements:
    - Passband: 0 to 100 kHz (flat for all audible + generous ultrasonic margin)
    - Stopband: 352.8 kHz to 1411.2 kHz (output Nyquist to input Nyquist)
    - Transition band: 252.8 kHz (very wide, making this an easy design)

    Args:
        num_taps: Filter length (must be odd).

    Returns:
        Right-half coefficients of the symmetric filter.
    """
    if num_taps % 2 == 0:
        num_taps += 1

    fs = 2_822_400.0  # DSD64 sample rate

    passband_hz = 100_000.0   # generous: covers all audio + ultrasonic margin
    stopband_hz = 352_800.0   # output Nyquist

    # Normalize to [0, 0.5] for remez (fs=1.0 convention).
    bands = [0, passband_hz / fs, stopband_hz / fs, 0.5]
    desired = [1, 0]

    # Target ~100 dB stopband attenuation (sufficient for DSD noise suppression).
    stopband_weight = 10 ** (100 / 20)
    weights = [1, stopband_weight]

    taps = signal.remez(num_taps, bands, desired, weight=weights, fs=1.0)

    center = len(taps) // 2
    right_half = taps[center:]

    return right_half, taps


def verify_gesemann():
    """Verify the Gesemann 96-tap filter meets its stated specifications."""
    # Gesemann's right-half (48 coefficients).
    htaps = [
        0.09950731974056658, 0.09562845727714668, 0.08819647126516944,
        0.07782552527068175, 0.06534876523171299, 0.05172629311427257,
        0.0379429484910187,  0.02490921351762261, 0.0133774746265897,
        0.003883043418804416, -0.003284703416210726, -0.008080250212687497,
        -0.01067241812471033, -0.01139427235000863, -0.0106813877974587,
        -0.009007905078766049, -0.006828859761015335, -0.004535184322001496,
        -0.002425035959059578, -0.0006922187080790708, 0.0005700762133516592,
        0.001353838005269448, 0.001713709169690937, 0.001742046839472948,
        0.001545601648013235, 0.001226696225277855, 0.0008704322683580222,
        0.0005381636200535649, 0.000266446345425276, 7.002968738383528e-05,
        -5.279407053811266e-05, -0.0001140625650874684, -0.0001304796361231895,
        -0.0001189970287491285, -9.396247155265073e-05, -6.577634378272832e-05,
        -4.07492895872535e-05, -2.17407957554587e-05, -9.163058931391722e-06,
        -2.017460145032201e-06, 1.249721855219005e-06, 2.166655190537392e-06,
        1.930520892991082e-06, 1.319400334374195e-06, 7.410039764949091e-07,
        3.423230509967409e-07, 1.244182214744588e-07, 3.130441005359396e-08,
    ]

    # Reconstruct full symmetric filter.
    full_taps = list(reversed(htaps)) + htaps
    full_taps = np.array(full_taps)

    fs = 2_822_400.0  # DSD64
    w, h = signal.freqz(full_taps, worN=16384, fs=fs)

    mag_db = 20 * np.log10(np.abs(h) + 1e-30)

    # Check passband (0-48 kHz): should be flat (within 0.1 dB).
    passband_mask = w <= 48000
    passband_ripple = np.max(np.abs(mag_db[passband_mask]))
    print(f"  Passband (0-48 kHz): max deviation = {passband_ripple:.4f} dB")

    # Check stopband at multiple points to show rejection profile.
    # 176.4 kHz is right at the transition band edge — expect modest rejection.
    # Deep stopband (>352.8 kHz) is where the high rejection occurs.
    checkpoints = [176_400, 250_000, 352_800, 500_000, 705_600, 1_000_000]
    for freq in checkpoints:
        mask = w >= freq
        if np.any(mask):
            atten = -np.max(mag_db[mask])
            print(f"  Stopband (>={freq/1000:.0f} kHz): min attenuation = {atten:.1f} dB")

    return passband_ripple


def analyze_filter(name, taps_full, fs, passband_hz, stopband_hz):
    """Print frequency response analysis for a filter."""
    w, h = signal.freqz(taps_full, worN=8192, fs=fs)
    mag_db = 20 * np.log10(np.abs(h) + 1e-30)

    passband_mask = w <= passband_hz
    passband_ripple = np.max(np.abs(mag_db[passband_mask]))

    stopband_mask = w >= stopband_hz
    if np.any(stopband_mask):
        stopband_atten = -np.max(mag_db[stopband_mask])
    else:
        stopband_atten = float('inf')

    print(f"  {name}:")
    print(f"    Taps: {len(taps_full)}")
    print(f"    Passband (0-{passband_hz/1000:.1f} kHz): max deviation = {passband_ripple:.4f} dB")
    print(f"    Stopband (>{stopband_hz/1000:.1f} kHz): attenuation = {stopband_atten:.1f} dB")

    return passband_ripple, stopband_atten


def format_go_coefficients(name, right_half, indent="\t"):
    """Format coefficients as a Go float64 array literal."""
    lines = []
    lines.append(f"var {name} = [...]float64{{")
    for coeff in right_half:
        lines.append(f"{indent}{coeff:.17e},")
    lines.append("}")
    return "\n".join(lines)


def emit_go_source(halfband_right, halfband_full, fir4to1_right, fir4to1_full):
    """Print the generated Go source file."""
    lines = []
    lines.append("// Code generated by generate_filters.py using Parks-McClellan (scipy.signal.remez).")
    lines.append("// DO NOT EDIT.")
    lines.append("")
    lines.append("package dsd")
    lines.append("")

    # Halfband filter.
    lines.append("// halfbandTaps holds the right half (including center tap) of a symmetric")
    lines.append(f"// halfband 2:1 decimation filter ({len(halfband_full)} taps total).")
    lines.append("//")
    lines.append("//nolint:gochecknoglobals")
    lines.append(format_go_coefficients("halfbandTaps", halfband_right))
    lines.append("")

    # 4:1 filter.
    lines.append("// fir4to1Taps holds the right half (including center tap) of a symmetric")
    lines.append(f"// 4:1 decimation FIR filter ({len(fir4to1_full)} taps total) for DSD64 to 705.6 kHz.")
    lines.append("//")
    lines.append("//nolint:gochecknoglobals")
    lines.append(format_go_coefficients("fir4to1Taps", fir4to1_right))
    lines.append("")

    return "\n".join(lines)


def main():
    parser = argparse.ArgumentParser(description="Generate DSD filter coefficients")
    parser.add_argument("--verify", action="store_true", help="Verify filters and print specs")
    parser.add_argument("--halfband-taps", type=int, default=127,
                        help="Number of taps for halfband filter (default: 127)")
    parser.add_argument("--fir4to1-taps", type=int, default=63,
                        help="Number of taps for 4:1 filter (default: 63)")
    args = parser.parse_args()

    print("Designing halfband 2:1 filter...", file=sys.stderr)
    halfband_right, halfband_full = design_halfband(num_taps=args.halfband_taps)

    print("Designing 4:1 filter (DSD64->705.6k)...", file=sys.stderr)
    fir4to1_right, fir4to1_full = design_4to1_filter(num_taps=args.fir4to1_taps)

    if args.verify:
        print("\n=== Filter Verification ===\n", file=sys.stderr)

        print("Gesemann 96-tap (8:1, DSD64):", file=sys.stderr)
        verify_gesemann()

        print(file=sys.stderr)

        # Analyze halfband at a reference rate (352.8 kHz input, 176.4 kHz output).
        fs_ref = 352800.0
        analyze_filter(
            f"Halfband 2:1 ({len(halfband_full)} taps)",
            halfband_full,
            fs=fs_ref,
            passband_hz=fs_ref * 0.45 / 2,  # 0.45 of Nyquist
            stopband_hz=fs_ref * 0.55 / 2,  # 0.55 of Nyquist
        )

        print(file=sys.stderr)

        # Analyze 4:1 filter at DSD64 rate.
        analyze_filter(
            f"4:1 FIR ({len(fir4to1_full)} taps, DSD64->705.6k)",
            fir4to1_full,
            fs=2_822_400.0,
            passband_hz=100_000.0,
            stopband_hz=352_800.0,
        )

        return

    # Emit Go source.
    go_src = emit_go_source(halfband_right, halfband_full, fir4to1_right, fir4to1_full)
    print(go_src)


if __name__ == "__main__":
    main()
