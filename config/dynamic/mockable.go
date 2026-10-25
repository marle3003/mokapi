package dynamic

type Mockable interface {
	// Mockability checks whether the spec contains anything that Mokapi
	// can handle. An empty "reason" means that nothing should be reported
	Mockability() Mockability
}

type MockabilityLevel int

const (
	MockabilityOK MockabilityLevel = iota
	MockabilityInfo
	MockabilityWarn
)

type Mockability struct {
	Level  MockabilityLevel
	Reason string
}
