package generator

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"math"
	"strings"
	"time"
)

var personal = []*Node{
	{
		Name: "person",
		Attributes: []string{
			"person",
			"owner",
		},
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
			{
				Name:       "birthday",
				Attributes: []string{"birthday", "birth"},
				Weight:     0.5,
				Fake:       fakeBirthday,
				Children: []*Node{
					{
						Name: "date",
						Fake: fakeBirthday,
					},
				},
			},
			{
				Name:      "title",
				DependsOn: []string{"gender", "sex"},
				Fake:      fakePersonTitle,
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
		Children: []*Node{
			{
				Name: "number",
				Fake: fakePhone,
			},
		},
	},
	{
		Name: "fax",
		Fake: fakePhone,
		Children: []*Node{
			{
				Name: "number",
				Fake: fakePhone,
			},
		},
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
			{
				Name: "type",
				Fake: fakeContactType,
			},
		},
		Fake: fakeContact,
	},
}

func fakePersonName(r *Request) (any, error) {
	var err error

	first, ok := r.ctx.store["firstname"]
	if !ok {
		first, err = fakeFirstname(r)
		if err != nil {
			return nil, err
		}
	}
	last, ok := r.ctx.store["lastname"]
	if !ok {
		last, err = fakeLastname(r)
		if err != nil {
			return nil, err
		}
	}

	return fmt.Sprintf("%s %s", first, last), nil
}

func fakeFirstname(r *Request) (any, error) {
	if v, ok := r.ctx.store["firstname"]; ok {
		return v, nil
	}

	v, err := fakeGender(r)
	if err != nil {
		return nil, err
	}
	sex := v.(string)

	pool := femaleFirstNames
	if sex[0] == 'm' {
		pool = maleFirstNames
	}

	index := gofakeit.Number(0, len(pool)-1)
	firstname := pool[index]
	r.ctx.store["firstname"] = firstname
	return firstname, nil
}

func fakeLastname(r *Request) (any, error) {
	if v, ok := r.ctx.store["lastname"]; ok {
		return v, nil
	}

	index := gofakeit.Number(0, len(lastNames)-1)
	last := lastNames[index]
	r.ctx.store["lastname"] = last
	return last, nil
}

func fakeGender(r *Request) (any, error) {
	if v, ok := r.ctx.store["gender"]; ok {
		return v, nil
	}
	if v, ok := r.ctx.store["sex"]; ok {
		return v, nil
	}

	v := gofakeit.Gender()

	if r.Schema != nil && r.Schema.MaxLength != nil {
		m := *r.Schema.MaxLength
		if m < len(v) {
			v = v[:m]
		}
	}

	r.ctx.store["gender"] = v
	r.ctx.store["sex"] = v

	return v, nil
}

func fakePersonAge(r *Request) (any, error) {
	min, max := getRangeWithDefault(1, 100, r.Schema)
	return gofakeit.Number(int(min), int(max)), nil
}

func fakePerson(r *Request) (any, error) {
	gender, _ := fakeGender(r)
	first, _ := fakeFirstname(r)
	last, _ := fakeLastname(r)
	email, _ := fakeEmail(r)

	return map[string]any{
		"firstname": first,
		"lastname":  last,
		"gender":    gender,
		"email":     email,
	}, nil
}

func fakePhone(r *Request) (any, error) {
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

func fakeContact(r *Request) (any, error) {
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

func fakeBirthday(r *Request) (any, error) {
	return fakeDateInPastWithMinYear(r, 1940)
}

func fakeDateInPastWithMinYear(r *Request, minYear int) (any, error) {
	now := time.Now()

	year := gofakeit.Number(1940, time.Now().Year())
	year = int(math.Max(float64(year), float64(minYear)))
	month := gofakeit.Number(1, 12)
	if year == time.Now().Year() {
		month = gofakeit.Number(1, int(now.Month()))
	}

	day := gofakeit.Number(1, maxDayInMonth[month-1])
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

func fakeContactType(r *Request) (any, error) {
	index := gofakeit.Number(0, len(contactTypes)-1)
	return contactTypes[index], nil
}

func fakePersonTitle(r *Request) (any, error) {
	if v, ok := r.ctx.store["title"]; ok {
		return v, nil
	}

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
	pool := femaleTitles
	if gender[0] == 'm' {
		pool = maleTitles
	}

	index := gofakeit.Number(0, len(pool)-1)
	title := pool[index]
	r.ctx.store["title"] = title
	return title, nil
}

var (
	femaleFirstNames = []string{
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

	maleFirstNames = []string{
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

	lastNames = []string{
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

	contactTypes = []string{
		"billing", "technical", "sales", "support", "legal", "marketing", "general",
	}

	unisexTitles = []string{
		"Mx.", "Dr.", "Prof.", "Rev.",
	}
	maleTitles   = append(unisexTitles, "Mr.")
	femaleTitles = append(unisexTitles, "Miss", "Mrs.", "Ms.")
)
