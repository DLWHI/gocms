package image

import (
	"image"
	"math"

	"golang.org/x/image/draw"
)

func Contain(src image.Image, boxWidth int, boxHeight int) image.Image {
	srcBounds := src.Bounds()
	srcW := float64(srcBounds.Dx())
	srcH := float64(srcBounds.Dy())

	scaleW := float64(boxWidth) / srcW
	scaleH := float64(boxHeight) / srcH
	scale := math.Min(scaleW, scaleH)

	scaledW := int(math.Round(srcW * scale))
	scaledH := int(math.Round(srcH * scale))

	dstBounds := image.Rect(0, 0, boxWidth, boxHeight)
	dst := image.NewRGBA(dstBounds)

	offsetX := (boxWidth - scaledW) / 2
	offsetY := (boxHeight - scaledH) / 2

	scaledRect := image.Rect(offsetX, offsetY, offsetX+scaledW, offsetY+scaledH)
	draw.ApproxBiLinear.Scale(dst, scaledRect, src, srcBounds, draw.Over, nil)

	return dst
}
