package main

import (
	"fmt"

	"github.com/SyaibanAhmadRamadhan/go-foundation-kit/confy/envfileloader"
)

type Conf struct {
	Name string `env:"name"`
	Env  string `env:"env"`
	DB   ConfDB `env:"db"`
}

type ConfDB struct {
	Username string `env:"username"`
	Password string `env:"password"`
}

var loader *envfileloader.Loader[Conf]

func GetConf() *Conf {
	if loader == nil {
		return &Conf{}
	}

	return loader.Get()
}

func loadConfig() {
	const cfgPath = "env.json"

	writeJSONAtomic(cfgPath, `{
  "name": "my-app",
  "env":  "dev",
  "db":   { "username": "devuser", "password": "devpass" }
}`)

	fnForOnChangeWatch := func(c *Conf, err error) {
		if err != nil {
			fmt.Println("[WATCH] error:", err)
			return
		}
		if c == nil {
			fmt.Println("[WATCH] nil config (unexpected)")
			return
		}
		fmt.Printf("[WATCH] reloaded: name=%s env=%s db.user=%s\n", c.Name, c.Env, c.DB.Username)
	}
	var err error
	loader, err = envfileloader.New(fnForOnChangeWatch,
		// config
		envfileloader.WithDelimiter("."),
		envfileloader.WithFileType("json"),
		envfileloader.WithTag("env"),
		envfileloader.WithFiles(cfgPath),
		envfileloader.WithWatch(true),
	)
	if err != nil {
		panic(err)
	}

	cfg := loader.Get()

	fmt.Printf("[BOOT ] snapshot: name=%s env=%s db.user=%s\n", cfg.Name, cfg.Env, cfg.DB.Username)

}

func loadConfigWithOutCallback() {
	const cfgPath = "env.json"

	writeJSONAtomic(cfgPath, `{
  "name": "my-app",
  "env":  "dev",
  "db":   { "username": "devuser", "password": "devpass" }
}`)

	var err error
	loader, err = envfileloader.New[Conf](nil,
		// config
		envfileloader.WithDelimiter("."),
		envfileloader.WithFileType("json"),
		envfileloader.WithTag("env"),
		envfileloader.WithFiles(cfgPath),
		envfileloader.WithWatch(true),
	)
	if err != nil {
		panic(err)
	}

	cfg := loader.Get()

	fmt.Printf("[BOOT ] snapshot: name=%s env=%s db.user=%s\n", cfg.Name, cfg.Env, cfg.DB.Username)

}

func loadConfigWithCallbackWhenKeyIsTrue() {
	const cfgPath = "env.json"

	writeJSONAtomic(cfgPath, `{
  "name": "my-app",
  "env":  "dev",
  "db":   { "username": "devuser", "password": "devpass" }
}`)

	fnForOnChangeWatch := func(c *Conf, err error) {
		if err != nil {
			fmt.Println("[WATCH] error:", err)
			return
		}
		if c == nil {
			fmt.Println("[WATCH] nil config (unexpected)")
			return
		}
		fmt.Printf("[WATCH] reloaded: name=%s env=%s db.user=%s\n", c.Name, c.Env, c.DB.Username)
	}
	var err error
	loader, err = envfileloader.New(fnForOnChangeWatch,
		// config
		envfileloader.WithDelimiter("."),
		envfileloader.WithFileType("json"),
		envfileloader.WithTag("env"),
		envfileloader.WithFiles(cfgPath),
		envfileloader.WithWatch(true),
		envfileloader.WithCallbackOnChangeWhenOnKeyTrue("log_env"),
	)
	if err != nil {
		panic(err)
	}

	cfg := loader.Get()

	fmt.Printf("[BOOT ] snapshot: name=%s env=%s db.user=%s\n", cfg.Name, cfg.Env, cfg.DB.Username)

}
