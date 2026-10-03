package config

// Rule 定義單個規則集的映射關係
type Rule struct {
	Name      string `yaml:"name"`
	URL       string `yaml:"url"`
	LocalFile string `yaml:"local_file"`
	Enabled   bool   `yaml:"enabled"`
}
