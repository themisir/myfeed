package main

import (
	"github.com/themisir/myfeed/static"
	"log"

	"github.com/themisir/myfeed/pkg/web"
)

func main() {
	config := &web.AppConfig{
		Address:      ":2342",
		AssetsRoot:   "assets",
		TemplateRoot: "views",
		StaticFS:     static.FS,
	}

	app := web.NewApp(config)
	log.Fatal(app.Run())
}
