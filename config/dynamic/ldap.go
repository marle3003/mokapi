package dynamic

type Ldap struct {
	Info    LdapInfo
	Server  map[string]interface{}
	Entries []map[string]interface{}
}

type LdapInfo struct {
	Name string `yaml:"title"`
}
