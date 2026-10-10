package generator

import (
	"time"

	"github.com/brianvoe/gofakeit/v7"
)

func dates() []*Node {
	nodes := []*Node{
		{
			Name:       "created",
			Attributes: []string{"created", "createdAt", "creationDate"},
			Fake:       fakePastDate(5),
		},
		{
			Name:       "modified",
			Attributes: []string{"modified", "modifiedAt", "modificationDate", "modifyDate"},
			Fake:       fakePastDate(5),
		},
		{
			Name:       "updated",
			Attributes: []string{"updated", "updatedAt", "updateDate"},
			Fake:       fakePastDate(5),
		},
		{
			Name:       "deleted",
			Attributes: []string{"deleted", "deletedAt", "deleteDate"},
			Fake:       fakePastDate(5),
		},
		{
			Name:       "foundation",
			Attributes: []string{"foundationDate"},
			Fake:       fakePastDate(100),
		},
	}
	nodes = append(nodes, timePairs()...)
	return nodes
}

func timePairs() []*Node {
	now := time.Now()
	minDate := now
	minDate.AddDate(-5, 0, 0)

	var nodes []*Node
	for k, v := range commonTimePairs {
		n1 := buildTimePairTree([]string{k}, nil, func(r *Request) (any, error) {
			d, err := fakeDateInPastWithMinYear(r, now.Year()-5)
			if err != nil {
				return nil, err
			}
			r.Context.Values[k] = d
			return d, nil
		})
		nodes = append(nodes, n1)
		for _, v2 := range v {
			md := minDate
			n2 := buildTimePairTree([]string{v2}, []string{k}, func(r *Request) (any, error) {
				if depValue, ok := r.Context.Values[k]; ok {
					t, err := time.Parse(time.RFC3339, depValue.(string))
					if err == nil {
						md = t
					}
				}
				return fakeDateWithYearRange(r, md, now.Year()+5)
			})
			nodes = append(nodes, n2)
		}
	}
	return nodes
}

func buildTimePairTree(tokens []string, dependsOn []string, fake func(r *Request) (any, error)) *Node {
	return &Node{
		Name:       tokens[0],
		Fake:       fake,
		DependsOn:  dependsOn,
		Attributes: tokens[1:],
	}
}

func fakePastDate(pastYears int) func(r *Request) (any, error) {
	year := time.Now().Year()
	return func(r *Request) (any, error) {
		return fakeDateInPastWithMinYear(r, year-pastYears)
	}
}

func fakeDateWithYearRange(r *Request, min time.Time, maxYear int) (any, error) {
	year := gofakeit.IntRange(min.Year(), maxYear)
	minMonth := int(min.Month())
	month := gofakeit.Number(minMonth, 12)
	minDay := min.Day() + 1
	if minDay > maxDayInMonth[month-1] {
		minDay = 1
		if month == 12 {
			month = 1
			if year == min.Year() {
				year++
			}
		} else {
			month += 1
		}
	}
	day := gofakeit.Number(minDay, maxDayInMonth[month-1])

	hour := gofakeit.Number(0, 23)
	minute := gofakeit.Number(0, 59)
	second := gofakeit.Number(0, 59)
	nanosecond := gofakeit.Number(0, 999999999)

	d := time.Date(year, time.Month(month), day, hour, minute, second, nanosecond, time.UTC)
	if r.Schema != nil && r.Schema.Format == "date-time" {
		return d.Format(time.RFC3339), nil
	}

	return d.Format("2006-01-02"), nil
}

var commonTimePairs = map[string][]string{
	"startDate":     {"endDate"},
	"validFrom":     {"invalidFrom", "validUntil"},
	"availableFrom": {"unavailableFrom", "availableUntil"},
	"activeFrom":    {"inactiveFrom", "activeUntil"},
	"enableFrom":    {"disabledFrom", "enableUntil"},
	"publishedAt":   {"unpublishedAt"},
	"publishedFrom": {"unpublishedFrom", "publishedUntil"},
	"releasedAt":    {"unreleasedAt", "deprecatedAt"},
}
