package dynamic

type Provider interface {
	ProvideService(channel chan<- ConfigMessage)
}
