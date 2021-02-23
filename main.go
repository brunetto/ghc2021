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
			teams = append(teams, &team{
				teamTypeID:  teamTypeID,
				peopleCount: teamTypeID + 2,
			})
		}
	}

	sort.Slice(teams, func(i, j int) bool { return teams[i].peopleCount < teams[j].peopleCount })

	// populate pizzas
	pizzas := make([]pizza, 0, pizzaCount)

	pizzaID := -1
	for s.Scan() {
		pizzaID++

		ings := strings.Split(s.Text(), " ")[1:]
		sort.Strings(ings)

		pizzas = append(pizzas, pizza{
			id:          pizzaID,
			ingredients: ings,
		})
	}

	pizzaNeedle := 0
	for _, team := range teams {
		leftovers := len(pizzas) - pizzaNeedle

		if team.peopleCount > leftovers {
			// no enough pizzas left for this team
			continue
		}

		teamPizzas := pizzas[pizzaNeedle : pizzaNeedle+team.peopleCount]

		team.pizzas = teamPizzas // deep copy

		ingredients := []string{}

		for _, p := range teamPizzas {
			ingredients = append(ingredients, p.ingredients...)
		}

		team.uniqueIngredients = NewSet().Add(ingredients...)
		team.ingredients = append(team.ingredients, ingredients...)

		pizzaNeedle += team.peopleCount
	}

	// WRITE OUTPUT SU FILE
	out, err := os.Create(fn + ".out")
	dieIf(err)
	defer out.Close()

	fmt.Fprintln(out, pizzaNeedle)

	totalScore := 0

	for _, team := range teams {
		pizzaIDs := []string{}

		for _, pizza := range team.pizzas {
			pizzaIDs = append(pizzaIDs, strconv.Itoa(pizza.id))
		}

		if len(team.pizzas) == 0 {
			continue
		}

		totalScore += team.Score()

		fmt.Fprintf(out, "%v %v\n", team.peopleCount, strings.Join(pizzaIDs, " "))
	}

	// RETURN OUT SUMMARY
	return stats{
		fn:    fn,
		score: totalScore,
	}
}

type team struct {
	teamTypeID        int
	peopleCount       int
	pizzas            pizzas
	uniqueIngredients Set
	ingredients       []string
}

func (t *team) Score() int {
	return len(t.uniqueIngredients) * len(t.uniqueIngredients)
}

type pizzas []pizza
type pizza struct {
	id          int
	ingredients []string
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
