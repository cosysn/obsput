package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type OBS struct {
	Name     string `yaml:"name"`
	Endpoint string `yaml:"endpoint"`
	Bucket   string `yaml:"bucket"`
	AK       string `yaml:"ak"`
	SK       string `yaml:"sk"`
}

type Config struct {
	Configs map[string]*OBS `yaml:"configs"`
}

func NewConfig() *Config {
	return &Config{
		Configs: make(map[string]*OBS),
	}
}

func (c *Config) AddOBS(name, endpoint, bucket, ak, sk string) {
	c.Configs[name] = &OBS{
		Name:     name,
		Endpoint: endpoint,
		Bucket:   bucket,
		AK:       ak,
		SK:       sk,
	}
}

func (c *Config) GetOBS(name string) *OBS {
	return c.Configs[name]
}

func (c *Config) ListOBS() []*OBS {
	obsList := make([]*OBS, 0, len(c.Configs))
	for _, obs := range c.Configs {
		obsList = append(obsList, obs)
	}
	return obsList
}

func (c *Config) RemoveOBS(name string) {
	delete(c.Configs, name)
}

func (c *Config) OBSExists(name string) bool {
	_, ok := c.Configs[name]
	return ok
}

func (c *Config) Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Configs == nil {
		cfg.Configs = make(map[string]*OBS)
	}
	return &cfg, nil
}
