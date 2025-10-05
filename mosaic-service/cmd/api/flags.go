package main

import "flag"

func (c *Config) flags() {
	flag.IntVar(&c.Port, "p", 80, "port to listen on")
	flag.StringVar(&c.Redis.Addr, "redis", "", "redis server URL")

	flag.Parse()
}
