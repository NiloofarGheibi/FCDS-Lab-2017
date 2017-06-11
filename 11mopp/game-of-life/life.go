package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"unicode"
)

var fin *string

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

func _Cell_Turn_On(board *[]uint8, row int, col int, size int) {
	var xoleft, xoright, yoabove, yobelow int

	// OFFSET CALCULATING

	if row == 0 {
		xoleft = size - 1
	} else {
		xoleft = -1
	}

	if col == 0 {
		yoabove = size * (size - 1)
	} else {
		yoabove = -size
	}

	if row == size-1 {
		xoright = -(size - 1)
	} else {
		xoright = 1
	}
	if col == size-1 {
		yobelow = -size * (size - 1)
	} else {
		yobelow = size
	}
	place := col*size + row
	(*board)[place] |= 0x01 // first bit = state

	(*board)[place+yoabove+xoleft] += 2
	(*board)[place+yoabove] += 2
	(*board)[place+yoabove+xoright] += 2
	(*board)[place+xoleft] += 2
	(*board)[place+xoright] += 2
	(*board)[place+yobelow+xoleft] += 2
	(*board)[place+yobelow] += 2
	(*board)[place+yobelow+xoright] += 2

}

func _Cell_Turn_Off(board *[]uint8, row int, col int, size int) {
	var xoleft, xoright, yoabove, yobelow int

	// OFFSET CALCULATING

	if row == 0 {
		xoleft = size - 1
	} else {
		xoleft = -1
	}

	if col == 0 {
		yoabove = size * (size - 1)
	} else {
		yoabove = -size
	}

	if row == size-1 {
		xoright = -(size - 1)
	} else {
		xoright = 1
	}
	if col == size-1 {
		yobelow = -size * (size - 1)
	} else {
		yobelow = size
	}

	place := col*size + row

	(*board)[place] &= 0xFE // first bit = state

	(*board)[place+yoabove+xoleft] -= 2
	(*board)[place+yoabove] -= 2
	(*board)[place+yoabove+xoright] -= 2
	(*board)[place+xoleft] -= 2
	(*board)[place+xoright] -= 2
	(*board)[place+yobelow+xoleft] -= 2
	(*board)[place+yobelow] -= 2
	(*board)[place+yobelow+xoright] -= 2
}

func Next_Generation(board []uint8, size int) []uint8 {

	for y := 0; y < size; y++ {
		x := 0
		//fmt.Println("before do place = ", place)
		for x < size {

			// Off and no neighbors
			for board[(y*size)+x] == 0 {
				fmt.Println("pass")
				x++
				if x >= size {
					goto RowDone
				}

			}
			// 0 0 0 0   0 0 0 0
			a := (board[(y*size)+x] >> 1) & 0x17 // Number of neighbors
			fmt.Println("Neigbours = ", a, "Value = ", board[(y*size)+x]&0x01)

			if board[(y*size)+x]&0x01 == 1 {
				if a != 2 && a != 3 {
					_Cell_Turn_Off(&board, y, x, size)
				}

			} else {
				if a == 3 {
					_Cell_Turn_On(&board, y, x, size)
				}
			}
			x++
		}
	RowDone:
	}
	return board
}

func play(board []uint8, size int) []uint8 {
	//var a int
	newboard := allocate_board(size)
	/* for each cell, apply the rules of Life */
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {

			// if board[(y*size)+x] == 0 {
			// 	continue
			// } else {

			for board[(y*size)+x] == 0 {
				x++
				if x >= size {
					goto RowDone
				}
			}
			a := (board[(y*size)+x] >> 1) & 0x17 // Number of neighbors
			if a == 2 {
				newboard[(y*size)+x] = board[(y*size)+x] & 0x01
			}

			if a == 3 {
				newboard[(y*size)+x] = 1

			}
			if a < 2 {
				newboard[(y*size)+x] = 0
			}
			if a > 3 {
				newboard[(y*size)+x] = 0
			}

		}
	RowDone:
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

func main() {

	file, err := ioutil.ReadAll(os.Stdin)
	must(err)
	size, steps, prev := Get(file)

	var tmp []uint8
	for i := 0; i < steps; i++ {
		//n := Next_Generation(prev, size)
		n := play(prev, size)
		//fmt.Println(i, "____________")
		//print(n, size)
		//print(n, size)
		tmp = n
		n = prev
		prev = tmp
	}
	print(prev, size)
}
