// Package png allows for loading png images and applying
// image flitering effects on them.
package png

import (
	"image/color"
)

// Grayscale applies a grayscale filtering effect to the image
func (img *ImageTask) Grayscale() {

	// Bounds returns defines the dimensions of the image. Always
	// use the bounds Min and Max fields to get out the width
	// and height for the image
	bounds := img.Out.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			//Returns the pixel (i.e., RGBA) value at a (x,y) position
			// Note: These get returned as int32 so based on the math you'll
			// be performing you'll need to do a conversion to float64(..)
			r, g, b, a := img.In.At(x, y).RGBA()

			//Note: The values for r,g,b,a for this assignment will range between [0, 65535].
			//For certain computations (i.e., convolution) the values might fall outside this
			// range so you need to clamp them between those values.
			greyC := clamp(float64(r+g+b) / 3)

			//Note: The values need to be stored back as uint16 (I know weird..but there's valid reasons
			// for this that I won't get into right now).
			img.Out.Set(x, y, color.RGBA64{greyC, greyC, greyC, uint16(a)})
		}
	}
}

// Sharpen applies a sharpen effect to the image
func (img *ImageTask) Sharpen() {
	kernel := [9]float64{0, -1, 0, -1, 5, -1, 0, -1, 0}
	img.convolve(kernel)
}

// EdgeDetection applies an edge detection effect to the image
func (img *ImageTask) EdgeDetection() {
	kernel := [9]float64{-1, -1, -1, -1, 8, -1, -1, -1, -1}
	img.convolve(kernel)
}

// Blur applies a blur effect to the image
func (img *ImageTask) Blur() {
	kernel := [9]float64{1 / 9.0, 1 / 9, 1 / 9.0, 1 / 9.0, 1 / 9.0, 1 / 9.0, 1 / 9.0, 1 / 9.0, 1 / 9.0}
	img.convolve(kernel)
}

// convolve performs a convolution on the input image using the given 3*3 kernel
func (img *ImageTask) convolve(kernel [9]float64) {
	// Get image bounds
	bounds := img.Out.Bounds()
	// Loop over every pixel in the image
	for y := bounds.Min.Y; y <= bounds.Max.Y; y++ {
		for x := bounds.Min.X; x <= bounds.Max.X; x++ {
			_, _, _, a := img.In.At(x, y).RGBA()
			var rSum, gSum, bSum float64
			for i := -1; i <= 1; i++ {
				for j := -1; j <= 1; j++ {
					//if x+i >= bounds.Min.X && x+i <= bounds.Max.X && y+j >= bounds.Min.Y && y+j <= bounds.Max.Y {
					// If the pixel in bounds
					r, g, b, _ := img.In.At(x+i, y+j).RGBA() // get rgb value for the pixel
					kx, ky := -i+1, -j+1                     // index for the kernel
					kVal := kernel[kx*3+ky]
					rSum += kVal * float64(r)
					gSum += kVal * float64(g)
					bSum += kVal * float64(b)
					//}
				}
			}
			// Set color for the pixel in output image
			img.Out.Set(x, y, color.RGBA64{clamp(rSum), clamp(gSum), clamp(bSum), uint16(a)})
		}
	}
}
