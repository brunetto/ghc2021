package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

var files = []string{
	"a_example.in",
	"b_little_bit_of_everything.in",
	"c_many_ingredients.in",
	"d_many_pizzas.in",
	"e_many_teams.in",
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

	pizzaCount := tmp[0]
	teamsCount := tmp[1:]
	teams := []*Team{}

	// populate teams
	for teamTypeID, teamTypeCount := range teamsCount {
		for j := 0; j < teamTypeCount; j++ {
			teams = append(teams, NewTeam(teamTypeID))
		}
	}

	sort.Slice(teams, func(i, j int) bool { return teams[i].peopleCount < teams[j].peopleCount })

	// populate pizzas
	pizzas := make(pizzas, 0, pizzaCount)

	pizzaID := -1
	for s.Scan() {
		pizzaID++

		pizzas = append(pizzas, NewPizza(pizzaID, strings.Split(s.Text(), " ")[1:]))
	}

	// delivery
	pizzaNeedle := 0
	for _, team := range teams {
		leftovers := len(pizzas) - pizzaNeedle

		if team.peopleCount > leftovers {
			// no enough pizzas left for this team
			continue
		}

		team.Delivery(pizzas[pizzaNeedle : pizzaNeedle+team.peopleCount])

		pizzaNeedle += team.peopleCount
	}

	// fmt.Println("LEFTOVER", len(pizzas)-pizzaNeedle, "/", len(pizzas))

	// sort by boredom
	// sort.Slice(teams, func(i, j int) bool { return teams[i].boredom > teams[j].boredom })

	// TODO: try to use leftovers pizzas and waste the previously assigned pizzas
	// for i, t := range teams[0 : len(pizzas)-pizzaNeedle] {
	// 	tt := t.Clone()
	// 	tt.ChangePizza(0, pizzas[pizzaNeedle+1])
	// 	if tt.Score() > t.Score() {
	// 		teams[i] = tt
	// 		pizzaNeedle++
	// 	}
	// }

	// TODO: try to sort by "sfiga" and swap pizzas IF the score improves

	// WRITE OUTPUT SU FILE
	out, err := os.Create(fn + ".out")
	dieIf(err)
	defer out.Close()

	fmt.Fprintln(out, pizzaNeedle)

	totalScore := 0

	for _, team := range teams {
		if !team.delivered {
			continue
		}

		totalScore += team.Score()

		fmt.Fprintf(out, "%v %v\n", team.peopleCount, strings.Join(team.PizzaIDs(), " "))
	}

	// RETURN OUT SUMMARY
	return &stats{
		fn:       fn,
		score:    totalScore,
		duration: time.Since(t0),
	}
}

type Team struct {
	delivered         bool
	teamTypeID        int
	peopleCount       int
	pizzas            pizzas
	uniqueIngredients Set
	ingredients       []string
	boredom           int
}

func NewTeam(teamTypeID int) *Team {
	return &Team{
		teamTypeID:  teamTypeID,
		peopleCount: teamTypeID + 2,
	}
}

func (t *Team) PizzaIDs() []string {
	ids := []string{}

	for _, pizza := range t.pizzas {
		ids = append(ids, strconv.Itoa(pizza.id))
	}

	return ids
}

func (t *Team) ChangePizza(i int, p Pizza) *Team {
	if i > len(t.pizzas) {
		dieIf(errors.New("pizza index greater than pizzas lenght"))
	}

	t.pizzas[i] = p
	t.Recalculate()

	return t
}

func (t *Team) Clone() *Team {
	tt := *t

	return &tt
}

func (t *Team) Delivery(p pizzas) {
	t.delivered = true
	t.pizzas = make(pizzas, t.peopleCount)
	copy(t.pizzas, p) // deep copy

	t.Recalculate()
}

func (t *Team) Recalculate() {
	t.ingredients = make([]string, 0, len(t.pizzas.ingredients()))
	t.ingredients = append(t.ingredients, t.pizzas.ingredients()...)

	t.uniqueIngredients = NewSet().Add(t.ingredients...)
	t.boredom = len(t.ingredients) - len(t.uniqueIngredients)
}

func (t *Team) Score() int {
	return len(t.uniqueIngredients) * len(t.uniqueIngredients)
}

type pizzas []Pizza

func (ps pizzas) ingredients() []string {
	out := []string{}

	for _, p := range ps {
		out = append(out, p.ingredients...)
	}

	return out
}

type Pizza struct {
	id          int
	ingredients []string
}

func NewPizza(id int, ingredients []string) Pizza {
	sort.Strings(ingredients)

	return Pizza{
		id:          id,
		ingredients: ingredients,
	}
}

/////////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////

func main() {
	t0 := time.Now()
	defer func() { log.Println("done in ", time.Since(t0)) }()

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

type Set map[string]nothing
type nothing struct{}

func NewSet() Set {
	return Set{}
}

func (s Set) Add(items ...string) Set {
	if s == nil {
		s = NewSet()
	}

	for _, i := range items {
		s[i] = nothing{}
	}

	return s
}

func (s Set) Contains(item string) bool {
	_, exists := s[item]
	return exists
}

func (s Set) Intersect(other Set) Set {
	var smaller, larger Set

	if len(s) < len(other) {
		smaller = s
		larger = other
	} else {
		smaller = other
		larger = s
	}

	res := make(Set, len(smaller))

	for item := range smaller {
		if larger.Contains(item) {
			res.Add(item)
		}
	}

	return res
}
