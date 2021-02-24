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

func run(fn string) *stats {
	t0 := time.Now()

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
	return &stats{
		fn:       fn,
		duration: time.Since(t0),
	}
}

/////////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////

func main() {
	t0 := time.Now()

	wkrs := sync.WaitGroup{}
	rdc := sync.WaitGroup{}
	out := make(chan *stats, len(files))

	// print result as they arrive, concurrency safe
	rdc.Add(1)
	go func() {
		defer rdc.Done()

		total := &stats{}
		for res := range out {
			total.Add(res)

			fmt.Println(res)
		}

		fmt.Printf("\n-------------------------------" +
			"--------------------------------------------" +
			"--------------------------------------------" +
			"-----------------------------------------\n\n")
		fmt.Println(total)
	}()

	// run tasks
	for _, fn := range files {
		wkrs.Add(1)

		go func(wg *sync.WaitGroup, fn string, out chan *stats) {
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
	isAggregation bool
	score         int
	maxScore      int
	fn            string
	duration      time.Duration
	details       interface{}
}

func (s *stats) String() string {
	fn := s.fn
	if s.isAggregation {
		fn = "aggregated"
	}

	perc := 0.
	if s.maxScore != 0 {
		perc = 100 * float64(s.maxScore-s.score) / float64(s.maxScore)
	}

	return fmt.Sprintf("file: %40v | score: %12v | max score: %12v | difference: %12v | perc. missing: %3.2f%% | duration: %15v | %v",
		fn, s.score, s.maxScore, s.maxScore-s.score, perc, s.duration, s.details)
}

func (s *stats) Add(s1 *stats) {
	s.isAggregation = true
	s.score += s1.score
	s.maxScore += s1.maxScore
	s.duration = time.Duration(s.duration.Nanoseconds()+s.duration.Nanoseconds()) * time.Nanosecond
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
