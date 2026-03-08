package config

import "time"

type Config struct {
	GrpcServer GrpcServer        `yaml:"grpc_server"`
	HttpServer HttpServer        `yaml:"http_server"`
	Graceful   Graceful          `yaml:"graceful"`
	Targets    map[string]string `yaml:"service"`
}

type GrpcServer struct {
	Port              int           `yaml:"port"               env:"GRPC_PORT"`
	MaxConnectionIdle time.Duration `yaml:"max_connection_idle"`
	MaxConnectionAge  time.Duration `yaml:"max_connection_age"`
	Time              time.Duration `yaml:"time"`
	Timeout           time.Duration `yaml:"timeout"`
}

type HttpServer struct {
	Port      int `yaml:"port"       env:"HTTP_PORT"`
	AdminPort int `yaml:"admin_port" env:"ADMIN_HTTP_PORT"`
}

type Graceful struct {
	Timeout time.Duration `yaml:"timeout"`
}
