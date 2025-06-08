package server

import (
	"io"
	"io/fs"

	"gopkg.in/yaml.v3"
)

type HTTPPatterns struct {
	API struct {
		Permissions map[string]struct {
			Values map[string]string `yaml:"values"`
		} `yaml:"permissions"`
	} `yaml:"api"`
}

func loadPatterns(f fs.File) error {
	httpPatterns := HTTPPatterns{}
	out, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(out, &httpPatterns); err != nil {
		return err
	}
}
