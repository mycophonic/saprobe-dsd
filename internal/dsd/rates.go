package dsd

import "fmt"

// InputRate represents a DSD sample rate in Hz.
type InputRate int

// Supported DSD rates.
const (
	DSD64   InputRate = 2_822_400
	DSD128  InputRate = 5_644_800
	DSD256  InputRate = 11_289_600
	DSD512  InputRate = 22_579_200
	DSD1024 InputRate = 45_158_400
)

// String returns a human-readable label for the DSD rate.
func (r InputRate) String() string {
	switch r {
	case DSD64:
		return "DSD64"
	case DSD128:
		return "DSD128"
	case DSD256:
		return "DSD256"
	case DSD512:
		return "DSD512"
	case DSD1024:
		return "DSD1024"
	default:
		return fmt.Sprintf("InputRate(%d)", int(r))
	}
}

// OutputRate represents a PCM output sample rate in Hz.
type OutputRate int

// Supported output rates (44.1 kHz family only).
const (
	Rate44100  OutputRate = 44_100
	Rate88200  OutputRate = 88_200
	Rate176400 OutputRate = 176_400
	Rate352800 OutputRate = 352_800
	Rate705600 OutputRate = 705_600
)

// String returns a human-readable label for the output rate.
func (r OutputRate) String() string {
	switch r {
	case Rate44100:
		return "44.1 kHz"
	case Rate88200:
		return "88.2 kHz"
	case Rate176400:
		return "176.4 kHz"
	case Rate352800:
		return "352.8 kHz"
	case Rate705600:
		return "705.6 kHz"
	default:
		return fmt.Sprintf("OutputRate(%d)", int(r))
	}
}
