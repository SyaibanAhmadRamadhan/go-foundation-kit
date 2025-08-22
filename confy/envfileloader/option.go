package envfileloader

import "strings"

// options holds all configuration options for the Loader.
// These are set via functional Option helpers like WithFiles, WithTag, etc.
type options struct {
	delimiter string   // Key path delimiter, e.g. "." for "db.host"
	filePaths []string // Candidate file paths to load from
	fileType  string   // Explicit file type ("json", "yaml", "yml"). Auto-detected if empty
	tag       string   // Struct tag to use when unmarshalling (default "koanf")
	flat      bool     // Use flat paths when unmarshalling (default false)

	envPrefix string              // If set, load environment variables with this prefix
	envMapFn  func(string) string // Custom mapping function for env var -> config key, by default: APP_DB__HOST -> db.host

	watch                         bool   // Enable fsnotify watcher for the config file
	callbackOnChangeWhenOnKeyTrue string // Only trigger onChange callback when this key evaluates to true (or empty = always)
}

// Option is a functional option setter for Loader.
type Option func(*options)

// Standard option setters
func WithDelimiter(d string) Option { return func(o *options) { o.delimiter = d } }
func WithFiles(paths ...string) Option {
	return func(o *options) { o.filePaths = append(o.filePaths, paths...) }
}
func WithFileType(t string) Option       { return func(o *options) { o.fileType = strings.ToLower(t) } }
func WithTag(tag string) Option          { return func(o *options) { o.tag = tag } }
func WithFlatPaths(flat bool) Option     { return func(o *options) { o.flat = flat } }
func WithEnvPrefix(prefix string) Option { return func(o *options) { o.envPrefix = prefix } }
func WithEnvMap(fn func(string) string) Option {
	return func(o *options) { o.envMapFn = fn }
}
func WithWatch(enable bool) Option { return func(o *options) { o.watch = enable } }

// WithCallbackOnChangeWhenOnKeyTrue sets a key name.
// The onChange callback will only be fired when this key's value is true (bool).
// If left empty, callback is always fired.
func WithCallbackOnChangeWhenOnKeyTrue(key string) Option {
	return func(o *options) {
		o.callbackOnChangeWhenOnKeyTrue = key
	}
}
