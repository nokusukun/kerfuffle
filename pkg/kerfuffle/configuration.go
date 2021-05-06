/*
 * Copyright (c) 2021 @nokusukun.
 * This file is part of Kerfuffle which is released under Apache.
 * See file LICENSE or go to https://github.com/nokusukun/kerfuffle/blob/master/LICENSE for full license details.
 */

package kerfuffle

type Meta struct {
	Name string `toml:"name" json:"name"`
}

type Provision struct {
	Id                   string     `toml:"-" json:"id,omitempty"`
	HealthEndpoint       string     `toml:"health_endpoint" json:"health_endpoint,omitempty"`
	EventUrl             string     `toml:"event_url" json:"event_url,omitempty"`
	Run                  [][]string `toml:"run" json:"run,omitempty"`
	EnvironmentVariables []string   `toml:"envs" json:"environment_variables,omitempty"`
	BaseDirectory        string     `toml:"base_dir" json:"base_directory,omitempty"`
}

type Proxy struct {
	Host      []string `toml:"host" json:"host"`
	BindPort  string   `toml:"bind_port" json:"bind_port"`
	StaticDir string   `toml:"static_dir" json:"static_dir"`
	Hold      bool     `json:"hold"`
}

type Cloudflare struct {
	Host    []string `toml:"host" json:"host,omitempty"`
	Zone    string   `toml:"zone" json:"zone,omitempty"`
	Proxied bool     `toml:"proxied" json:"proxied,omitempty"`
}
