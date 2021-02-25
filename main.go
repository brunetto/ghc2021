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
	"a.txt",
	"b.txt",
	"c.txt",
	"d.txt",
	"e.txt",
	"f.txt",
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
	tmp := lineToIntSlice(s.Text())

	// sim desription
	simDurationSecs := tmp[0]
	intersectionsCount := tmp[1]
	streetCount := tmp[2]
	carCount := tmp[3]
	bonus := tmp[4]

	// init intersections
	intersections := make(Intersections, intersectionsCount)
	for i := range intersections {
		intersections[i] = &Intersection{ID: i}
	}

	// read streets
	streets := Streets{}
	for i := 0; i < streetCount; i++ {
		if !s.Scan() {
			dieIf(errors.New("failed reading streets"))
		}

		street := NewStreet(s.Text())
		streets[street.Name] = street

		// populate intersections
		intersections[street.BeginID].AddOut(street)
		intersections[street.EndID].AddIn(&street)
	}

	// read cars
	cars := make(Cars, 0, carCount)
	for i := 0; i < carCount; i++ {
		if !s.Scan() {
			dieIf(errors.New("failed reading cars"))
		}

		cars = append(cars, NewCar(s.Text(), streets))
	}
	//
	//
	//
	//
	// DO THINGS

	for _, car := range cars {
		for _, street := range car.Path {
			intersections[street.EndID].AddCarToTheTransits(street.Name)
		}
	}

	for _, in := range intersections {
		in.SwitchGreenProportionalPerc(simDurationSecs)
	}

	//
	//
	//
	//
	//
	// WRITE OUTPUT SU FILE
	out, err := os.Create(fn + ".out")
	dieIf(err)
	defer out.Close()

	_, err = out.WriteString(intersections.Print())
	dieIf(err)

	// RETURN OUT SUMMARY
	return &stats{
		fn:       fn,
		duration: time.Since(t0),
		score:    0,
		maxScore: cars.MaxScore(bonus, simDurationSecs),
	}
}

type Streets map[string]Street
type Street struct {
	BeginID  int
	EndID    int
	Name     string
	Length   int
	Transits int
}

func NewStreet(line string) Street {
	s := Street{}

	tmp := strings.Split(line, " ")
	begin, err := strconv.Atoi(tmp[0])
	dieIf(err)
	end, err := strconv.Atoi(tmp[1])
	dieIf(err)

	s.BeginID, s.EndID = begin, end
	s.Name = tmp[2]
	lenght, err := strconv.Atoi(tmp[3])
	dieIf(err)

	s.Length = lenght

	return s
}

type Cars []Car

func (cs Cars) MaxScore(bonus, simDurationSecs int) int {
	var sc int

	for _, c := range cs {
		sc += c.MaxScore(bonus, simDurationSecs)
	}

	return sc
}

type Car struct {
	StreetCount int
	Path        []Street
}

func (c Car) MaxScore(bonus, simDurationSecs int) int {
	var length int

	for _, street := range c.Path[1:] {
		length += street.Length
	}

	return bonus + (simDurationSecs - length)
}

// func (c Car) WhereAmI(t int) Position {
// 	return Position{}
// }

// type Position struct {
// 	StreetName     string
// 	StreetPosition string
// }

type Intersections []*Intersection

func (is *Intersections) SwitchFirstIfPresent(simDurationSecs int) {
	for _, i := range *is {
		if len(i.In) == 0 {
			continue
		}

		i.Schedule = append(i.Schedule, NewGreen(i.In[0].Name, simDurationSecs, 1))
	}
}

func (is *Intersections) SwitchAllOneSec(simDurationSecs int) {
	for _, i := range *is {
		for _, s := range i.In {
			i.Schedule = append(i.Schedule, NewGreen(s.Name, simDurationSecs, 1))
		}
	}
}

func (is *Intersections) SwitchAllNSec(simDurationSecs, n int) {
	for _, i := range *is {
		for _, s := range i.In {
			i.Schedule = append(i.Schedule, NewGreen(s.Name, simDurationSecs, 1))
		}
	}
}

func (in *Intersection) SwitchGreenProportional(simDurationSecs int) {
	for _, s := range in.In {
		in.Schedule = append(in.Schedule, NewGreen(s.Name, simDurationSecs, s.Transits))
	}
}

func (in *Intersection) SwitchGreenProportionalPerc(simDurationSecs int) {
	var totalTransits int
	for _, s := range in.In {
		totalTransits += s.Transits
	}

	minPerc := 0

	for _, s := range in.In {
		// perc := 100 * (float64(s.Transits) / float64(totalTransits))
		perc := float64(simDurationSecs) * (float64(s.Transits) / float64(totalTransits))
		if perc < 1 {
			perc = 1
		}

		v := int(perc)
		if v < minPerc || minPerc == 0 {
			minPerc = v
		}

		in.Schedule = append(in.Schedule, NewGreen(s.Name, simDurationSecs, int(perc)))
	}

	for _, g := range in.Schedule {
		g.Duration /= minPerc
		if g.Duration < 1 {
			g.Duration = 1
		}
	}
}

func (is *Intersections) Print() string {
	out := ""

	// print scheduled intersec count
	ints := 0
	for _, i := range *is {
		if len(i.Schedule) == 0 {
			continue
		}

		ints++ // scheduled intersections count

		out += fmt.Sprintf("%v\n", i.ID)
		out += fmt.Sprintf("%v\n", len(i.Schedule))

		for _, g := range i.Schedule {
			out += fmt.Sprintf("%v %v\n", g.Street, g.Duration)
		}
	}

	// scheduled intersection count
	out = fmt.Sprintf("%v\n", ints) + out

	return out
}

type Intersection struct {
	ID int
	In []*Street
	// Out      []*Street
	Schedule Greens
}

func (i *Intersection) AddIn(s *Street) {
	i.In = append(i.In, s)
}
func (i *Intersection) AddOut(s Street) {
	// i.Out = append(i.Out, s)
}

func (i *Intersection) AddCarToTheTransits(street string) {
	idx, exists := IsInStreet(i.In, street)
	if !exists {
		dieIf(errors.New("whoah"))
	}

	i.In[idx].Transits++

}

func (in *Intersection) SwitchGreenFor(street string, simDurationSecs, length int) {
	_, is := IsIn(in.Schedule, street)
	if is {
		return
	}

	in.Schedule = append(in.Schedule, NewGreen(street, simDurationSecs, int(length)))
}

// func (in *Intersection) AddCarToQueue(car Car) {
// 	in.Queue = append(in.Queue, car)
// }

func IsInStreet(list []*Street, street string) (int, bool) {
	for i, s := range list {
		if s.Name == street {
			return i, true
		}
	}

	return 0, false
}

func IsIn(list Greens, street string) (int, bool) {
	for i, g := range list {
		if g.Street == street {
			return i, true
		}
	}

	return 0, false
}

func NewGreen(street string, simDurationSecs int, duration int) *Green {
	if duration >= simDurationSecs || duration < 1 {
		// log.Println("duration aaaaaaaa", duration, simDurationSecs)
		// duration = duration / int(3)
		// if duration < 1 {
		// 	duration = 1
		// }
		duration = 1
	}

	return &Green{
		Street:   street,
		Duration: duration,
	}
}

type Greens []*Green
type Green struct {
	Street   string
	Duration int
}

func NewCar(line string, streets Streets) Car {
	c := Car{}

	tmp := strings.Split(line, " ")
	sc, err := strconv.Atoi(tmp[0])
	dieIf(err)

	c.StreetCount = sc

	for _, streetName := range tmp[1:] {
		c.Path = append(c.Path, streets[streetName])
	}

	return c
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
