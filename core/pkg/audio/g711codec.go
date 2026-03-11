// @Title        g711codec
// @Description  main
// @Create       yiyiyi 2025/12/22 14:44

package audio

import "fmt"

const (
	SIGN_BIT   = 0x80 // Sign bit for a A-law/u-law byte
	QUANT_MASK = 0x0F // Quantization field mask
	NSEGS      = 8    // Number of segments
	SEG_SHIFT  = 4    // Left shift for segment number
	SEG_MASK   = 0x70 // Segment field mask
	BIAS       = 0x84 // Bias for linear code
)

var segEnd = []int16{0xFF, 0x1FF, 0x3FF, 0x7FF,
	0xFFF, 0x1FFF, 0x3FFF, 0x7FFF}

var segAEnd = []int16{0x1F, 0x3F, 0x7F, 0xFF,
	0x1FF, 0x3FF, 0x7FF, 0xFFF}

// search finds the segment for the given value
func search(val int, table []int16, size int) int {
	for i := 0; i < size; i++ {
		if val <= int(table[i]) {
			return i
		}
	}
	return size
}

// Alaw2Linear converts an A-law value to 16-bit linear PCM
func Alaw2Linear(aVal byte) int16 {
	t := int(aVal ^ 0x55)
	seg := (t & SEG_MASK) >> SEG_SHIFT

	t = (t & QUANT_MASK) << 4

	switch seg {
	case 0:
		t += 8
	case 1:
		t += 0x108
	default:
		t += 0x108
		t <<= uint(seg - 1)
	}

	if (aVal & SIGN_BIT) != 0 {
		return int16(t)
	}
	return int16(-t)
}

// Ulaw2Linear converts a u-law value to 16-bit linear PCM
func Ulaw2Linear(uVal byte) int16 {
	// Complement to obtain normal u-law value
	uVal = ^uVal

	// Extract and bias the quantization bits
	t := ((int(uVal) & QUANT_MASK) << 3) + BIAS
	// Shift up by the segment number
	t <<= uint((uVal & SEG_MASK) >> SEG_SHIFT)

	if (uVal & SIGN_BIT) != 0 {
		return int16(BIAS - t)
	}
	return int16(t - BIAS)
}

// Linear2Alaw converts a 16-bit linear PCM value to 8-bit A-law
func Linear2Alaw(pcmVal int16) byte {
	var mask byte
	var seg int
	var aval byte

	// Right shift to reduce precision
	val := int(pcmVal) >> 3

	if val >= 0 {
		mask = 0xD5 // sign (7th) bit = 1
	} else {
		mask = 0x55 // sign bit = 0
		val = -val - 1
	}

	// Convert the scaled magnitude to segment number
	seg = search(val, segAEnd, NSEGS)

	// Combine the sign, segment, and quantization bits
	if seg >= NSEGS {
		// Out of range, return maximum value
		return 0x7F ^ mask
	}

	aval = byte(seg << SEG_SHIFT)
	if seg < 2 {
		aval |= byte((val >> 1) & QUANT_MASK)
	} else {
		aval |= byte((val >> uint(seg)) & QUANT_MASK)
	}

	return aval ^ mask
}

// Linear2Ulaw converts a linear PCM value to u-law
func Linear2Ulaw(pcmVal int16) byte {
	var mask byte
	var seg int
	var uval byte

	// Get the sign and the magnitude of the value
	if pcmVal < 0 {
		val := BIAS - int(pcmVal)
		mask = 0x7F
		seg = search(val, segEnd, NSEGS)
	} else {
		val := int(pcmVal) + BIAS
		mask = 0xFF
		seg = search(val, segEnd, NSEGS)
	}

	// Combine the sign, segment, quantization bits and complement the code word
	if seg >= NSEGS {
		// Out of range, return maximum value
		return 0x7F ^ mask
	}

	if pcmVal < 0 {
		val := BIAS - int(pcmVal)
		uval = byte(seg<<4) | byte((val>>uint(seg+3))&QUANT_MASK)
	} else {
		val := int(pcmVal) + BIAS
		uval = byte(seg<<4) | byte((val>>uint(seg+3))&QUANT_MASK)
	}

	return uval ^ mask
}

// G711aDecode decodes A-law encoded data to 16-bit PCM
func G711aDecode(g711aData []byte) []int16 {
	amp := make([]int16, len(g711aData))
	for i, code := range g711aData {
		amp[i] = Alaw2Linear(code)
	}
	return amp
}

// G711uDecode decodes μ-law encoded data to 16-bit PCM
func G711uDecode(g711uData []byte) []int16 {
	amp := make([]int16, len(g711uData))
	for i, code := range g711uData {
		amp[i] = Ulaw2Linear(code)
	}
	return amp
}

// G711aEncode encodes 16-bit PCM data to A-law
func G711aEncode(amp []int16) []byte {
	g711Data := make([]byte, len(amp))
	for i, sample := range amp {
		g711Data[i] = Linear2Alaw(sample)
	}
	return g711Data
}

// G711uEncode encodes 16-bit PCM data to μ-law
func G711uEncode(amp []int16) []byte {
	g711Data := make([]byte, len(amp))
	for i, sample := range amp {
		g711Data[i] = Linear2Ulaw(sample)
	}
	return g711Data
}

// G711aDecodeToBytes decodes A-law encoded data to PCM bytes (little-endian)
func G711aDecodeToBytes(g711aData []byte) []byte {
	amp := G711aDecode(g711aData)
	result := make([]byte, len(amp)*2)
	for i, sample := range amp {
		result[i*2] = byte(sample & 0xFF)
		result[i*2+1] = byte((sample >> 8) & 0xFF)
	}
	return result
}

// G711uDecodeToBytes decodes μ-law encoded data to PCM bytes (little-endian)
func G711uDecodeToBytes(g711uData []byte) []byte {
	amp := G711uDecode(g711uData)
	result := make([]byte, len(amp)*2)
	for i, sample := range amp {
		result[i*2] = byte(sample & 0xFF)
		result[i*2+1] = byte((sample >> 8) & 0xFF)
	}
	return result
}

// Encodes PCM bytes (little-endian) to A-law
func G711AEncode(pcmData []byte) ([]byte, error) {
	if len(pcmData)%2 != 0 {
		return nil, fmt.Errorf("PCM data length must be even")
	}

	amp := make([]int16, len(pcmData)/2)
	for i := 0; i < len(amp); i++ {
		amp[i] = int16(pcmData[i*2]) | int16(pcmData[i*2+1])<<8
	}
	return G711aEncode(amp), nil
}

// Encodes PCM bytes (little-endian) to μ-law
func G711UEncode(pcmData []byte) ([]byte, error) {
	if len(pcmData)%2 != 0 {
		return nil, fmt.Errorf("PCM data length must be even")
	}

	amp := make([]int16, len(pcmData)/2)
	for i := 0; i < len(amp); i++ {
		amp[i] = int16(pcmData[i*2]) | int16(pcmData[i*2+1])<<8
	}
	return G711uEncode(amp), nil
}

// g711a, err := audiocodec.G711AEncode(message)
// if len(g711a) > 0 {
// dev.SendAudio(voiceHandle, g711a)
// }
