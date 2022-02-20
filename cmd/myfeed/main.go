package main

import (
	"log"

	"github.com/themisir/myfeed/static"

	"github.com/themisir/myfeed/pkg/web"
)

func main() {
	config := &web.AppConfig{
		Address:      ":2342",
		AssetsRoot:   "assets",
		TemplateRoot: "views",
		StaticFS:     static.FS,
		DataSource:   "postgres://misir:@localhost:5432/myfeed_db?sslmode=disable",
	}

	app := web.NewApp(config)
	log.Fatal(app.Run())
}
