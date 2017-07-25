package main

import (
	"fmt"
	"image"
	"os"
	"log"
	"math/rand"
	_ "image/jpeg"
	"image/png"
	"image/color"
)

type Star struct {
	X int
	Y int
	Size int
}


func normalizeImage(img image.Image) [][]float64 {
	var fimg [][]float64;

	bounds := img.Bounds()

	fimg = make([][]float64, bounds.Max.Y)

	ema := 0.
	ema_sf := 0.
	ema_n := 0.


	// convert image to an array of float and take 10% samples to
	// calculate average pixel value this is used to create an
	// estimate of the background "value" (assuming space is
	// mostly black)
	for y := 0 ; y < bounds.Max.Y; y++ {
		fimg[y] = make([]float64, bounds.Max.X);
		for x := 0 ; x < bounds.Max.X; x++ {
			_, g, _, _ := img.At(x, y).RGBA()
			val := float64(g) / 65535

			// take a 10% sample to calculate average background noise value
			if rand.Intn(100) < 30 {
				ema_n++
				ema_sf = 2 / (1 + ema_n)
				ema = (val * ema_sf) + (ema * (1 - ema_sf))
			}
			fimg[y][x] = val
		}
        }

	// we subtract the background and by scaling the EMA by 10% and
	// subtracting the value from every pixel making sure
	// it does drop below 0
	scaled_ema := ema + (ema * .1)
	
	for y := 0 ; y < bounds.Max.Y; y++ {
		for x := 0 ; x < bounds.Max.X; x++ {
			val := fimg[y][x]
			new_val := val - scaled_ema
			if new_val < 0 {
				new_val = 0
			}

			// now rescale
			fimg[y][x] = (new_val / (1 - scaled_ema))
		}
	}

	fmt.Printf("ema = %v scaled_ema = %v\n", ema, scaled_ema)
	return fimg;
}


func DetectStars(img image.Image) []Star {
	var stars []Star;

	bounds := img.Bounds()
	fmt.Printf("Image size: %d x %d\n", bounds.Max.X, bounds.Max.Y)

	fimg := normalizeImage(img)
	saveFloatImage(fimg)
	
	return stars;
}


func saveFloatImage(fimg [][]float64) {
	y_max := len(fimg)
	x_max := len(fimg[0])

	img := image.NewRGBA(image.Rect(0, 0, x_max, y_max))

	for y := 0; y < y_max ; y++ {
		for x := 0 ; x < x_max ; x++ {
			val := fimg[y][x]
			c := uint8(val * 255)
			switch {
			case val > 0.4:
				img.Set(x, y, color.RGBA{255, 0, 0, 255})
			case val > 0.3:
				img.Set(x, y, color.RGBA{0, 255, 0, 255})
			case val > 0.2:
				img.Set(x, y, color.RGBA{0, 0, 255, 255})
			default:
				img.Set(x, y, color.RGBA{c, c, c, 255})
			}
		}
	}

	file, _ := os.Create("out.png")
	defer file.Close()

	png.Encode(file, img);
}



func main() {
	reader, err := os.Open(os.Args[1])
	defer reader.Close()

	if err != nil {
		log.Fatal(err)
	}

	m, _, err := image.Decode(reader)

	if err != nil {
		log.Fatal(err)
	}


	stars := DetectStars(m)
	fmt.Printf("stars: %v\n", stars)
}
