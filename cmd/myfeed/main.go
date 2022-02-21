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

	if dataSource, ok := os.LookupEnv("DATABASE_URL"); ok {
		config := &web.AppConfig{
			Address:      ":2342",
			AssetsRoot:   "assets",
			TemplateRoot: "views",
			StaticFS:     static.FS,
			DataSource:   dataSource,
		}

		app := web.NewApp(config)
		app.Run()
	}
}
