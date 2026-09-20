package images

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"

	"github.com/disintegration/imaging"
)

var ErrInvalid = errors.New("invalid image preparation")

func PrepareRaster(ctx context.Context, input []byte, mime string, binding Binding, limits Limits) (*Prepared, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !limits.valid() {
		return nil, ErrInvalid
	}
	for _, value := range []string{binding.Owner, binding.Installation, binding.Asset, binding.SourceDigest} {
		if len(value) == 0 || len(value) > 128 {
			return nil, ErrInvalid
		}
	}
	if len(input) > limits.MaxSourceBytes {
		return nil, ErrLimit
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(input))
	if err != nil {
		return nil, ErrInvalid
	}
	if err := validateContainer(input, mime, format); err != nil {
		return nil, err
	}
	if config.Width <= 0 || config.Height <= 0 || config.Width > limits.MaxDimension || config.Height > limits.MaxDimension || int64(config.Width)*int64(config.Height) > int64(limits.MaxPixels) {
		return nil, ErrLimit
	}
	img, err := imaging.Decode(bytes.NewReader(input), imaging.AutoOrientation(true))
	if err != nil {
		return nil, ErrInvalid
	}
	resized := imaging.Fit(img, limits.LongEdge, limits.LongEdge, imaging.Lanczos)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	opaque := image.NewRGBA(resized.Bounds())
	draw.Draw(opaque, opaque.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	draw.Draw(opaque, opaque.Bounds(), resized, resized.Bounds().Min, draw.Over)
	out := boundedBuffer{limit: (limits.MaxRequestBytes - len("data:image/jpeg;base64,")) / 4 * 3}
	if err := jpeg.Encode(&out, opaque, &jpeg.Options{Quality: 85}); err != nil {
		return nil, ErrLimit
	}
	if err := ctx.Err(); err != nil {
		clear(out.Bytes())
		return nil, err
	}
	sum := sha256.Sum256(out.Bytes())
	policy := fmt.Sprintf("jpeg-white-q85-v1:%d:%d:%d:%d:%d", limits.LongEdge, limits.MaxPixels, limits.MaxDimension, limits.MaxSourceBytes, limits.MaxRequestBytes)
	return &Prepared{state: &preparedState{data: out.Bytes(), metadata: Metadata{Binding: binding, Width: resized.Bounds().Dx(), Height: resized.Bounds().Dy(), MIME: "image/jpeg", Digest: hex.EncodeToString(sum[:]), Policy: policy}}}, nil
}
