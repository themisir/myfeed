package web

import (
	"errors"
	"fmt"
	"github.com/themisir/myfeed/pkg/log"
	"io/fs"
	"net/http"
	"net/mail"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/themisir/myfeed/pkg/adding"
	"github.com/themisir/myfeed/pkg/auth"
	"github.com/themisir/myfeed/pkg/models"
	"github.com/themisir/myfeed/pkg/sources"
	"github.com/themisir/myfeed/pkg/storage/memory"
	"github.com/themisir/myfeed/pkg/storage/postgres"
	"github.com/themisir/myfeed/pkg/web/renderer"
)

type AppConfig struct {
	Address      string
	TemplateRoot string
	AssetsRoot   string
	StaticFS     fs.FS
	DataSource   string
}

type App struct {
	fs     fs.FS
	config *AppConfig

	sources models.SourceRepository
	feeds   models.FeedRepository
	posts   models.PostRepository
	users   models.UserRepository

	logger   log.Logger
	renderer *renderer.MetadataRenderer

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

func (a *App) Run() {
	e := echo.New()

	// Configure renderer
	a.renderer = renderer.Metadata(renderer.Layout("layout.html", renderer.Template(a.fs, a.config.TemplateRoot)))
	e.Renderer = a.renderer

	e.Pre(middleware.RemoveTrailingSlash())

	e.Use(middleware.Logger())
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Filesystem: http.FS(a.fs),
		Root:       a.config.AssetsRoot,
	}))

	a.logger = e.Logger

	a.initStorage()
	a.initAuth(e)
	a.initManager()
	a.initRoutes(e)

	if err := e.Start(a.config.Address); err != nil {
		a.logger.Errorf("failed to start server: %s", err)
	}
}

func (a *App) initAuth(e *echo.Echo) {
	handler := auth.New(auth.CookieSchema([]byte("hello"), 30*24*time.Hour))

	e.Use(handler.Init)

	e.GET("/login", func(c echo.Context) error {
		return c.Render(http.StatusOK, "login.html", echo.Map{"Title": "Login"})
	})

	e.POST("/login", func(c echo.Context) error {
		username := c.FormValue("username")
		password := c.FormValue("password")

		user, _ := a.users.FindUserByUsername(username)
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
		username := c.FormValue("username")
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

		if _, err := a.users.FindUserByUsername(username); err == nil {
			return c.Render(http.StatusOK, "register.html", echo.Map{
				"Error": "Username is already in use",
				"Title": "Register",
			})
		}

		user, err := a.users.AddUser(adding.UserData{
			Email:        email,
			Username:     username,
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

	e.GET("/logout", func(c echo.Context) error {
		handler.SignOut(c)
		return c.Redirect(http.StatusSeeOther, "/")
	})

	e.Use(Authorize(false))

	a.renderer.SetDyn("IsLoggedIn", func(c echo.Context) interface{} {
		return handler.GetUserId(c) != ""
	})

	a.renderer.SetDyn("User", func(c echo.Context) interface{} {
		userId := handler.GetUserId(c)
		if userId != "" {
			user, err := a.users.GetUserById(userId)
			if err != nil {
				a.logger.Errorf("failed to get user by id '%s': %s", userId, err)
				return nil
			}
			return user
		}
		return nil
	})
}

func (a *App) initStorage() {
	if a.config.DataSource == "memory://" {
		a.initMemoryStorage()
	} else {
		a.initDbStorage()
	}
}

func (a *App) initDbStorage() {
	db, err := postgres.Connect(a.config.DataSource)
	initerr(err, "failed to connect to the database: %s")
	a.feeds, err = db.Feeds()
	initerr(err, "failed to create feed repository: %s")
	a.posts, err = db.Posts()
	initerr(err, "failed to create post repository: %s")
	a.sources, err = db.Sources()
	initerr(err, "failed to create source repository: %s")
	a.users, err = db.Users()
	initerr(err, "failed to create user repository: %s")
}

func (a *App) initMemoryStorage() {
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
	a.sourceManager = sources.NewManager(a.sources, a.posts, a.feeds, a.logger)
	if err := a.sourceManager.Start(); err != nil {
		initerr(err, "failed to start source manager: %s")
	}
}

func (a *App) initRoutes(e *echo.Echo) {
	e.GET("/", a.getIndexHandler)

	e.GET("/feeds", a.getFeedsHandler, Authorize(true))

	e.GET("/feeds/:feedId", a.getFeedHandler)

	e.GET("/feeds/create", a.getFeedsCreateHandler, Authorize(true))
	e.POST("/feeds/create", a.postFeedsCreateHandler, Authorize(true))

	e.POST("/feeds/delete", a.postFeedsDeleteHandler, Authorize(true))

	e.GET("/feeds/:feedId/edit", a.getFeedsEditHandler, Authorize(true))
	e.POST("/feeds/:feedId/edit", a.postFeedsEditHandler, Authorize(true))
}

func initerr(err error, format string) {
	if err != nil {
		panic(fmt.Sprintf(format, err))
	}
}
