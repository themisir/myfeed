package main

import (
	"os"

	"github.com/joho/godotenv"

	"github.com/themisir/myfeed/pkg/web"
	"github.com/themisir/myfeed/static"
)

//goland:noinspection GoUnhandledErrorResult
func main() {
	godotenv.Load()

	// TODO: Improve config loading using https://github.com/spf13/viper

	if dataSource, ok := os.LookupEnv("DATABASE_URL"); ok {
		address, ok := os.LookupEnv("ADDRESS")
		if !ok {
			address = ":2342"
		}

		config := &web.AppConfig{
			Address:      address,
			AssetsRoot:   "assets",
			TemplateRoot: "views",
			StaticFS:     static.FS,
			DataSource:   dataSource,
		}

		app := web.NewApp(config)
		app.Run()
	} else {
		panic("DATABASE_URL environment variable is missing")
	}
}
