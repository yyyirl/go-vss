package rule

import (
	"skeyevss/core/app/sev/vss/internal/config"
	"skeyevss/core/repositories/models/settings"
)

type Config struct {
	conf    config.Config
	setting *settings.Item
}

func NewConfig(conf config.Config, setting *settings.Item) *Config {
	if setting == nil {
		setting = &settings.Item{
			Content: new(settings.Content),
		}
	}

	return &Config{
		conf:    conf,
		setting: setting,
	}
}

func (c *Config) Conv() *Config {
	c.setting.ItemCorrection(&settings.ItemCorrectionParams{
		BaseConf:   c.conf.SevBase,
		SipConf:    c.conf.Sip,
		InternalIp: c.conf.InternalIp,
		ExternalIp: c.conf.ExternalIp,
	})

	return c
}

func (c *Config) SipIP() string {
	if c.conf.Sip.UseExternalWan {
		return c.conf.ExternalIp
	}

	return c.conf.InternalIp
}

func (c *Config) Content() *settings.Content {
	return c.setting.Content
}

func (c *Config) Setting() *settings.Item {
	return c.setting
}
