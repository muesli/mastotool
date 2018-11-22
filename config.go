package main

import (
	"encoding/json"
	"io/ioutil"
)

type Option struct {
	Name  string
	Value interface{}
}

type Config struct {
	Options []Option
}

func LoadConfig(filename string) (Config, error) {
	config := Config{}

	j, err := ioutil.ReadFile(filename)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(j, &config)
	return config, err
}

func (c Config) Save(filename string) error {
	j, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, j, 0644)
}

func (c Config) Value(name string) interface{} {
	for _, v := range c.Options {
		if v.Name == name {
			return v.Value
		}
	}

	return nil
}

func (c *Config) Set(name, value string) interface{} {
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
	return nil
}
