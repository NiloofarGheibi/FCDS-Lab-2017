package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

var fin *string

type message struct {
	count  []int
	row    []int
	row_id int
}

type result struct {
	row    []int
	row_id int
}

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

func play(board [][]int, size int, a [][]int) [][]int {
	//var a int
	newboard := allocate_board(size)
	/* for each cell, apply the rules of Life */
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {

			if board[i][j] == 0 && a[i][j] == 0 {
				continue
			} else {

				if a[i][j] == 2 {
					newboard[i][j] = board[i][j]
				}

				if a[i][j] == 3 {
					newboard[i][j] = 1
				}
				if a[i][j] < 2 {
					newboard[i][j] = 0
				}
				if a[i][j] > 3 {
					newboard[i][j] = 0
				}
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
func Count(board [][]int, size int) [][]int {
	c := allocate_board(size)
	// updating the count
	for j := 0; j < size; j++ {
		for i := 0; i < size; i++ {
			c[j][i] = adjacent_to(board, size, j, i)
		}
	}
	return c
}

func worker(tasksCh <-chan message, wg *sync.WaitGroup, size int, next chan result) {
	defer wg.Done()
	for {
		newboard := make([]int, size)
		task, ok := <-tasksCh
		if !ok {
			return
		}
		//mt.Println("processing task", task)
		// ____________ Processing ____________
		for j := 0; j < size; j++ {

			if task.row[j] == 0 && task.count[j] == 0 {
				continue
			} else {
				if task.count[j] == 2 {
					newboard[j] = task.row[j]
				}

				if task.count[j] == 3 {
					newboard[j] = 1
				}
				if task.count[j] < 2 {
					newboard[j] = 0
				}
				if task.count[j] > 3 {
					newboard[j] = 0
				}
			}
		}
		//_______________ END ________________
		next <- result{newboard, task.row_id}
	}
}

func pool(wg *sync.WaitGroup, workers, tasks int, board [][]int, a [][]int, next chan result) {
	tasksCh := make(chan message)

	for i := 0; i < workers; i++ {
		go worker(tasksCh, wg, tasks, next)
	}

	for i := 0; i < tasks; i++ {
		tasksCh <- message{a[i], board[i], i}
	}

	close(tasksCh)

}

func main() {
	file, err := ioutil.ReadAll(os.Stdin)
	must(err)
	size, steps, prev := Get(file)
	var tmp, n [][]int
	for i := 0; i < steps; i++ {
		var wg sync.WaitGroup
		next := make(chan result, size)
		c := Count(prev, size)
		wg.Add(size)
		go pool(&wg, size, size, prev, c, next)
		wg.Wait()
		close(next)
		// Data for each row is ready
		//fmt.Println("After WAIT")
		n = allocate_board(size)
		for i := range next {
			//fmt.Println("Row = ", i.row, "id = ", i.row_id)
			n[i.row_id] = i.row
		}

		//print(n, size)
		tmp = n
		n = prev
		prev = tmp
	}
	print(prev, size)

}
