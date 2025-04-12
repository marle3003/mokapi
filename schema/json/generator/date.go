package generator

import "time"

func dates() []*Node {
	return []*Node{
		{
			Name: "created",
			Fake: fakePastDate(5),
			Children: []*Node{
				{
					Name: "at",
					Fake: fakePastDate(5),
				},
			},
		},
		{
			Name: "creation",
			Children: []*Node{
				{
					Name: "date",
					Fake: fakePastDate(5),
				},
			},
		},
		{
			Name: "modified",
			Fake: fakePastDate(5),
			Children: []*Node{
				{
					Name: "at",
					Fake: fakePastDate(5),
				},
			},
		},
		{
			Name: "modify",
			Children: []*Node{
				{
					Name: "date",
					Fake: fakePastDate(5),
				},
			},
		},
		{
			Name: "updated",
			Fake: fakePastDate(5),
			Children: []*Node{
				{
					Name: "at",
					Fake: fakePastDate(5),
				},
			},
		},
		{
			Name: "update",
			Children: []*Node{
				{
					Name: "date",
					Fake: fakePastDate(5),
				},
			},
		},
		{
			Name: "deleted",
			Fake: fakePastDate(5),
			Children: []*Node{
				{
					Name: "at",
					Fake: fakePastDate(5),
				},
			},
		},
		{
			Name: "delete",
			Children: []*Node{
				{
					Name: "date",
					Fake: fakePastDate(5),
				},
			},
		},
		{
			Name: "foundation",
			Children: []*Node{
				{
					Name: "date",
					Fake: fakePastDate(100),
				},
			},
		},
	}
}

func fakePastDate(pastYears int) func(r *Request) (any, error) {
	year := time.Now().Year()
	return func(r *Request) (any, error) {
		return fakeDateInPastWithMinYear(r, year-5)
	}
}
