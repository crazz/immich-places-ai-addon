package images

import (
	"bytes"
	"encoding/binary"
	"mime"

	_ "golang.org/x/image/webp"
)

func validateContainer(data []byte, declared, decoded string) error {
	media, _, err := mime.ParseMediaType(declared)
	if err != nil || media != "image/"+decoded {
		return ErrInvalid
	}
	switch decoded {
	case "jpeg":
		return validateJPEGMetadata(data)
	case "png":
		return validatePNG(data)
	case "webp":
		return validateWebP(data)
	default:
		return ErrInvalid
	}
}

func validatePNG(data []byte) error {
	if len(data) < 8 {
		return ErrInvalid
	}
	remainingText := maxTextMetadata
	for pos := 8; pos < len(data); {
		if len(data)-pos < 12 {
			return ErrInvalid
		}
		size := int64(binary.BigEndian.Uint32(data[pos : pos+4]))
		if size > int64(len(data)-pos-12) {
			return ErrInvalid
		}
		kind := string(data[pos+4 : pos+8])
		if kind == "acTL" || kind == "fcTL" || kind == "fdAT" || kind == "eXIf" {
			return ErrInvalid
		}
		if kind == "tEXt" || kind == "zTXt" || kind == "iTXt" {
			if err := validatePNGText(kind, data[pos+8:pos+8+int(size)], &remainingText); err != nil {
				return err
			}
		}
		pos += int(size) + 12
		if kind == "IEND" {
			if size != 0 || pos != len(data) {
				return ErrInvalid
			}
			return nil
		}
	}
	return ErrInvalid
}

func validateWebP(data []byte) error {
	if len(data) < 12 || string(data[:4]) != "RIFF" || string(data[8:12]) != "WEBP" || int64(binary.LittleEndian.Uint32(data[4:8])) != int64(len(data)-8) {
		return ErrInvalid
	}
	images := 0
	for pos := 12; pos < len(data); {
		if len(data)-pos < 8 {
			return ErrInvalid
		}
		size := int64(binary.LittleEndian.Uint32(data[pos+4 : pos+8]))
		padded := size + size%2
		if padded > int64(len(data)-pos-8) {
			return ErrInvalid
		}
		kind := string(data[pos : pos+4])
		if kind == "ANIM" || kind == "ANMF" || kind == "EXIF" {
			return ErrInvalid
		}
		if kind == "XMP " && unsupportedTextOrientation(data[pos+8:pos+8+int(size)]) {
			return ErrInvalid
		}
		if kind == "VP8X" && (size != 10 || data[pos+8]&0x0a != 0) {
			return ErrInvalid
		}
		if kind == "VP8 " || kind == "VP8L" {
			images++
		}
		pos += 8 + int(padded)
	}
	if images != 1 {
		return ErrInvalid
	}
	return nil
}

func validateJPEGMetadata(data []byte) error {
	if len(data) < 4 || data[0] != 255 || data[1] != 216 {
		return ErrInvalid
	}
	seenExif := false
	for pos := 2; pos < len(data); {
		if data[pos] != 255 {
			return ErrInvalid
		}
		for pos < len(data) && data[pos] == 255 {
			pos++
		}
		if pos >= len(data) {
			return ErrInvalid
		}
		marker := data[pos]
		pos++
		if marker == 0xda {
			return nil
		}
		if marker == 0xd9 {
			return ErrInvalid
		}
		if len(data)-pos < 2 {
			return ErrInvalid
		}
		size := int(binary.BigEndian.Uint16(data[pos : pos+2]))
		if size < 2 || size > len(data)-pos {
			return ErrInvalid
		}
		segment := data[pos+2 : pos+size]
		if marker == 0xe1 {
			if bytes.HasPrefix(segment, []byte("Exif\x00\x00")) {
				if seenExif || validateEXIF(segment[6:]) != nil {
					return ErrInvalid
				}
				seenExif = true
			} else if bytes.Contains(bytes.ToLower(segment), []byte("orientation")) {
				return ErrInvalid
			}
		}
		pos += size
	}
	return ErrInvalid
}

// Imaging applies JPEG orientation but treats malformed EXIF as absent. Validate
// the orientation directory first so unsupported values cannot silently pass.
func validateEXIF(data []byte) error {
	if len(data) < 8 {
		return ErrInvalid
	}
	var order binary.ByteOrder
	switch string(data[:2]) {
	case "II":
		order = binary.LittleEndian
	case "MM":
		order = binary.BigEndian
	default:
		return ErrInvalid
	}
	if order.Uint16(data[2:4]) != 42 {
		return ErrInvalid
	}
	offset := int64(order.Uint32(data[4:8]))
	if offset < 8 || offset > int64(len(data)-2) {
		return ErrInvalid
	}
	count := int64(order.Uint16(data[offset : offset+2]))
	start := offset + 2
	if count*12+4 > int64(len(data))-start {
		return ErrInvalid
	}
	seen := false
	for i := int64(0); i < count; i++ {
		entry := data[start+i*12 : start+(i+1)*12]
		if order.Uint16(entry[:2]) != 0x112 {
			continue
		}
		value := order.Uint16(entry[8:10])
		if seen || order.Uint16(entry[2:4]) != 3 || order.Uint32(entry[4:8]) != 1 || value < 1 || value > 8 {
			return ErrInvalid
		}
		seen = true
	}
	return nil
}
