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
	"time"
	"unicode"
)

var size int
var steps int
var CPU_NUM int
var CHUNK int
var prev [][]int
var next [][]int
var newboard [][]int

type message struct {
	row    []int
	row_id int
}
type result struct {
	row    []int
	row_id int
}

func allocate_board(size int) [][]int {
	board := make([][]int, size)
	for i := range board {
		board[i] = make([]int, size)
	}
	return board
}

func worker(tasksCh <-chan message, wg *sync.WaitGroup, results chan<- result) {
	defer wg.Done()
	for {
		task, ok := <-tasksCh
		if !ok {
			return
		}
		//newboard := make([]int, CHUNK)
		count := make([]int, size)
		for i := 0; i < size; i++ {
			count[i] = adjacent_to(prev, size, task.row_id, i)
		}
		for j := 0; j < size; j++ {

			// for task.row[j] == 0 && count[j] == 0 {
			// 	j++
			// 	if j >= size {
			// 		goto RowDone
			// 	}

			// }
			if count[j] == 2 {
				newboard[task.row_id][j] = task.row[j]
			}

			if count[j] == 3 {
				newboard[task.row_id][j] = 1
			}
			if count[j] < 2 {
				newboard[task.row_id][j] = 0
			}
			if count[j] > 3 {
				newboard[task.row_id][j] = 0
			}
			//}
		}
		//RowDone:
		//_______________ END ________________
		//fmt.Println("processing task")
		results <- result{newboard[task.row_id], task.row_id}
	}
}

func pool(wg *sync.WaitGroup, workers, tasks int, resultCh chan result) {
	tasksCh := make(chan message)
	//resultCh := make(chan result)

	for i := 0; i < workers; i++ {
		go worker(tasksCh, wg, resultCh)
	}

	for j := 0; j < tasks; j++ {
		//log.Println("prev = ", prev[j])
		tasksCh <- message{prev[j], j}
	}

	for a := 0; a < tasks; a++ {
		//combine(a, <-resultCh)
		<-resultCh
	}
	close(tasksCh)
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	CPU_NUM = runtime.NumCPU()
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
		}
	}

	count = count - board[i][j]

	return count
}
func count(board [][]int) [][]int {
	c := allocate_board(size)
	// updating the count
	for j := 0; j < size; j++ {
		for i := 0; i < size; i++ {
			c[j][i] = adjacent_to(board, size, j, i)
		}
	}
	return c
}

func combine(id int, rows result) {
	fmt.Println("copy", id, rows.row_id)
	time.Sleep(100)
	copy(next[rows.row_id], rows.row)
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
	//log.Println("1 line : ", split[0])
	ints := strings.SplitAfter(split[0], " ")
	//log.Println("first line data: ", ints[0], ints[1])
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
	for j := 1; j <= size; j++ {
		/* get a string */
		/* copy the string to the life board */
		//fmt.Println("j = ", j, "=> ", split[j])
		for i := 0; i < size; i++ {
			if byte(split[j][i]) == 'x' {
				board[i][j-1] = 1
				//log.Println("x")
			} else {
				board[i][j-1] = 0
				//log.Println("_")
			}

		}
		//fscanf(f,"\n");
	}

	return size, steps, board
}

func must(err error) {
	if err != nil {
		log.Print(err)
	}
}

func main() {
	file, err := ioutil.ReadAll(os.Stdin)
	must(err)
	size, steps, prev = Get(file)
	newboard = allocate_board(size)
	//log.Println("size = ", size, "steps = ", steps)

	for k := 0; k < steps; k++ {
		var wg sync.WaitGroup
		resultCh := make(chan result, size)
		wg.Add(CPU_NUM)
		go pool(&wg, CPU_NUM, size, resultCh)
		wg.Wait()
		//close(resultCh)
		//fmt.Println("Step = ", k)
		//print(newboard, size)
		tmp := newboard
		newboard = prev
		prev = tmp
		//_count = count(board)
	}
	print(prev, size)
}
