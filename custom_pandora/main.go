package main

import (
	"github.com/spf13/afero"
	"github.com/yandex/pandora/cli"
	grpc "github.com/yandex/pandora/components/grpc/import"
	http "github.com/yandex/pandora/components/phttp/import"
	coreimport "github.com/yandex/pandora/core/import"
)

func main() {
	fs := afero.NewOsFs()
	coreimport.Import(fs)
	http.Import(fs)
	grpc.Import(fs)

	Import(fs)

	cli.Run()
}
