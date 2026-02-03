package dsd

// DefaultNoiseShapeCoeffs contains noise shaping coefficients for 16-bit
// quantization of DSD-to-PCM output. Two second-order sections (SOS) that
// can preserve a dynamic range of 135 dB below 30 kHz.
//
// Layout: b1, b2, a1, a2 per section.
//
//nolint:gochecknoglobals
var DefaultNoiseShapeCoeffs = []float32{
	// Section 1
	-1.62666423, 0.79410094, 0.61367127, 0.23311013,
	// Section 2
	-1.44870017, 0.54196219, 0.03373857, 0.70316556,
}
