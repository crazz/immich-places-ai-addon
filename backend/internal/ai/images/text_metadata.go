package images

import (
	"bytes"
	"compress/zlib"
	"io"
)

const maxTextMetadata = 64 << 10

func unsupportedTextOrientation(data []byte) bool {
	return len(data) > maxTextMetadata || bytes.Contains(bytes.ToLower(data), []byte("orientation"))
}

func validatePNGText(kind string, data []byte, remaining *int) error {
	end := bytes.IndexByte(data, 0)
	if end < 1 || end > 79 {
		return ErrInvalid
	}
	keyword, text := data[:end], data[end+1:]
	if unsupportedTextOrientation(keyword) || bytes.Contains(bytes.ToLower(keyword), []byte("raw profile type exif")) {
		return ErrInvalid
	}
	compressed := false
	switch kind {
	case "zTXt":
		if len(text) < 1 || text[0] != 0 {
			return ErrInvalid
		}
		compressed, text = true, text[1:]
	case "iTXt":
		if len(text) < 2 || text[0] > 1 || text[1] != 0 {
			return ErrInvalid
		}
		compressed, text = text[0] == 1, text[2:]
		for range 2 {
			end = bytes.IndexByte(text, 0)
			if end < 0 {
				return ErrInvalid
			}
			text = text[end+1:]
		}
	}
	if compressed {
		reader, err := zlib.NewReader(bytes.NewReader(text))
		if err != nil {
			return ErrInvalid
		}
		defer reader.Close()
		text, err = io.ReadAll(io.LimitReader(reader, int64(*remaining)+1))
		if err != nil {
			return ErrInvalid
		}
		defer clear(text)
	}
	if len(text) > *remaining {
		return ErrLimit
	}
	*remaining -= len(text)
	if unsupportedTextOrientation(text) {
		return ErrInvalid
	}
	return nil
}
