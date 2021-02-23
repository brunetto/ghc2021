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

func run(fn string) stats {
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
	teams := []*team{}

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
	return stats{
		fn:       fn,
		score:    totalScore,
		duration: time.Since(t0),
	}
}

type team struct {
	delivered         bool
	teamTypeID        int
	peopleCount       int
	pizzas            pizzas
	uniqueIngredients Set
	ingredients       []string
}

func NewTeam(teamTypeID int) *team {
	return &team{
		teamTypeID:        teamTypeID,
		peopleCount:       teamTypeID + 2,
		ingredients:       make([]string, 0, teamTypeID+2),
		uniqueIngredients: NewSet(),
	}
}

func (t *team) PizzaIDs() []string {
	ids := []string{}

	for _, pizza := range t.pizzas {
		ids = append(ids, strconv.Itoa(pizza.id))
	}

	return ids
}

func (t *team) Delivery(pizzas pizzas) {
	t.delivered = true
	copy(t.pizzas, pizzas) // deep copy
	t.ingredients = append(t.ingredients, pizzas.ingredients()...)
	t.uniqueIngredients.Add(t.ingredients...)
}

func (t *team) Score() int {
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

func main() {
	t0 := time.Now()
	defer func() { log.Println("done in ", time.Since(t0)) }()

	wkrs := sync.WaitGroup{}
	rdc := sync.WaitGroup{}
	out := make(chan stats, len(files))

	// print result as they arrive, concurrency safe
	rdc.Add(1)
	go func() {
		defer rdc.Done()

		var (
			sumP, sumT    int
			totalDuration int64
		)
		for res := range out {
			sumP += res.score
			sumT += res.maxScore
			totalDuration += res.duration.Nanoseconds()

			fmt.Printf("file: %v, score: %v, max score: %v, difference: %v, duration: %v\n", res.fn, res.score, res.maxScore, res.maxScore-res.score, res.duration)
		}

		fmt.Println("total")
		fmt.Printf("score: %v, max score: %v, difference: %v, perc. missing: %f%%, duration %v: \n",
			sumP, sumT, sumT-sumP, 100*float64(sumT-sumP)/float64(sumT), time.Duration(totalDuration)*time.Nanosecond)
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

}

type stats struct {
	score    int
	maxScore int
	fn       string
	duration time.Duration
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
