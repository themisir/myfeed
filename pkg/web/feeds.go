package web

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/themisir/myfeed/pkg/adding"
	"github.com/themisir/myfeed/pkg/updating"
)

// GET /
func (a *App) getIndexHandler(c echo.Context) error {
	return c.Redirect(http.StatusTemporaryRedirect, "/feeds")
}

// GET /feeds
func (a *App) getFeedsHandler(c echo.Context) error {
	userId, err := GetUserId(c)
	if err != nil {
		c.Logger().Errorf("Failed to get user id: %s", err)
		return echo.ErrInternalServerError
	}

	feeds, err := a.feeds.GetUserFeeds(userId)
	if err != nil {
		c.Logger().Errorf("Failed to fetch feeds: %s", err)
		return echo.ErrInternalServerError
	}

	return c.Render(http.StatusOK, "feeds/list.html", echo.Map{
		"Feeds": feeds,
	})
}

// GET /feeds/create
func (a *App) getFeedsCreateHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "feeds/create.html", nil)
}

// POST /feeds/create
func (a *App) postFeedsCreateHandler(c echo.Context) error {
	userId, err := GetUserId(c)
	if err != nil {
		c.Logger().Errorf("Failed to get user id: %s", err)
		return echo.ErrInternalServerError
	}

	_, err = a.feeds.AddFeed(adding.FeedData{
		Name:   c.FormValue("name"),
		UserId: userId,
	})
	if err != nil {
		return echo.ErrBadRequest
	}

	return c.Redirect(http.StatusSeeOther, "/feeds")
}

// GET /feeds/:feedId/edit
func (a *App) getFeedsEditHandler(c echo.Context) error {
	userId, err := GetUserId(c)
	if err != nil {
		c.Logger().Errorf("Failed to get user id: %s", err)
		return echo.ErrInternalServerError
	}

	// Parse feed id
	feedId, err := strconv.Atoi(c.Param("feedId"))
	if err != nil {
		return echo.ErrNotFound
	}

	// Find feed
	feed, err := a.feeds.GetFeed(feedId)
	if err != nil {
		return echo.ErrNotFound
	}
	if feed.UserId() != userId {
		// TODO: access check
	}

	// Get feed sources
	sources, err := a.sources.GetFeedSources(feedId)
	if err != nil {
		c.Logger().Errorf("Failed to list feed '%v' sources: %s", feedId, err)
		return echo.ErrInternalServerError
	}

	return c.Render(http.StatusOK, "feeds/edit.html", echo.Map{
		"Feed":    feed,
		"Sources": sources,
	})
}

type postFeedsEditDto struct {
	Name    string   `form:"name"`
	Sources []string `form:"sources"`
}

// POST /feeds/:feedId/edit
func (a *App) postFeedsEditHandler(c echo.Context) error {
	userId, err := GetUserId(c)
	if err != nil {
		c.Logger().Errorf("Failed to get user id: %s", err)
		return echo.ErrInternalServerError
	}

	// Parse feed id
	feedId, err := strconv.Atoi(c.Param("feedId"))
	if err != nil {
		return echo.ErrNotFound
	}

	// Find feed
	feed, err := a.feeds.GetFeed(feedId)
	if err != nil {
		return echo.ErrNotFound
	}
	if feed.UserId() != userId {
		// TODO: access check
	}

	// Bind request body
	body := new(postFeedsEditDto)
	if err := c.Bind(body); err != nil {
		c.Logger().Errorf("Failed to bind body: %s", err)
		return echo.ErrBadRequest
	}

	// Update feed details
	if err := a.feeds.UpdateFeed(feedId, updating.Feed{Name: body.Name}); err != nil {
		c.Logger().Errorf("Failed to update feed '%v': %s", feedId, err)
		return echo.ErrInternalServerError
	}

	// Update feed sources
	if err := a.sourceManager.UpdateFeedSources(feedId, body.Sources...); err != nil {
		c.Logger().Errorf("Failed to update feed sources '%v': %s", feedId, err)
		return echo.ErrInternalServerError
	}

	return c.Redirect(http.StatusSeeOther, "/feeds")
}

// GET /feeds/:feedId
func (a *App) getFeedHandler(c echo.Context) error {
	// Parse feed id
	feedId, err := strconv.Atoi(c.Param("feedId"))
	if err != nil {
		return echo.ErrNotFound
	}

	// Find feed
	feed, err := a.feeds.GetFeed(feedId)
	if err != nil {
		return echo.ErrNotFound
	}

	// Get posts
	posts, err := a.posts.GetFeedPosts(feedId)
	if err != nil {
		c.Logger().Errorf("Failed to get feed '%v' posts: %s", feedId, err)
		return echo.ErrInternalServerError
	}

	return c.Render(http.StatusOK, "feeds/single.html", echo.Map{
		"Feed":  feed,
		"Posts": posts,
	})
}
