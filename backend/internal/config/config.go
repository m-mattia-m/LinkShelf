package config

import (
	"flag"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App struct {
		Name        string `yaml:"name" json:"name" mapstructure:"name"`
		Description string `yaml:"description" json:"description" mapstructure:"description"`
		Environment string `yaml:"environment" json:"environment" mapstructure:"environment"`
		Logo        string `yaml:"logo" json:"logo" mapstructure:"logo"`
	} `yaml:"app" json:"app" mapstructure:"app"`

	Server struct {
		Scheme string `yaml:"scheme" json:"scheme" mapstructure:"scheme"`
		Host   string `yaml:"host" json:"host" mapstructure:"host"`
		Port   string `yaml:"port" json:"port" mapstructure:"port"`
	}

	Database struct {
		Engine   string `yaml:"engine" json:"engine" mapstructure:"engine"`
		Host     string `yaml:"host" json:"host" mapstructure:"host"`
		Port     string `yaml:"port" json:"port" mapstructure:"port"`
		Username string `yaml:"username" json:"username" mapstructure:"username"`
		Password string `yaml:"password" json:"password" mapstructure:"password"`
		Name     string `yaml:"name" json:"name" mapstructure:"name"`
		Params   string `yaml:"params" json:"params" mapstructure:"params"`
	} `yaml:"database" json:"database" mapstructure:"database"`

	Logging struct {
		Level string `yaml:"level" json:"level" mapstructure:"level"`
	} `yaml:"logging" json:"logging" mapstructure:"logging"`

	Domain struct {
		OpenAPI struct {
			UserPort string `yaml:"userPort" json:"userPort" mapstructure:"userPort"`
		}
		Authentication struct {
			SkipAuthentication bool `yaml:"skipAuthentication" json:"skipAuthentication" mapstructure:"skipAuthentication"`
		} `yaml:"authentication" json:"authentication" mapstructure:"authentication"`
	}
}

func LoadConfig() error {
	viper.SetEnvPrefix("APP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetConfigType("yaml")

	if isRunningTests() {
		viper.SetConfigName("config.test")
	} else {
		viper.SetConfigName("config.default")
	}

	viper.AddConfigPath(".")
	viper.AddConfigPath("..")
	viper.AddConfigPath("../..")
	viper.AddConfigPath("../../..")
	viper.AddConfigPath("../../../..")
	viper.AddConfigPath("backend")
	viper.AddConfigPath("./backend")

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	viper.SetConfigName("config")
	_ = viper.MergeInConfig()

	return nil
}
func isRunningTests() bool {
	return flag.Lookup("test.v") != nil
}
