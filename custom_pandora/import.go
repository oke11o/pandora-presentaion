package main

import (
	"github.com/spf13/afero"
	"github.com/yandex/pandora/core"
	"github.com/yandex/pandora/core/register"
)

func Import(fs afero.Fs) {
	register.Gun("custom_http_gun", func(conf GunConfig) core.Gun {
		return &Gun{
			conf: conf,
		}
	}, defaultGunConfig)
}
