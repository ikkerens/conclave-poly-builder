package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/png"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/furstenheim/ConcaveHull"
)

var path string

func init() {
	flag.StringVar(&path, "path", "", "URL to directory to build an area map of")
}

func main() {
	flag.Parse()

	dir, err := ioutil.ReadDir(path)
	if err != nil {
		panic(err)
	}

	for _, fi := range dir {
		name := filepath.Join(path, fi.Name())
		fmt.Println("Opening", name)
		f, err := os.Open(name)
		if err != nil {
			panic(err)
		}
		if str, err := processImage(f); err != nil {
			panic(err)
		} else {
			fmt.Println(fi.Name(), ": ", str)
		}
		_ = f.Close()
	}
}

func processImage(f *os.File) (string, error) {
	img, _, err := image.Decode(f)
	if err != nil {
		return "", err
	}

	// First, we determine the outline. Any pixel that has at least one transparent pixel around it is an outline
	outline := make([]float64, 0)
	for y := 0; y < img.Bounds().Max.Y; y++ {
		for x := 0; x < img.Bounds().Max.X; x++ {
			if isTransparent(img, x, y) {
				continue
			}

			isOutline := false
			switch {
			case x > 0 && isTransparent(img, x-1, y):
				fallthrough
			case y > 0 && isTransparent(img, x, y-1):
				fallthrough
			case x <= img.Bounds().Max.X && isTransparent(img, x+1, y):
				fallthrough
			case y <= img.Bounds().Max.Y && isTransparent(img, x, y+1):
				isOutline = true
			}

			if isOutline {
				outline = append(outline, float64(x), float64(y))
			}
		}
	}
	fmt.Println("Found outline of", len(outline)/2, "pixels")

	// Apply the concave hull algorithm
	hull := ConcaveHull.ComputeWithOptions(ConcaveHull.FlatPoints(outline), &ConcaveHull.Options{
		Seglength: 1,
	})

	// Cut off the first coordinate, as it loops back
	hull = hull[2:]

	fmt.Println(len(hull)/2, "pixels remaining after filtering")

	// Now we convert that into a string
	numbers := make([]string, 0, len(outline))
	for _, number := range hull {
		numbers = append(numbers, strconv.Itoa(int(number)))
	}
	return strings.Join(numbers, ","), nil
}

func isTransparent(img image.Image, x, y int) bool {
	_, _, _, a := img.At(x, y).RGBA()
	return a == 0
}
