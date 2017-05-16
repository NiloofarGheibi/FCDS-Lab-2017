package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

var fin *string
var (
	errBadHeader   = errors.New("ppm: invalid header")
	errNotEnough   = errors.New("ppm: not enough image data")
	errUnsupported = errors.New("ppm: unsupported format (maxVal != 255)")
)

// decoder is the type used to decode a PPM file.
type decoder struct {
	br *bufio.Reader

	// from header
	magicNumber string
	width       int
	height      int
	maxVal      int // 255
}

type PPMImage struct {
	red   uint32
	green uint32
	blue  uint32
}

func must(err error) {
	if err != nil {
		log.Print(err)
	}
}

func Decode(r io.Reader) (decoder, image.Image, error) {
	var d decoder
	if dd, img, err := d.decode(r); err != nil {
		return d, nil, err
	} else {
		return dd, img, nil
	}
}

func (d *decoder) decode(r io.Reader) (decoder, image.Image, error) {
	d.br = bufio.NewReader(r)
	var err error

	// decode header
	err = d.decodeHeader()
	if err != nil {
		return *d, nil, err
	}
	// decode image
	pixel := make([]byte, 3)

	img := image.NewRGBA(image.Rect(0, 0, d.width, d.height))

	for y := 0; y < d.height; y++ {
		for x := 0; x < d.width; x++ {
			_, err = io.ReadFull(d.br, pixel)
			if err != nil {
				return *d, nil, errNotEnough
			}
			//fmt.Println(pixel[0])
			//fmt.Println(pixel[1])
			//fmt.Println(pixel[2])
			img.SetRGBA(x, y, color.RGBA{pixel[0], pixel[1], pixel[2], 0xff})
		}
	}
	return *d, img, nil
}

func (d *decoder) decodeHeader() error {
	var err error
	var b byte
	header := make([]byte, 0)

	comment := false
	for fields := 0; fields < 4; {
		b, _ = d.br.ReadByte()
		if b == '#' {
			comment = true
		} else if !comment {
			header = append(header, b)
		}
		if comment && b == '\n' {
			comment = false
		} else if !comment && (b == ' ' || b == '\n' || b == '\t') {
			fields++
		}
	}
	headerFields := bytes.Fields(header)

	d.magicNumber = string(headerFields[0])
	if d.magicNumber != "P6" {
		return errBadHeader
	}
	d.width, err = strconv.Atoi(string(headerFields[1]))
	if err != nil {
		return errBadHeader
	}
	d.height, err = strconv.Atoi(string(headerFields[2]))
	if err != nil {
		return errBadHeader
	}

	d.maxVal, err = strconv.Atoi(string(headerFields[3]))
	if err != nil {
		return errBadHeader
	} else if d.maxVal != 255 {
		return errUnsupported
	}
	return nil
}

func init() {

	fin = flag.String("in", "judge.in", "input file")
	flag.Parse()

}

func Read() *os.File {

	pwd, _ := os.Getwd()
	//file, err := ioutil.ReadFile(pwd + "/" + *fin) // pass the file name
	file, err := os.Open(pwd + "/" + *fin)
	must(err)

	return file

}

func Test(d decoder, img image.Image) [64]float32 {
	bounds := img.Bounds()
	size := d.width * d.height
	var h [64]float32
	image := make([]PPMImage, size)
	counter := 0

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			r = r >> 8
			g = g >> 8
			b = b >> 8

			image[counter].red = (r >> 6) | 0
			image[counter].green = (g >> 6) | 0
			image[counter].blue = (b >> 6) | 0
			counter++
		}
	}
	n := float32(d.width * d.height)

	count := float32(0)
	x := 0

	for j := uint32(0); j <= 3; j++ {
		for k := uint32(0); k <= 3; k++ {
			for l := uint32(0); l <= 3; l++ {
				for i := 0; i < size; i++ {
					if image[i].red == j && image[i].green == k && image[i].blue == l {
						count++
					}
				}
				h[x] = count / n
				x++
				count = 0

			}
		}
	}

	return h
}

func Parse(file []byte) {

	str := string(file) // convert content to a 'string'
	split := strings.SplitAfter(str, "\n")

	// string to parse
	format := split[0]
	log.Println("format : ", format)
	ints := strings.SplitAfter(split[1], " ")
	log.Println("x y : ", ints[0], ints[1])
	max := split[2]
	log.Println(" maximum : ", max)
}

func main() {

	//Parse(Read())

	d, img, _ := Decode(Read())
	//fmt.Println(d)
	//hist := Histogram(d, img)
	hist := Test(d, img)
	for i := 0; i < 64; i++ {
		fmt.Printf("%0.3f ", hist[i])
	}
	fmt.Println("\n")

}
