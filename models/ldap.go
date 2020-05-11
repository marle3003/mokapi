package models

type LdapServer struct {
	Listen  string
	Root    *Entry
	Entries []*Entry
}

type Entry struct {
	Dn         string
	Attributes map[string][]string
}
