package pgxx

import (
	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/databases/pgxx/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Option interface {
	apply(*pgxpool.Config)
}

type optFunc func(*pgxpool.Config)

func (o optFunc) apply(cfg *pgxpool.Config) {
	o(cfg)
}

func WithOtel(opts ...otelpgx.Option) Option {
	return optFunc(func(cfg *pgxpool.Config) {
		opts = append(opts, otelpgx.WithTrimSQLInSpanName())
		cfg.ConnConfig.Tracer = otelpgx.NewTracer(
			opts...,
		)
	})
}
