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

	for j := 1; j <= size; j++ {
		for i := 0; i < size; i++ {
			if byte(split[j][i]) == 'x' {
				board[(j-1)*size+i] = 1
			} else {
				board[(j-1)*size+i] = 0
			}
		}
	}

	for j := 0; j < size; j++ {
		for i := 0; i < size; i++ {
			var sk, ek, sl, el int
			k := 0
			l := 0
			count := uint8(0)

			if j > 0 {
				sk = j - 1
			} else {
				sk = j
			}

			if j+1 < size {
				ek = j + 1
			} else {
				ek = i
			}

			if i > 0 {
				sl = i - 1
			} else {
				sl = i
			}

			if i+1 < size {
				el = i + 1
			} else {
				el = i
			}

			for k = sk; k <= ek; k++ {

				for l = sl; l <= el; l++ {
					count += board[k*size+l] & 0x01
					//fmt.Printf(" value of count %v \n", count)
				}
			}

			count = count - board[j*size+i]&0x01
			board[j*size+i] |= count << 1

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
		var i int
		for i = 0; i < numOfGoRoutines; i++ {
			//fmt.Println("for", i, "Split = ", i*split*size, "->", (size+(i*size))*split)
			go sqr(tasksCh, board[i*split*size:(size+(i*size))*split], split*size, size, split, i*split*size, (size+(i*size))*split, &wg)
		}
		//go sqr(tasksCh, board[i*split*size:size*size], size*size-i*split*size, size, size%numOfGoRoutines, i*split*size, size*size, &wg)

	} else {
		tasksCh = make(chan result, numOfGoRoutines+1)
		split := size / numOfGoRoutines
		wg.Add(numOfGoRoutines + 1)
		var i int
		for i = 0; i < numOfGoRoutines; i++ {
			//fmt.Println("range = ", i*split*size, (size+(i*size))*split)
			go sqr(tasksCh, board[i*split*size:(size+(i*size))*split], split*size, size, split, i*split*size, (size+(i*size))*split, &wg)
		}
		//fmt.Println("range = ", i*split*size, size*size)
		go sqr(tasksCh, board[i*split*size:size*size], size*size-i*split*size, size, size%numOfGoRoutines, i*split*size, size*size, &wg)

	}

	wg.Wait()
	//fmt.Println("After WAIT")
	close(tasksCh)
	newboard := allocate_board(size)
	for rows := range tasksCh {
		copy(newboard[rows.start:rows.end], rows.board[:])
	}

	for j := 0; j < size; j++ {
		for i := 0; i < size; i++ {
			var sk, ek, sl, el int
			k := 0
			l := 0
			count := uint8(0)

			if j > 0 {
				sk = j - 1
			} else {
				sk = j
			}

			if j+1 < size {
				ek = j + 1
			} else {
				ek = i
			}

			if i > 0 {
				sl = i - 1
			} else {
				sl = i
			}

			if i+1 < size {
				el = i + 1
			} else {
				el = i
			}

			for k = sk; k <= ek; k++ {

				for l = sl; l <= el; l++ {
					count += newboard[k*size+l] & 0x01
					//fmt.Printf(" value of count %v \n", count)
				}
			}

			count = count - newboard[j*size+i]&0x01
			newboard[j*size+i] |= count << 1

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
