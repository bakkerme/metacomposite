package main

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
)

// API provides an implementation of the API API
type API struct {
	cfg *Config
}

// GetFeeds returns all available feeds
func (api *API) GetFeeds(ctx echo.Context) error {
	resp, err := json.Marshal(api.cfg.Feeds)
	if err != nil {
		return err
	}

	return ctx.String(http.StatusOK, string(resp))
}

// GetFeedsFeedID returns a feed by it's feedID
func (api *API) GetFeedsFeedID(ctx echo.Context, feedID string) error {
	return ctx.String(http.StatusNotImplemented, "Not Implemented")
}

// GetFeedsFeedIDPosts returns all posts associated with a given feed ID
func (api *API) GetFeedsFeedIDPosts(ctx echo.Context, feedID string) error {
	return ctx.String(http.StatusNotImplemented, "Not Implemented")
}

// GetGroupGroupID returns a group by it's groupID
func (api *API) GetGroupGroupID(ctx echo.Context, groupID string) error {
	return ctx.String(http.StatusNotImplemented, "Not Implemented")
}

// GetGroups returns all groups
func (api *API) GetGroups(ctx echo.Context) error {
	resp, err := json.Marshal(api.cfg.Groups)
	if err != nil {
		return err
	}

	return ctx.String(http.StatusOK, string(resp))
}

// GetGroupsGroupIDFeeds returns a list of feeds associated with a given group ID
func (api *API) GetGroupsGroupIDFeeds(ctx echo.Context, groupID string) error {
	return ctx.String(http.StatusNotImplemented, "Not Implemented")
}

// GetGroupGroupIDPosts returns a list of posts associated with a given group
func (api *API) GetGroupGroupIDPosts(ctx echo.Context, groupID string) error {
	return ctx.String(http.StatusNotImplemented, "Not Implemented")
}