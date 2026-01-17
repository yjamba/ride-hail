package handlers

import "fmt"

type ServerConfig struct {
	Addr string
	Port int
}

func NewServerConfig(addr string, port int) *ServerConfig {
	return &ServerConfig{
		Addr: addr,
		Port: port,
	}
}

func (c *ServerConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Addr, c.Port)
}
