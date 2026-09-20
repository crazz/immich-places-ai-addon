package images

import (
	"bytes"
	"context"
	"encoding/binary"
	"image"
	"image/color"
	"image/jpeg"
	"testing"
)

func orientedJPEG(t *testing.T, orientation uint16) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, 40, 20))
	colors := []color.NRGBA{{240, 20, 20, 255}, {20, 240, 20, 255}, {20, 20, 240, 255}, {240, 240, 20, 255}}
	for y := 0; y < 20; y++ {
		for x := 0; x < 40; x++ {
			img.SetNRGBA(x, y, colors[(y/10)*2+x/20])
		}
	}
	var out bytes.Buffer
	if err := jpeg.Encode(&out, img, &jpeg.Options{Quality: 100}); err != nil {
		t.Fatal(err)
	}
	exif := []byte{'E', 'x', 'i', 'f', 0, 0, 'I', 'I', 42, 0, 8, 0, 0, 0, 1, 0, 0x12, 1, 3, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	binary.LittleEndian.PutUint16(exif[24:26], orientation)
	segment := []byte{255, 225, 0, byte(len(exif) + 2)}
	result := append(bytes.Clone(out.Bytes()[:2]), segment...)
	result = append(result, exif...)
	return append(result, out.Bytes()[2:]...)
}

func TestPrepareRasterNormalizesEveryEXIFOrientation(t *testing.T) {
	expected := [][4]int{{0, 1, 2, 3}, {1, 0, 3, 2}, {3, 2, 1, 0}, {2, 3, 0, 1}, {0, 2, 1, 3}, {2, 0, 3, 1}, {3, 1, 2, 0}, {1, 3, 0, 2}}
	colors := [][3]int{{240, 20, 20}, {20, 240, 20}, {20, 20, 240}, {240, 240, 20}}
	for i, corners := range expected {
		input := orientedJPEG(t, uint16(i+1))
		limits := DefaultLimits()
		limits.LongEdge = 20
		p, err := PrepareRaster(context.Background(), input, "image/jpeg", testBinding(), limits)
		if err != nil {
			t.Fatal(err)
		}
		data, _ := p.Bytes()
		p.Release()
		img, err := jpeg.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		w, h := 20, 10
		if i >= 4 {
			w, h = h, w
		}
		if img.Bounds().Dx() != w || img.Bounds().Dy() != h {
			t.Fatalf("orientation %d: wrong dimensions %v", i+1, img.Bounds())
		}
		for j, want := range corners {
			x := (1 + 2*(j%2)) * w / 4
			y := (1 + 2*(j/2)) * h / 4
			r, g, b, _ := img.At(x, y).RGBA()
			for k, got := range []int{int(r >> 8), int(g >> 8), int(b >> 8)} {
				if d := got - colors[want][k]; d < -35 || d > 35 {
					t.Fatalf("orientation %d corner %d: unexpected color", i+1, j)
				}
			}
		}
		if bytes.Contains(data, []byte("Exif")) {
			t.Fatal("EXIF retained")
		}
	}
}

func TestPrepareRasterDropsPrivateMetadataWithoutChangingInput(t *testing.T) {
	input := orientedJPEG(t, 6)
	private := []byte("GPSLatitude=51.5074 GPSLongitude=-0.1278 Camera=private Filename=private.jpg")
	segment := []byte{255, 254, 0, byte(len(private) + 2)}
	encoded := append(bytes.Clone(input[:2]), segment...)
	encoded = append(encoded, private...)
	encoded = append(encoded, input[2:]...)
	before := bytes.Clone(encoded)
	p, err := PrepareRaster(context.Background(), encoded, "image/jpeg", testBinding(), DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	defer p.Release()
	result, _ := p.Bytes()
	if !bytes.Equal(encoded, before) || bytes.Contains(result, private) || bytes.Contains(result, []byte("Exif")) {
		t.Fatal("input modified or metadata retained")
	}
	for pos := 2; pos < len(result); {
		if result[pos] != 255 {
			t.Fatal("invalid JPEG marker")
		}
		marker := result[pos+1]
		if marker == 0xda {
			break
		}
		if marker == 0xfe || marker >= 0xe1 && marker <= 0xef {
			t.Fatalf("private metadata marker %x retained", marker)
		}
		pos += 2 + int(binary.BigEndian.Uint16(result[pos+2:pos+4]))
	}
}
