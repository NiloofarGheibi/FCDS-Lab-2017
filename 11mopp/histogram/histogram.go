package main

import (
	"bufio"
	"bytes"
	"errors"
	//"flag"
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
	//"math"
	"os"
	"runtime"
	"strconv"
	"sync"
)

var fin *string
var CPU_NUM int
var d decoder
var img image.Image
var n float32
var _image []PPMImage
var size int
var done bool

type message struct {
	j   uint32
	k   uint32
	l   uint32
	x   int
	his float32
}

type result struct {
	his float32
	x   int
}

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

	runtime.GOMAXPROCS(runtime.NumCPU())
	CPU_NUM = runtime.NumCPU()

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
			//fmt.Println("red = ", r, "green = ", g, "blue = ", b)
			r = r & 0x00FF
			g = g & 0x00FF
			b = b & 0x00FF
			//fmt.Println("Before red = ", r, "green = ", g, "blue = ", b)
			image[counter].red = (r >> 6) | 0
			image[counter].green = (g >> 6) | 0
			image[counter].blue = (b >> 6) | 0
			//fmt.Println("After red = ", image[counter].red, "green = ", image[counter].green, "blue = ", image[counter].blue)
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

func worker(tasksCh <-chan message, wg *sync.WaitGroup, resCh chan<- result) {
	defer wg.Done()
	for {
		task, ok := <-tasksCh
		if !ok {
			return
		}
		if done == false {
			Adjust()
			done = true
		}
		count := float32(0)

		for i := 0; i < size; i++ {
			if _image[i].red == task.j && _image[i].green == task.k && _image[i].blue == task.l {
				count++
			}
		}
		//_______________ END ________________
		resCh <- result{count, task.x}
	}
}

func pool(wg *sync.WaitGroup, workers, tasks int, h *[64]float32) {
	tasksCh := make(chan message)
	resCh := make(chan result, 64)
	x := 0
	//Adjust()
	for i := 0; i < workers; i++ {
		go worker(tasksCh, wg, resCh)
	}

	for j := uint32(0); j <= 3; j++ {
		for k := uint32(0); k <= 3; k++ {
			for l := uint32(0); l <= 3; l++ {
				tasksCh <- message{j, k, l, x, 0.0}
				x++
			}
		}
	}

	for i := 0; i < tasks; i++ {
		val := <-resCh
		h[val.x] = val.his / n
	}
	close(tasksCh)
}

func Adjust() {
	bounds := img.Bounds()
	_image = make([]PPMImage, size)
	counter := 0
	//log.Println(bounds.Min.Y, bounds.Max.Y, bounds.Min.X, bounds.Max.X)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			//fmt.Println("red = ", r, "green = ", g, "blue = ", b)
			r = r & 0x00FF
			g = g & 0x00FF
			b = b & 0x00FF
			_image[counter].red = (r >> 6) | 0
			_image[counter].green = (g >> 6) | 0
			_image[counter].blue = (b >> 6) | 0
			//fmt.Println("After red = ", image[counter].red, "green = ", image[counter].green, "blue = ", image[counter].blue)
			counter++
		}
	}
}
func main() {

	d, img, _ = Decode(os.Stdin)
	n = float32(d.width * d.height)
	var hist [64]float32
	size = d.width * d.height
	done = false
	//hist := Histogram(d, img)
	var wg sync.WaitGroup
	wg.Add(CPU_NUM)
	go pool(&wg, CPU_NUM, 64, &hist)
	wg.Wait()

	//hist := Test(d, img)
	for i := 0; i < 64; i++ {
		fmt.Printf("%0.3f ", hist[i])
	}
	fmt.Println("\n")

}
