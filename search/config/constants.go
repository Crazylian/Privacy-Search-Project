package config

import "math"

type Config struct {
	preamble string
}

func MakeConfig(preambleStr string) *Config {
	c := Config{
		preamble: preambleStr,
	}
	return &c
}

func (c *Config) PREAMBLE() string {
	return c.preamble
}

func (c *Config) DEFAULT_EMBEDDINGS_HINT_SZ() uint64 {
	return 500
}

func (c *Config) DEFAULT_URL_HINT_SZ() uint64 {
	return 100
}

func (c *Config) EMBEDDINGS_DIM() uint64 {
	return 192
}

func SLOT_BITS() uint64 {
	return 5
}

func (c *Config) TOTAL_NUM_CLUSTERS() int {
	return 14000
}

func (c *Config) MAX_EMBEDDINGS_SERVERS() int {
	return 1
}

func (c *Config) EMBEDDINGS_CLUSTERS_PER_SERVER() int {
	clustersPerServer := float64(c.TOTAL_NUM_CLUSTERS()) / float64(c.MAX_EMBEDDINGS_SERVERS())
	return int(math.Ceil(clustersPerServer))
}

func (c *Config) URL_CLUSTERS_PER_SERVER() int {
	clustersPerServer := float64(c.TOTAL_NUM_CLUSTERS()) / float64(c.MAX_URL_SERVERS())
	return int(math.Ceil(clustersPerServer))
}

func (c *Config) MAX_URL_SERVERS() int {
	return 1
}

func (c *Config) SIMPLEPIR_EMBEDDINGS_RECORD_LENGTH() int {
	return 17
}
