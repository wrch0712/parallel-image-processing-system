// Package png allows for loading png images and applying
// image flitering effects on them
package png

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

// The ImageTask represents a structure for working with PNG images.
// From Professor Samuels: You are allowed to update this and change it as you wish!
type ImageTask struct {
	In             *image.RGBA64   //The original pixels before applying the effect
	Out            *image.RGBA64   //The updated pixels after applying the effect
	Bounds         image.Rectangle //The size of the image
	Effects        []interface{}   // Effects need to apply
	OutputFilePath string          //
}

//
// Public functions
//

// Load returns an Image that was loaded based on the filePath parameter
// From Professor Samuels:  You are allowed to modify and update this as you wish
func Load(filePath string) (*ImageTask, error) {

	inReader, err := os.Open(filePath)

	if err != nil {
		return nil, err
	}
	defer inReader.Close()

	inOrig, err := png.Decode(inReader)

	if err != nil {
		return nil, err
	}

	bounds := inOrig.Bounds()

	outImg := image.NewRGBA64(bounds)
	inImg := image.NewRGBA64(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := inOrig.At(x, y).RGBA()
			inImg.Set(x, y, color.RGBA64{uint16(r), uint16(g), uint16(b), uint16(a)})
		}
	}
	task := &ImageTask{}
	task.In = inImg
	task.Out = outImg
	task.Bounds = bounds
	return task, nil
}

// Save saves the image to the given file
// From Professor Samuels:  You are allowed to modify and update this as you wish
func (img *ImageTask) Save() error {

	outWriter, err := os.Create(img.OutputFilePath)
	if err != nil {
		return err
	}
	defer outWriter.Close()

	err = png.Encode(outWriter, img.Out)
	if err != nil {
		return err
	}
	return nil
}

// clamp will clamp the comp parameter to zero if it is less than zero or to 65535 if the comp parameter
// is greater than 65535.
func clamp(comp float64) uint16 {
	return uint16(math.Min(65535, math.Max(0, comp)))
}
