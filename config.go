package main

import (
	"encoding/json"
	"io/ioutil"
)

// Option is a single configuration option.
type Option struct {
	Name  string
	Value interface{}
}

// Config is a configuration file.
type Config struct {
	Options []Option
}

// LoadConfig loads a configuration file.
func LoadConfig(filename string) (Config, error) {
	config := Config{}

	j, err := ioutil.ReadFile(filename)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(j, &config)
	return config, err
}

// Save saves the configuration to a file.
func (c Config) Save(filename string) error {
	j, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, j, 0600)
}

// Value returns the value of a configuration option.
func (c Config) Value(name string) interface{} {
	for _, v := range c.Options {
		if v.Name == name {
			return v.Value
		}
	}

	return nil
}

// Set sets the value of a configuration option.
func (c *Config) Set(name, value string) {
	found := false
	var opts []Option
	for _, v := range c.Options {
		if v.Name == name {
			v.Value = value
			found = true
		}

		opts = append(opts, v)
	}

	if !found {
		opts = append(opts, Option{name, value})
	}

	c.Options = opts
}
