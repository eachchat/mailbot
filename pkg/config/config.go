package config

import (
	"io/ioutil"

	"github.com/eachchat/mailbot/internal/db"
	"github.com/eachchat/mailbot/pkg/logger"
	"gopkg.in/yaml.v3"
)

// Configuration  the global configurations for app.
type Configuration struct {
	Log        logger.LogConf `yaml:"log"`
	MatrixConf `yaml:",inline"`
	DB         db.DBCONF `yaml:"db"`
}

// MatrixConf the configurations for matrix
type MatrixConf struct {
	MatrixServer             string   `yaml:"matrixserver"`
	MatrixUserPassword       string   `yaml:"matrixuserpassword"`
	Matrixaccesstoken        string   `yaml:"matrixaccesstoken"`
	MatrixUserID             string   `yaml:"matrixuserid"`
	AllowedServers           []string `yaml:"allowed_servers"`
	DefaultMailCheckInterval int      `yaml:"defaultmailCheckInterval"`
	MarkdownEnabledByDefault bool     `yaml:"markdownEnabledByDefault"`
	HtmlDefault              bool     `yaml:"htmlDefault"`
}

func (c *Configuration) InitConf(confPath string) error {
	b, err := ioutil.ReadFile(confPath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(b, c)
	if err != nil {
		return err
	}
	return nil
}
