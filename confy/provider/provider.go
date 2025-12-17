package provider

type Provider interface {
	Name() string

	// ReadBytes returns the entire configuration as raw []bytes to be parsed.
	// with a Parser.
	ReadBytes() ([]byte, error)

	// Read returns the parsed configuration as a nested map[string]interface{}.
	Read() (map[string]any, error)

	Parser() Parser
}

type Parser interface {
	Unmarshal([]byte) (map[string]any, error)
	Marshal(map[string]any) ([]byte, error)
}
