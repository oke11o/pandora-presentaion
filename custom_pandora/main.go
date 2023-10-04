package main

import (
	"github.com/spf13/afero"
	"github.com/yandex/pandora/cli"
	grpc "github.com/yandex/pandora/components/grpc/import"
	phttp "github.com/yandex/pandora/components/phttp/import"
	"github.com/yandex/pandora/core"
	coreimport "github.com/yandex/pandora/core/import"
)

type Ammo struct {
	UserID   int64  `json:"user_id"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

func main() {
	fs := afero.NewOsFs()
	coreimport.Import(fs)
	phttp.Import(fs)
	grpc.Import(fs)
	Import(fs)
	// Custom imports. Integrate your custom types into configuration system.
	coreimport.RegisterCustomJSONProvider("users/json", func() core.Ammo { return &Ammo{} })

	cli.Run()
}
