package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

var fin *string

type result struct {
	board []uint8
	start int
	end   int
}

func must(err error) {
	if err != nil {
		log.Print(err)
	}
}
func allocate_board(size int) []uint8 {
	board := make([]uint8, size*size)
	for i := range board {
		board[i] = 0
	}
	return board
}

func _play(board []uint8, start int, end int, size int, rowsize int, split int) []uint8 {
	//var a int
	newboard := make([]uint8, size)
	/* for each cell, apply the rules of Life */
	//fmt.Println("play => - board", board, start, end, "size = ", size)
	//fmt.Println("play => - newboard", newboard)

	for y := 0; y < split; y++ {
		for x := 0; x < rowsize; x++ {
			for board[(y*rowsize)+x] == 0 {
				x++
				if x >= rowsize {
					goto RowDone
				}
			}
			a := (board[(y*rowsize)+x] >> 1) & 0x17 // Number of neighbors

			if a == 2 {
				newboard[(y*rowsize)+x] = board[(y*rowsize)+x] & 0x01
			}

			if a == 3 {
				newboard[(y*rowsize)+x] = 1
			}
			if a < 2 {
				newboard[(y*rowsize)+x] = 0
			}
			if a > 3 {
				newboard[(y*rowsize)+x] = 0
			}
		}
	RowDone:
	}

	return newboard

}

/* print the life board */
func print(board []uint8, size int) {
	/* for each row */
	for j := 0; j < size; j++ {
		/* print each column position... */
		for i := 0; i < size; i++ {
			if board[(j*size)+i]&0x01 == 1 {
				fmt.Printf("%c", 'x')
			} else {
				fmt.Printf("%c", ' ')
			}
		}
		/* followed by a carriage return */
		fmt.Printf("\n")
	}
}

func SpaceMap(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, str)
}

func Get(file []byte) (int, int, []uint8) {

	str := string(file) // convert content to a 'string'
	split := strings.SplitAfter(str, "\n")
	ints := strings.SplitAfter(split[0], " ")
	size, err := strconv.Atoi(SpaceMap(ints[0]))
	must(err)
	steps, err := strconv.Atoi(SpaceMap(strings.SplitAfter(ints[1], "\n")[0]))
	must(err)
	if s, err := strconv.Atoi(string(ints[0])); err == nil {
		fmt.Printf("%T, %v\n", s, s)
	}
	board := allocate_board(size)

	for j := 1; j < size; j++ {
		for i := 0; i < size; i++ {
			if byte(split[j][i]) == 'x' {
				board[(j-1)*size+i] = 1
			} else {
				board[(j-1)*size+i] = 0
			}
		}
	}

	// Counting the neighbors
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			count := uint8(0)
			var xoleft, xoright, yoabove, yobelow int
			if y == 0 {
				xoleft = size - 1
			} else {
				xoleft = -1
			}

			if x == 0 {
				yoabove = size * (size - 1)
			} else {
				yoabove = -size
			}

			if y == size-1 {
				xoright = -(size - 1)
			} else {
				xoright = 1
			}
			if x == size-1 {
				yobelow = -size * (size - 1)
			} else {
				yobelow = size
			}
			place := x*size + y

			count += board[place+yoabove+xoleft] & 0x01
			count += board[place+yoabove] & 0x01
			count += board[place+yoabove+xoright] & 0x01
			count += board[place+xoleft] & 0x01
			count += board[place+xoright] & 0x01
			count += board[place+yobelow+xoleft] & 0x01
			count += board[place+yobelow] & 0x01
			count += board[place+yobelow+xoright] & 0x01

			board[place] |= count << 1

		}
	}

	return size, steps, board
}

func sqr(ch chan result, board []uint8, size int, rowsize int, split int, start int, end int, wg *sync.WaitGroup) {
	//fmt.Println("sqr => board - ", board, start, end, "size = ", size)
	newboard := _play(board, start, end, size, rowsize, split)
	ch <- result{newboard, start, end}
	wg.Done()
}

func parallel(board []uint8, size int) []uint8 {
	var wg sync.WaitGroup

	numOfGoRoutines := runtime.NumCPU()
	runtime.GOMAXPROCS(runtime.NumCPU())
	//fmt.Println("Go", numOfGoRoutines)
	var tasksCh chan result
	if size%numOfGoRoutines == 0 {
		tasksCh = make(chan result, numOfGoRoutines)
		split := size / numOfGoRoutines
		wg.Add(numOfGoRoutines)
		for i := 0; i < numOfGoRoutines; i++ {
			go sqr(tasksCh, board[i*split*size:(size+(i*size))*split], split*size, size, split, i*split*size, (size+(i*size))*split, &wg)
		}

	} else {
		tasksCh = make(chan result, numOfGoRoutines+1)
		split := size / numOfGoRoutines
		wg.Add(numOfGoRoutines + 1)
		var i int
		for i = 0; i < numOfGoRoutines; i++ {
			//fmt.Println("range = ", i*split*size, (size+(i*size))*split)
			go sqr(tasksCh, board[i*split*size:(size+(i*size))*split], split*size, size, split, i*split*size, (size+(i*size))*split, &wg)
		}
		go sqr(tasksCh, board[i*split*size:size*size], size*size-i*split*size, size, size%numOfGoRoutines, i*split*size, size*size, &wg)

	}

	wg.Wait()
	//fmt.Println("After WAIT")
	close(tasksCh)
	newboard := allocate_board(size)
	for rows := range tasksCh {
		copy(newboard[rows.start:rows.end], rows.board[:])
	}
	// Counting the neighbors
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			count := uint8(0)
			var xoleft, xoright, yoabove, yobelow int
			if y == 0 {
				xoleft = size - 1
			} else {
				xoleft = -1
			}

			if x == 0 {
				yoabove = size * (size - 1)
			} else {
				yoabove = -size
			}

			if y == size-1 {
				xoright = -(size - 1)
			} else {
				xoright = 1
			}
			if x == size-1 {
				yobelow = -size * (size - 1)
			} else {
				yobelow = size
			}
			place := x*size + y

			count += newboard[place+yoabove+xoleft] & 0x01
			count += newboard[place+yoabove] & 0x01
			count += newboard[place+yoabove+xoright] & 0x01
			count += newboard[place+xoleft] & 0x01
			count += newboard[place+xoright] & 0x01
			count += newboard[place+yobelow+xoleft] & 0x01
			count += newboard[place+yobelow] & 0x01
			count += newboard[place+yobelow+xoright] & 0x01

			newboard[place] |= count << 1

		}
	}

	tasksCh = nil
	//print(newboard, size)
	return newboard
}

func main() {

	file, err := ioutil.ReadAll(os.Stdin)
	must(err)
	size, steps, prev := Get(file)

	var tmp []uint8
	for i := 0; i < steps; i++ {
		n := parallel(prev, size)
		tmp = n
		n = prev
		prev = tmp
	}
	print(prev, size)
}
