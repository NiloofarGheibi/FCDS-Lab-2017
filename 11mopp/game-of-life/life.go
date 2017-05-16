/*
 * The Game of Life
 *
 * a cell is born, if it has exactly three neighbours
 * a cell dies of loneliness, if it has less than two neighbours
 * a cell dies of overcrowding, if it has more than three neighbours
 * a cell survives to the next generation, if it does not die of loneliness
 * or overcrowding
 *
 * In this version, a 2D array of ints is used.  A 1 cell is on, a 0 cell is off.
 * The game plays a number of steps (given by the input), printing to the screen each time.  'x' printed
 * means on, space means off.
 *
 */
package main

import (
	"flag"
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
func allocate_board(size int) [][]int {
	board := make([][]int, size)
	for i := range board {
		board[i] = make([]int, size)
	}
	return board
}

/* return the number of on cells adjacent to the i,j cell */
func adjacent_to(board [][]int, size int, i int, j int) int {
	var sk, ek, sl, el int
	k := 0
	l := 0
	count := 0

	if i > 0 {
		sk = i - 1
	} else {
		sk = i
	}

	if i+1 < size {
		ek = i + 1
	} else {
		ek = i
	}

	if j > 0 {
		sl = j - 1
	} else {
		sl = j
	}

	if j+1 < size {
		el = j + 1
	} else {
		el = j
	}

	for k = sk; k <= ek; k++ {

		for l = sl; l <= el; l++ {
			count += board[k][l]
			//fmt.Printf(" value of count %v \n", count)
		}
	}

	count = count - board[i][j]

	return count
}

func play(board [][]int, size int) [][]int {
	var a int
	newboard := allocate_board(size)
	/* for each cell, apply the rules of Life */
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			a = adjacent_to(board, size, i, j)
			if a == 2 {
				newboard[i][j] = board[i][j]
			}

			if a == 3 {
				newboard[i][j] = 1
			}
			if a < 2 {
				newboard[i][j] = 0
			}
			if a > 3 {
				newboard[i][j] = 0
			}
		}
	}

	return newboard

}

/* print the life board */
func print(board [][]int, size int) {
	/* for each row */
	for j := 0; j < size; j++ {
		/* print each column position... */
		for i := 0; i < size; i++ {
			if board[i][j] == 1 {
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

func Read(fin string) []byte {

	pwd, _ := os.Getwd()
	file, err := ioutil.ReadFile(pwd + "/" + fin) // pass the file name
	must(err)

	return file
}

func Get(file []byte) (int, int, [][]int) {

	str := string(file) // convert content to a 'string'
	split := strings.SplitAfter(str, "\n")

	// string to parse
	//log.Println("first line : ", split[0])
	ints := strings.SplitAfter(split[0], " ")
	//log.Println("first line : ", ints[0], ints[1])
	size, err := strconv.Atoi(SpaceMap(ints[0]))
	must(err)
	steps, err := strconv.Atoi(SpaceMap(strings.SplitAfter(ints[1], "\n")[0]))
	must(err)
	//v := "10"
	if s, err := strconv.Atoi(string(ints[0])); err == nil {
		fmt.Printf("%T, %v\n", s, s)
	}
	//log.Println("size: , steps: ", size, steps)
	board := allocate_board(size)
	for j := 1; j < size; j++ {
		/* get a string */
		/* copy the string to the life board */
		for i := 0; i < size; i++ {
			if byte(split[j][i]) == 'x' {
				board[i][j] = 1
				//log.Println("x")
			} else {
				board[i][j] = 0
				//log.Println("_")
			}

		}
		//fscanf(f,"\n");
	}

	return size, steps, board
}

func init() {

	fin = flag.String("in", "judge.in", "input file")
	flag.Parse()

}
func main() {

	file := Read(*fin)
	//fmt.Println(file)
	size, steps, prev := Get(file)
	//next := allocate_board(size)
	var tmp, next [][]int
	for i := 0; i < steps; i++ {
		next = play(prev, size)
		//fmt.Printf("%v ----------\n", i)
		//print(next, size)
		tmp = next
		next = prev
		prev = tmp
	}

	//fmt.Printf("LAST \n")
	print(prev, size)
}
