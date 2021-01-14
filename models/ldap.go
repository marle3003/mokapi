package models

type LdapServer struct {
	Name    string
	Address string
	Root    *Entry
	Entries map[string]*Entry
}

func (l *LdapServer) Key() string {
	return l.Name
}

type Entry struct {
	Dn         string
	Attributes map[string][]string
}
