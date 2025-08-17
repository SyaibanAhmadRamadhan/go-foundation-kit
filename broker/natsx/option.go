package natsx

type Option interface {
	apply(*clientConfig)
}

type optFunc func(*clientConfig)

func (o optFunc) apply(c *clientConfig) {
	o(c)
}
