package web

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/themisir/myfeed/pkg/models"
	"github.com/themisir/myfeed/pkg/sources"
	"github.com/themisir/myfeed/pkg/storage/memory"
	"github.com/themisir/myfeed/pkg/web/renderer"
	"io/fs"
	"net/http"
	"os"
)

type AppConfig struct {
	Address      string
	TemplateRoot string
	AssetsRoot   string
	StaticFS     fs.FS
}

type App struct {
	fs     fs.FS
	config *AppConfig

	sources models.SourceRepository
	feeds   models.FeedRepository
	posts   models.PostRepository

	sourceManager *sources.Manager
}

func NewApp(config *AppConfig) *App {
	app := &App{
		config: config,
	}

	if config.StaticFS == nil {
		app.fs = os.DirFS(".")
	} else {
		app.fs = config.StaticFS
	}

	return app
}

func (a *App) Run() error {
	e := echo.New()
	e.Renderer = renderer.Layout("layout.html", renderer.Template(a.fs, a.config.TemplateRoot))

	e.Pre(middleware.RemoveTrailingSlash())

	e.Use(middleware.Logger())
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Filesystem: http.FS(a.fs),
		Root:       a.config.AssetsRoot,
	}))

	a.initStorage()
	a.initManager()
	a.initRoutes(e)

	return e.Start(a.config.Address)
}

func (a *App) initStorage() {
	feedRepository := memory.NewFeedRepository()
	feedRepository.Persist(memory.JSON("data/feeds.json"))

	sourceRepository := memory.NewSourceRepository(feedRepository)
	sourceRepository.Persist(memory.JSON("data/sources.json"))

	postRepository := memory.NewPostRepository(feedRepository, sourceRepository)
	postRepository.Persist(memory.JSON("data/posts.json"))

	a.feeds = feedRepository
	a.sources = sourceRepository
	a.posts = postRepository
}

func (a *App) initManager() {
	a.sourceManager = sources.NewManager(a.sources, a.posts, a.feeds)
	if err := a.sourceManager.Start(); err != nil {
		panic(err)
	}
}

func (a *App) initRoutes(e *echo.Echo) {
	e.GET("/", a.getIndexHandler)

	e.GET("/feeds", a.getFeedsHandler)

	e.GET("/feeds/:feedId", a.getFeedHandler)

	e.GET("/feeds/create", a.getFeedsCreateHandler)
	e.POST("/feeds/create", a.postFeedsCreateHandler)

	e.GET("/feeds/:feedId/edit", a.getFeedsEditHandler)
	e.POST("/feeds/:feedId/edit", a.postFeedsEditHandler)
}
