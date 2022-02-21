package web

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/themisir/myfeed/pkg/adding"
	"github.com/themisir/myfeed/pkg/listing"
	"github.com/themisir/myfeed/pkg/updating"
	"net/http"
	"strconv"
)

// GET /
func (a *App) getIndexHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "index.html", nil)
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
		"Title": "Your feeds",
	})
}

// GET /feeds/create
func (a *App) getFeedsCreateHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "feeds/create.html", echo.Map{
		"Title": "Create a new feed",
	})
}

// POST /feeds/create
func (a *App) postFeedsCreateHandler(c echo.Context) error {
	userId, err := GetUserId(c)
	if err != nil {
		c.Logger().Errorf("Failed to get user id: %s", err)
		return echo.ErrInternalServerError
	}

	feed, err := a.feeds.AddFeed(adding.FeedData{
		Name:     c.FormValue("name"),
		IsPublic: true,
		UserId:   userId,
	})
	if err != nil {
		return echo.ErrBadRequest
	}

	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/feeds/%v/edit", feed.Id()))
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
		return echo.ErrForbidden
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
		"Title":   "Edit feed",
	})
}

// POST /feeds/delete
func (a *App) postFeedsDeleteHandler(c echo.Context) error {
	userId, err := GetUserId(c)
	if err != nil {
		c.Logger().Errorf("Failed to get user id: %s", err)
		return echo.ErrInternalServerError
	}

	// Parse feed id
	feedId, err := strconv.Atoi(c.FormValue("feedId"))
	if err != nil {
		return echo.ErrNotFound
	}

	// Find feed
	feed, err := a.feeds.GetFeed(feedId)
	if err != nil {
		return echo.ErrNotFound
	}
	if feed.UserId() != userId {
		return echo.ErrForbidden
	}

	// Remove feed
	if err := a.feeds.RemoveFeed(feedId); err != nil {
		a.logger.Errorf("failed to remove feed by id '%v': %s", feedId, err)
		return echo.ErrInternalServerError
	}

	return c.Redirect(http.StatusSeeOther, "/feeds")
}

type postFeedsEditDto struct {
	Name    string   `form:"name"`
	Sources []string `form:"sources"`
	Privacy string   `form:"privacy"`
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
		return echo.ErrForbidden
	}

	// Bind request body
	body := new(postFeedsEditDto)
	if err := c.Bind(body); err != nil {
		c.Logger().Errorf("Failed to bind body: %s", err)
		return echo.ErrBadRequest
	}

	// Update feed details
	if err := a.feeds.UpdateFeed(feedId, updating.Feed{
		Name:     body.Name,
		IsPublic: body.Privacy == "public",
	}); err != nil {
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

	// Check access
	if !feed.IsPublic() {
		userId, err := GetUserId(c)
		if err != nil || userId != feed.UserId() {
			return echo.ErrForbidden
		}
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
		"Title": feed.Name(),
	})
}

func (a *App) createFirstFeed(user listing.User) error {
	feedName := "My feed"
	_, err := a.feeds.AddFeed(adding.FeedData{
		Name:     feedName,
		UserId:   user.Id(),
		IsPublic: false,
	})
	return err
}
