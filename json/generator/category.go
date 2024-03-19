package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"strings"
)

func Category() *Tree {
	return &Tree{
		Name: "Category",
		Test: func(r *Request) bool {
			last := r.Last()
			name := strings.ToLower(last.Name)
			s := last.Schema
			return name == "category" && (s.IsAnyString() || s.IsAny())
		},
		Fake: func(r *Request) (interface{}, error) {
			index := gofakeit.Number(0, len(category)-1)
			return category[index], nil
		},
	}
}

var category = []string{
	"Technology", "Fashion", "Food", "Travel", "Health", "Education", "Finance", "Entertainment", "Sports", "Automotive", "Home & Garden", "Beauty", "Pets", "Gaming", "Music", "Literature", "Fitness", "Art", "Photography", "Business", "Science", "History", "Cooking", "DIY", "Parenting", "Gardening", "Crafts", "Design", "Architecture", "Film", "Television", "Comedy", "Drama", "Action", "Adventure", "Mystery", "Romance", "Thriller", "Horror", "Fantasy", "Science Fiction", "Biography", "Memoir", "Self-Help", "Religion", "Philosophy", "Politics", "Environment", "Travelogue", "Lifestyle"}
