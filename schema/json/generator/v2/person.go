package v2

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"strings"
)

func personal() []*Node {
	return []*Node{
		{
			Name: "person",
			Fake: fakePerson,
			Children: []*Node{
				{
					Name:      "name",
					DependsOn: []string{"firstname", "lastname"},
					Fake:      fakePersonName,
				},
				{
					Name:      "firstname",
					DependsOn: []string{"gender", "sex"},
					Weight:    1.0,
					Fake:      fakeFirstname,
				},
				{
					Name:   "lastname",
					Weight: 1.0,
					Fake:   fakeLastname,
				},
				{
					Name:   "gender",
					Weight: 1.0,
					Fake:   fakeGender,
				},
				{
					Name:   "sex",
					Weight: 1.0,
					Fake:   fakeGender,
				},
				{
					Name:   "age",
					Weight: 0.5,
					Fake:   fakePersonAge,
				},
			},
		},
		{
			Name: "gender",
			Fake: fakeGender,
		},
		{
			Name: "sex",
			Fake: fakeGender,
		},
		{
			Name: "phone",
			Fake: fakePhone,
		},
		{
			Name:      "contact",
			DependsOn: []string{"firstname", "lastname"},
			Children: []*Node{
				{
					Name:   "phone",
					Weight: 0.5,
					Fake:   fakePhone,
				},
				{
					Name:   "email",
					Weight: 0.5,
					Fake:   fakeEmail,
				},
			},
			Fake: fakeContact,
		},
	}

}

func fakePersonName(r *Request) (interface{}, error) {
	var err error

	first := r.ctx.store["firstname"]
	if first == nil {
		first, err = fakeFirstname(r)
		if err != nil {
			return nil, err
		}
	}
	last := r.ctx.store["lastname"]
	if last == nil {
		last, err = fakeLastname(r)
		if err != nil {
			return nil, err
		}
	}

	return fmt.Sprintf("%s %s", first, last), nil
}

func fakeFirstname(r *Request) (interface{}, error) {
	var gender string
	if v, ok := r.ctx.store["gender"]; ok {
		gender = v.(string)
	}
	if gender == "" {
		v, err := fakeGender(r)
		if err != nil {
			return nil, err
		}
		gender = v.(string)
	}
	pool := femaleFirstNames
	if gender[0] == 'm' {
		pool = maleFirstNames
	}

	index := gofakeit.Number(0, len(pool)-1)
	firstname := pool[index]
	r.ctx.store["firstname"] = firstname
	return firstname, nil
}

func fakeLastname(r *Request) (interface{}, error) {
	index := gofakeit.Number(0, len(lastNames)-1)
	last := lastNames[index]
	r.ctx.store["lastname"] = last
	return last, nil
}

func fakeGender(r *Request) (interface{}, error) {
	v := gofakeit.Gender()

	if r.Schema != nil && r.Schema.MaxLength != nil {
		m := *r.Schema.MaxLength
		if m < len(v) {
			return v[:m], nil
		}
	}
	r.ctx.store["gender"] = v
	return v, nil
}

func fakePersonAge(r *Request) (interface{}, error) {
	min, max := getRangeWithDefault(1, 100, r.Schema)
	return gofakeit.Number(int(min), int(max)), nil
}

func fakePerson(r *Request) (interface{}, error) {
	r.ctx.Snapshot()
	defer r.ctx.Restore()

	gender, _ := fakeGender(r)
	first, _ := fakeFirstname(r)
	last, _ := fakeLastname(r)
	email, _ := fakeEmail(r)

	return map[string]interface{}{
		"firstname": first,
		"lastname":  last,
		"gender":    gender,
		"email":     email,
	}, nil
}

func fakePhone(r *Request) (interface{}, error) {
	s := r.Schema

	countryCode := gofakeit.IntRange(1, 999)
	countryCodeLen := len(fmt.Sprintf("%v", countryCode))
	max := 15 - countryCodeLen
	min := 4
	if s != nil && s.MinLength != nil {
		min = *s.MinLength - countryCodeLen - 1
	}
	if s != nil && s.MaxLength != nil {
		max = *s.MaxLength - countryCodeLen - 1
	}
	nationalCodeLen := gofakeit.IntRange(min, max)
	return fmt.Sprintf("+%v%v", countryCode, gofakeit.Numerify(strings.Repeat("#", nationalCodeLen))), nil
}

func fakeContact(r *Request) (interface{}, error) {
	phone, err := fakePhone(r)
	if err != nil {
		return nil, err
	}
	email, err := fakeEmail(r)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"phone": phone,
		"email": email,
	}, nil
}

var femaleFirstNames = []string{
	"Emma", "Olivia", "Ava", "Sophia", "Isabella",
	"Mia", "Charlotte", "Amelia", "Evelyn", "Abigail",
	"Emily", "Ella", "Elizabeth", "Camila", "Luna",
	"Sofia", "Avery", "Mila", "Aria", "Scarlett",
	"Penelope", "Layla", "Chloe", "Victoria", "Madison",
	"Eleanor", "Grace", "Nora", "Riley", "Zoey",
	"Hannah", "Hazel", "Lily", "Ellie", "Violet",
	"Lillian", "Zoe", "Stella", "Aurora", "Natalie",
	"Emilia", "Everly", "Leah", "Aubrey", "Willow",
	"Addison", "Lucy", "Audrey", "Bella", "Claire",
}

var maleFirstNames = []string{
	"Liam", "Noah", "Oliver", "Elijah", "James",
	"William", "Benjamin", "Lucas", "Henry", "Alexander",
	"Mason", "Michael", "Ethan", "Daniel", "Jacob",
	"Logan", "Jackson", "Levi", "Sebastian", "Mateo",
	"Jack", "Owen", "Theodore", "Aiden", "Samuel",
	"Joseph", "John", "David", "Wyatt", "Matthew",
	"Luke", "Asher", "Carter", "Julian", "Grayson",
	"Leo", "Jayden", "Gabriel", "Isaac", "Lincoln",
	"Anthony", "Hudson", "Dylan", "Ezra", "Thomas",
	"Charles", "Christopher", "Jaxon", "Maverick", "Josiah",
}

var lastNames = []string{
	"Smith", "Johnson", "Williams", "Brown", "Jones",
	"Garcia", "Miller", "Davis", "Rodriguez", "Martinez",
	"Hernandez", "Lopez", "Gonzalez", "Wilson", "Anderson",
	"Thomas", "Taylor", "Moore", "Jackson", "Martin",
	"Lee", "Perez", "Thompson", "White", "Harris",
	"Sanchez", "Clark", "Ramirez", "Lewis", "Robinson",
	"Walker", "Young", "Allen", "King", "Wright",
	"Scott", "Torres", "Nguyen", "Hill", "Flores",
	"Green", "Adams", "Nelson", "Baker", "Hall",
	"Rivera", "Campbell", "Mitchell", "Carter", "Roberts",
}
