package match

import (
	"fmt"
	"github.com/wesovilabs/goa/advice"
	"github.com/wesovilabs/goa/joinpoint"
)

// Matches struct
type Matches []*Match

// Match struct
type Match struct {
	JoinPoint *joinpoint.JoinPoint
	Advices   map[string]*advice.Advice
}

// GetMatches return the list of existing matches
func GetMatches(joinPoints *joinpoint.JoinPoints, advices *advice.Advices) Matches {
	matches := Matches{}

	for _, f := range joinPoints.List() {
		aspects := make(map[string]*advice.Advice)

		for index, d := range advices.List() {
			if d.Match(f.Path()) {
				aspects[fmt.Sprintf("advice%v", index)] = d
			}
		}

		if len(aspects) > 0 {
			matches = append(matches, &Match{
				JoinPoint: f,
				Advices:   aspects,
			})
		}
	}

	return matches
}
