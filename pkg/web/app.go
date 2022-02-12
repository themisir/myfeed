package web

import (
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/themisir/myfeed/pkg/adding"
	"github.com/themisir/myfeed/pkg/auth"
	"github.com/themisir/myfeed/pkg/models"
	"github.com/themisir/myfeed/pkg/sources"
	"github.com/themisir/myfeed/pkg/storage/memory"
	"github.com/themisir/myfeed/pkg/web/renderer"
	"io/fs"
	"net/http"
	"net/mail"
	"os"
	"time"
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
	users   models.UserRepository

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
	a.initAuth(e)
	a.initManager()
	a.initRoutes(e)

	return e.Start(a.config.Address)
}

func (a *App) initAuth(e *echo.Echo) {
	handler := auth.New(auth.CookieSchema([]byte("hello"), 30*24*time.Hour))

	e.Use(handler.Init)

	e.GET("/login", func(c echo.Context) error {
		return c.Render(http.StatusOK, "login.html", echo.Map{"Title": "Login"})
	})

	e.POST("/login", func(c echo.Context) error {
		email := c.FormValue("email")
		password := c.FormValue("password")

		user, _ := a.users.FindUserByEmail(email)
		if user != nil {
			err := handler.SignIn(c, user, password)
			if err == nil {
				return c.Redirect(http.StatusSeeOther, "/feeds")
			} else if !errors.Is(err, auth.ErrInvalidPassword) {
				c.Logger().Errorf("Failed to sign in with user '%s': %s", user.Id(), err)

				return c.Render(http.StatusOK, "login.html", echo.Map{
					"Error": "Email and password is correct but failed to sign in to the account, please try again",
				})
			}
		}

		return c.Render(http.StatusOK, "login.html", echo.Map{
			"Error": "Invalid email or password",
			"Title": "Login",
		})
	})

	e.GET("/register", func(c echo.Context) error {
		return c.Render(http.StatusOK, "register.html", echo.Map{"Title": "Register"})
	})

	e.POST("/register", func(c echo.Context) error {
		email := c.FormValue("email")
		password := c.FormValue("password")

		if _, err := mail.ParseAddress(email); err != nil {
			return c.Render(http.StatusOK, "register.html", echo.Map{
				"Error": "Invalid email address",
				"Title": "Register",
			})
		}

		if len(password) < 6 {
			return c.Render(http.StatusOK, "register.html", echo.Map{
				"Error": "Password must be at least 6 characters long",
				"Title": "Register",
			})
		}

		user, err := a.users.AddUser(adding.UserData{
			Email:        email,
			PasswordHash: handler.HashPassword(password),
		})
		if err != nil {
			c.Logger().Errorf("Failed to create user with email '%s': %s", email, err)

			return c.Render(http.StatusOK, "register.html", echo.Map{
				"Error": "Failed to create user account, please try again",
				"Title": "Register",
			})
		}

		if err := a.createFirstFeed(user); err != nil {
			c.Logger().Errorf("Failed to create first feed for user '%s': %s", user.Id(), err)
		}

		if err := handler.SignInWithoutPassword(c, user); err != nil {
			c.Logger().Errorf("Failed to sign in with user '%s': %s", user.Id(), err)
			return c.Redirect(http.StatusSeeOther, "/login")
		}

		return c.Redirect(http.StatusSeeOther, "/feeds")
	})

	e.Use(Authorize(false))
}

func (a *App) initStorage() {
	feedRepository := memory.NewFeedRepository()
	feedRepository.Persist(memory.JSON("data/feeds.json"))

	sourceRepository := memory.NewSourceRepository(feedRepository)
	sourceRepository.Persist(memory.JSON("data/sources.json"))

	postRepository := memory.NewPostRepository(feedRepository, sourceRepository)
	postRepository.Persist(memory.JSON("data/posts.json"))

	userRepository := memory.NewUserRepository()
	userRepository.Persist(memory.JSON("data/users.json"))

	a.feeds = feedRepository
	a.sources = sourceRepository
	a.posts = postRepository
	a.users = userRepository
}

func (a *App) initManager() {
	a.sourceManager = sources.NewManager(a.sources, a.posts, a.feeds)
	if err := a.sourceManager.Start(); err != nil {
		panic(err)
	}
}

func (a *App) initRoutes(e *echo.Echo) {
	e.GET("/", a.getIndexHandler)

	e.GET("/feeds", a.getFeedsHandler, Authorize(true))

	e.GET("/feeds/:feedId", a.getFeedHandler)

	e.GET("/feeds/create", a.getFeedsCreateHandler, Authorize(true))
	e.POST("/feeds/create", a.postFeedsCreateHandler, Authorize(true))

	e.GET("/feeds/:feedId/edit", a.getFeedsEditHandler, Authorize(true))
	e.POST("/feeds/:feedId/edit", a.postFeedsEditHandler, Authorize(true))
}
