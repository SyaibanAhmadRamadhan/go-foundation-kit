package confy

import (
	"time"

	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/confy/provider"
)

// options holds all configuration options for the Loader.
// These are set via functional Option helpers like WithFiles, WithTag, etc.
type options struct {
	delimiter string // Key path delimiter, e.g. "." for "db.host"
	tag       string // Struct tag to use when unmarshalling (default "koanf")
	flat      bool   // Use flat paths when unmarshalling (default false)

	refreshInterval time.Duration // Interval for periodic refreshes (if applicable)

	onRefresh func(map[string]any) // Callback function for interval refreshes

	providers []provider.Provider // all providers used
}

// Option is a functional option setter for Loader.
type Option func(*options)

// Standard option setters
func WithDelimiter(d string) Option { return func(o *options) { o.delimiter = d } }

// WithTag sets the struct tag to be used when unmarshalling config into structs.
// Default is "koanf".
func WithTag(tag string) Option { return func(o *options) { o.tag = tag } }

// WithFlatPaths
// If this is set to true, instead of unmarshalling nested structures
// based on the key path, keys are taken literally to unmarshal into
// a flat struct. For example:
// ```
//
//	type MyStuff struct {
//		Child1Name string `koanf:"parent1.child1.name"`
//		Child2Name string `koanf:"parent2.child2.name"`
//		Type       string `koanf:"json"`
//	}
//
// ```
func WithFlatPaths(flat bool) Option { return func(o *options) { o.flat = flat } }

// WithIntervalRefresh sets the interval for periodic refreshes.
// If not set, no periodic refreshes will be performed.
func WithIntervalRefresh(d time.Duration) Option {
	return func(o *options) {
		o.refreshInterval = d
	}
}

// WithCallbackIntervalRefresh sets a callback function to be called
// on each interval refresh with the latest configuration map.
func WithCallbackIntervalRefresh(fn func(map[string]any)) Option {
	return func(o *options) {
		o.onRefresh = fn
	}
}

func SetProviders(providers []provider.Provider) Option {
	return func(o *options) {
		o.providers = providers
	}
}

func SetProvider(p provider.Provider) Option {
	return func(o *options) {
		if len(o.providers) == 0 {
			o.providers = make([]provider.Provider, 0)
		}

		o.providers = append(o.providers, p)
	}
}
