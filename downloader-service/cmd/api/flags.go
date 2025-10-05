package main

import (
	"flag"
)

func (c *Config) flags() {
	flag.IntVar(&c.Port, "p", 80, "server port")
	flag.BoolVar(&c.Fs, "file-storage", false, "Start NATS as an embedded server")
	flag.BoolVar(&c.Nats.Embedded, "embed-nats", false, "Start NATS as an embedded server")
	flag.StringVar(&c.Nats.Url, "nats", "", "Nats server URL")
	flag.StringVar(&c.Redis.Addr, "redis", "", "Redis server address")

	flag.Parse()
}
