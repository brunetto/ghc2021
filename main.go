package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

var files = []string{
	"a_example.txt",
	"b_read_on.txt",
	"c_incunabula.txt",
	"d_tough_choices.txt",
	"e_so_many_books.txt",
	"f_libraries_of_the_world.txt",
}

func run(fn string) stats {
	// READ DATA
	in, err := os.Open(fn)
	dieIf(err)
	defer in.Close()

	s := bufio.NewScanner(in)
	bf := []byte{}
	s.Buffer(bf, 5e6)

	if !s.Scan() {
		dieIf(errors.New("failed on first line"))
	}
	_ /*tmp :*/ = lineToIntSlice(s.Text())

	// CALCULATE MAX SCORE
	//
	//
	//
	//
	//

	// DO THINGS
	//
	//
	//
	//
	//
	//
	//

	// WRITE OUTPUT SU FILE
	out, err := os.Create(fn + ".out")
	dieIf(err)
	defer out.Close()

	//
	//
	//
	//
	//
	//
	//
	//
	//

	// RETURN OUT SUMMARY
	return stats{}
}

func main() {
	t0 := time.Now()

	wkrs := sync.WaitGroup{}
	rdc := sync.WaitGroup{}
	out := make(chan stats, len(files))

	// print result as they arrive, concurrency safe
	rdc.Add(1)
	go func() {
		defer rdc.Done()

		var sumP, sumT int
		for res := range out {
			sumP += res.score
			sumT += res.maxScore
			fmt.Printf("file: %v, score: %v, max score: %v, difference: %v\n", res.fn, res.score, res.maxScore, res.maxScore-res.score)
		}

		fmt.Println("total")
		fmt.Printf("score: %v, max score: %v, difference: %v, perc. missing: %f%%: \n",
			sumP, sumT, sumT-sumP, 100*float64(sumT-sumP)/float64(sumT))
	}()

	// run tasks
	for _, fn := range files {
		wkrs.Add(1)

		go func(wg *sync.WaitGroup, fn string, out chan stats) {
			defer wg.Done()

			out <- run(fn)
		}(&wkrs, fn, out)
	}

	wkrs.Wait()
	close(out)

	rdc.Wait()

	fmt.Println()
	log.Println("done in ", time.Since(t0))
}

type stats struct {
	score    int
	maxScore int
	fn       string
}

func lineToIntSlice(line string) []int {
	fields := strings.Fields(line)
	out := make([]int, 0, len(fields))
	for _, field := range fields {
		num, err := strconv.ParseInt(field, 10, 64)
		dieIf(err)

		out = append(out, int(num))
	}
	return out
}

func dieIf(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
