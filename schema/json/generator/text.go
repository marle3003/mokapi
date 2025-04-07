package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"math"
)

func textNodes() []*Node {
	return []*Node{
		{
			Name: "description",
			Fake: fakeDescription,
		},
		{
			Name: "category",
			Fake: fakeCategory,
		},
	}
}

func fakeDescription(r *Request) (interface{}, error) {
	wordCount := 15
	if r.Schema != nil {
		avgWordLength := 5
		minLength := 0
		maxLength := 150
		if r.Schema.MinLength != nil {
			minLength = *r.Schema.MinLength
		}
		if r.Schema.MaxLength != nil {
			maxLength = *r.Schema.MaxLength
		}

		maxWords := int(math.Floor(float64(maxLength) / float64(avgWordLength)))

		for ; maxWords > 0; maxWords-- {
			v := gofakeit.Sentence(maxWords)
			if len(v) >= minLength && len(v) <= maxLength {
				return v, nil
			}
		}

		return nil, NotSupported
	}
	return gofakeit.Sentence(wordCount), nil
}

var (
	category = []string{
		"Technology", "Fashion", "Food", "Travel", "Health", "Education", "Finance", "Entertainment", "Sports", "Automotive", "Home & Garden", "Beauty", "Pets", "Gaming", "Music", "Literature", "Fitness", "Art", "Photography", "Business", "Science", "History", "Cooking", "DIY", "Parenting", "Gardening", "Crafts", "Design", "Architecture", "Film", "Television", "Comedy", "Drama", "Action", "Adventure", "Mystery", "Romance", "Thriller", "Horror", "Fantasy", "Science Fiction", "Biography", "Memoir", "Self-Help", "Religion", "Philosophy", "Politics", "Environment", "Travelogue", "Lifestyle",
	}
	category4 = []string{
		"News", "Tech", "Food", "Life", "Code", "Work", "Data", "Apps", "Game", "Art",
	}
	category5 = []string{
		"Media", "Books", "Tools", "Music", "Space", "Drama", "Cloud", "Style", "Photo", "Games",
	}
	category6 = []string{
		"Travel", "Health", "Movies", "Sports", "Design", "People", "Social", "Finance", "Garden", "Coding",
	}
)

func fakeCategory(r *Request) (interface{}, error) {
	var pool []string
	if r.Schema != nil {
		min := 0
		max := 20
		if r.Schema.MinLength != nil {
			min = *r.Schema.MinLength
		}
		if r.Schema.MaxLength != nil {
			max = *r.Schema.MaxLength
		}

		if min <= 4 && 4 <= max {
			pool = append(pool, category4...)
		}
		if min <= 5 && 5 <= max {
			pool = append(pool, category5...)
		}
		if min <= 6 && 6 <= max {
			pool = append(pool, category6...)
		}
		if min == 0 && max >= 10 {
			pool = append(pool, category...)
		}
	}

	index := gofakeit.Number(0, len(pool)-1)
	return pool[index], nil
}
