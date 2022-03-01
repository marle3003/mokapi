package metrics

type LdapMetrics struct {
	RequestCounter      *Counter
	RequestErrorCounter *Counter
}

func NewLdap() *LdapMetrics {
	return &LdapMetrics{
		RequestCounter:      NewCounter("ldap.requests.total"),
		RequestErrorCounter: NewCounter("ldap.requests.total.errors"),
	}
}
