package smtp

type Event struct {
	From string
	To   string
}

type Login struct {
	Username  string
	Password  string
	Anonymous bool
}
