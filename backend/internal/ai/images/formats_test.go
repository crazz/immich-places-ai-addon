package images

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/png"
	"os"
	"testing"
)

func pngChunk(kind string, data []byte) []byte {
	result := make([]byte, 12+len(data))
	binary.BigEndian.PutUint32(result, uint32(len(data)))
	copy(result[4:8], kind)
	copy(result[8:], data)
	binary.BigEndian.PutUint32(result[8+len(data):], crc32.ChecksumIEEE(result[4:8+len(data)]))
	return result
}
func pngFixture(t *testing.T) []byte {
	t.Helper()
	var b bytes.Buffer
	if err := png.Encode(&b, image.NewNRGBA(image.Rect(0, 0, 8, 4))); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}
func addPNGChunk(input []byte, kind string, data []byte) []byte {
	out := append(bytes.Clone(input[:33]), pngChunk(kind, data)...)
	return append(out, input[33:]...)
}
func addWebPChunk(input []byte, kind string) []byte {
	out := append(bytes.Clone(input), []byte(kind+"\x00\x00\x00\x00")...)
	binary.LittleEndian.PutUint32(out[4:8], uint32(len(out)-8))
	return out
}

func TestPrepareRasterRejectsUnsupportedContainersAndMIME(t *testing.T) {
	webp, err := os.ReadFile("testdata/static.webp")
	if err != nil {
		t.Fatal(err)
	}
	pngData := pngFixture(t)
	for _, tc := range []struct {
		name, mime string
		data       []byte
	}{
		{"MIME mismatch", "image/png", jpegFixture(t, 8, 4)},
		{"unknown MIME", "image/gif", jpegFixture(t, 8, 4)},
		{"truncated JPEG", "image/jpeg", jpegFixture(t, 8, 4)[:30]},
		{"APNG", "image/png", addPNGChunk(pngData, "acTL", []byte{0, 0, 0, 1, 0, 0, 0, 0})},
		{"PNG EXIF", "image/png", addPNGChunk(pngData, "eXIf", []byte("unsupported orientation"))},
		{"WebP animation", "image/webp", addWebPChunk(webp, "ANIM")},
		{"WebP EXIF", "image/webp", addWebPChunk(webp, "EXIF")},
		{"invalid EXIF orientation", "image/jpeg", orientedJPEG(t, 9)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, err := PrepareRaster(context.Background(), tc.data, tc.mime, testBinding(), DefaultLimits())
			if err == nil || p != nil {
				t.Fatal("unsupported input accepted")
			}
		})
	}
	p, err := PrepareRaster(context.Background(), webp, "image/webp", testBinding(), DefaultLimits())
	if err != nil {
		t.Fatalf("valid static WebP rejected: %v", err)
	}
	defer p.Release()
	info, _ := p.Info()
	if info.Width != 8 || info.Height != 4 {
		t.Fatal("WebP dimensions changed")
	}
}

func TestPrepareRasterRejectsTextualAndCompressedOrientationMetadata(t *testing.T) {
	pngData := pngFixture(t)
	xmp := []byte(`<x:xmpmeta tiff:Orientation="6"/>`)
	var compressed bytes.Buffer
	writer := zlib.NewWriter(&compressed)
	if _, err := writer.Write(xmp); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	webp, err := os.ReadFile("testdata/static.webp")
	if err != nil {
		t.Fatal(err)
	}
	chunk := make([]byte, 8+len(xmp)+len(xmp)%2)
	copy(chunk, "XMP ")
	binary.LittleEndian.PutUint32(chunk[4:8], uint32(len(xmp)))
	copy(chunk[8:], xmp)
	webp = append(webp, chunk...)
	binary.LittleEndian.PutUint32(webp[4:8], uint32(len(webp)-8))
	for _, tc := range []struct {
		mime string
		data []byte
	}{
		{"image/png", addPNGChunk(pngData, "tEXt", append([]byte("XML:com.adobe.xmp\x00"), xmp...))},
		{"image/png", addPNGChunk(pngData, "zTXt", append([]byte("XML:com.adobe.xmp\x00\x00"), compressed.Bytes()...))},
		{"image/png", addPNGChunk(pngData, "iTXt", append([]byte("XML:com.adobe.xmp\x00\x01\x00\x00\x00"), compressed.Bytes()...))},
		{"image/webp", webp},
	} {
		p, err := PrepareRaster(context.Background(), tc.data, tc.mime, testBinding(), DefaultLimits())
		if err == nil || p != nil {
			t.Fatal("unsupported textual orientation accepted")
		}
	}
}

func TestPrepareRasterBoundsTotalExpandedPNGMetadata(t *testing.T) {
	var compressed bytes.Buffer
	writer := zlib.NewWriter(&compressed)
	if _, err := writer.Write(bytes.Repeat([]byte("x"), 40000)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	chunk := append([]byte("Comment\x00\x00"), compressed.Bytes()...)
	input := addPNGChunk(addPNGChunk(pngFixture(t), "zTXt", chunk), "zTXt", chunk)
	p, err := PrepareRaster(context.Background(), input, "image/png", testBinding(), DefaultLimits())
	if err == nil || p != nil {
		t.Fatal("aggregate compressed metadata budget bypassed")
	}
}

func TestPrepareRasterRejectsWebPCanvasBitstreamMismatch(t *testing.T) {
	original, err := os.ReadFile("testdata/static.webp")
	if err != nil {
		t.Fatal(err)
	}
	extended := append(bytes.Clone(original[:12]), []byte{'V', 'P', '8', 'X', 10, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}...)
	extended = append(extended, original[12:]...)
	binary.LittleEndian.PutUint32(extended[4:8], uint32(len(extended)-8))
	p, err := PrepareRaster(context.Background(), extended, "image/webp", testBinding(), DefaultLimits())
	if err == nil || p != nil {
		t.Fatal("small canvas hid a larger bitstream")
	}
}
