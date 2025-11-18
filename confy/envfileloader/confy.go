package envfileloader

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/confy"
	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/confy/parser"
	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/confy/provider"
	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/observability"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"
)

// Loader is the main config loader.
// It loads config from file(s), merges env variables,
// unmarshals into a struct, and supports watching for file changes.
type Loader[T any] struct {
	k            *koanf.Koanf      // underlying koanf instance
	cur          atomic.Pointer[T] // holds the current config snapshot
	filepath     string            // active config file path
	fileProvider *provider.File    // file provider used for watching
	parser       koanf.Parser
	opt          options // applied options

	// optional
	envProvider *provider.Env // file provider used for watching
}

// New creates a new Loader with the provided options.
// If watch is enabled, it also attaches a watcher.
// onChangeWatcher is optional (can be nil).
func New[T any](onChangeWatcher func(*T, error), opts ...Option) (*Loader[T], error) {
	o := options{
		delimiter: ".",
		tag:       "koanf",
	}
	for _, fn := range opts {
		fn(&o)
	}
	if len(o.filePaths) == 0 {
		return nil, errors.New("confy: no file candidates; use WithFiles")
	}

	k := koanf.New(o.delimiter)
	path := firstExisting(o.filePaths)
	if path == "" {
		return nil, fmt.Errorf("confy: no config file found in candidates: %v", o.filePaths)
	}
	parser, err := pickParser(o.fileType, path)
	if err != nil {
		return nil, err
	}

	l := &Loader[T]{
		k:            k,
		opt:          o,
		fileProvider: provider.NewFile(path),
		parser:       parser,
		filepath:     path,
	}
	if l.opt.envPrefix != "" {
		if o.envMapFn == nil {
			o.envMapFn = func(s string) string {
				s = strings.TrimPrefix(s, o.envPrefix+"_")
				s = strings.ToLower(s)
				return strings.ReplaceAll(s, "__", ".")
			}
		}
		l.envProvider = provider.NewEnv(o.envPrefix, o.delimiter, o.envMapFn)
	}

	// Initial load
	if err := l.loadOnce(); err != nil {
		return nil, err
	}

	// Attach watcher if requested
	if l.opt.watch {
		if err := l.watch(onChangeWatcher); err != nil {
			return nil, err
		}
	}
	return l, nil
}

// loadOnce loads the config from file and env (if configured),
// and unmarshals into a new struct instance of type T.
func (l *Loader[T]) loadOnce() error {
	if err := l.k.Load(l.fileProvider, l.parser); err != nil {
		return fmt.Errorf("confy: load %s: %w", l.filepath, err)
	}

	if l.opt.envPrefix != "" {
		if err := l.k.Load(l.envProvider, nil); err != nil {
			return fmt.Errorf("confy: load env: %w", err)
		}
	}

	out := new(T)
	if err := l.k.UnmarshalWithConf("", out, koanf.UnmarshalConf{
		Tag:       l.opt.tag,
		FlatPaths: l.opt.flat,
	}); err != nil {
		return fmt.Errorf("confy: unmarshal: %w", err)
	}
	l.cur.Store(out)
	return nil
}

// watch attaches a file watcher that reloads config on change.
// It only triggers the onChange callback if the configured key is true
// (or if no key was specified).
func (l *Loader[T]) watch(onChange func(*T, error)) error {
	if l.fileProvider == nil {
		return errors.New("confy: watch enabled but provider not initialized")
	}

	return l.fileProvider.Watch(func(_ any, werr error) {
		if werr != nil {
			observability.Start(context.Background(), zerolog.ErrorLevel).
				Err(werr).Msgf("failed watcher file: %s", l.filepath)
			if onChange != nil && (l.k.Bool(l.opt.callbackOnChangeWhenOnKeyTrue) || l.opt.callbackOnChangeWhenOnKeyTrue == "") {
				onChange(nil, fmt.Errorf("watch error: %w", werr))
			}
			return
		}

		if err := l.loadOnce(); err != nil {
			observability.Start(context.Background(), zerolog.ErrorLevel).
				Err(err).Msgf("failed load file: %s", l.filepath)
			if onChange != nil && (l.k.Bool(l.opt.callbackOnChangeWhenOnKeyTrue) || l.opt.callbackOnChangeWhenOnKeyTrue == "") {
				onChange(nil, err)
			}
			return
		}

		if onChange != nil && (l.k.Bool(l.opt.callbackOnChangeWhenOnKeyTrue) || l.opt.callbackOnChangeWhenOnKeyTrue == "") {
			onChange(l.Get(), nil)
		}
		confy.Notify()
	})
}

// Get returns the current config snapshot.
func (l *Loader[T]) Get() *T {
	return l.cur.Load()
}

// Reload forces a reload and returns the new snapshot.
func (l *Loader[T]) Reload() (*T, error) {
	if err := l.loadOnce(); err != nil {
		return nil, err
	}
	return l.Get(), nil
}

func (l *Loader[T]) Unwatch() error {
	return l.fileProvider.Unwatch()
}

// pickParser returns a parser based on file type or extension.
func pickParser(fileType, path string) (koanf.Parser, error) {
	t := strings.ToLower(strings.TrimSpace(fileType))
	if t == "" {
		ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
		t = ext
	}
	switch t {
	case "json":
		return parser.NewJson(), nil
	case "yaml", "yml":
		return parser.NewYaml(), nil
	default:
		return nil, fmt.Errorf("confy: unsupported file type %q (path=%s)", t, path)
	}
}

// firstExisting returns the first existing file from candidates.
func firstExisting(candidates []string) string {
	for _, p := range candidates {
		if p == "" {
			continue
		}
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}
