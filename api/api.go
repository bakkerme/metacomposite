package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/bakkerme/metacomposite/v2/types"
	"github.com/labstack/echo/v4"
)

// feedErrors is a struct used to hold Posts and any errors found
// during the fetch process
type feedErrors struct {
	Posts *[]types.Post
	Err   types.Error
}

// Loaders is a type that contains a Loader for each type of content that can output a Feed
type Loaders struct {
	Reddit types.Loader
	RSS    types.Loader
}

func setGlobalHeaders(ctx echo.Context) {
	ctx.Response().Header().Set("Access-Control-Allow-Origin", "*")
}

// API provides an implementation of the API API
type API struct {
	CFG     *Config
	Loaders Loaders
}

// GetFeeds returns all available feeds
func (api *API) GetFeeds(ctx echo.Context) error {
	resp, err := json.Marshal(api.CFG.Feeds)
	if err != nil {
		return err
	}

	setGlobalHeaders(ctx)
	return ctx.String(http.StatusOK, string(resp))
}

type getFeedsPostsResponse struct {
	Posts  []types.Post  `json:"posts"`
	Errors []types.Error `json:"errors"`
}

// GetFeedsPosts gets all posts from all feeds
func (api *API) GetFeedsPosts(ctx echo.Context) error {
	feeds := api.CFG.Feeds

	ch := make(chan feedErrors)
	for _, feed := range feeds {
		go func(feedToLoad types.Feed) {
			posts, err := getPostsForFeed(api.Loaders, &feedToLoad)

			var errorReturn types.Error
			if err != nil {
				ids := []string{feedToLoad.ID}
				errMessage := err.Error()
				errorReturn = types.Error{
					Code:       errorFeedLoadFail,
					Message:    errMessage,
					RelatedIDs: ids,
				}
			}

			ch <- feedErrors{
				Posts: posts,
				Err:   errorReturn,
			}
		}(feed)
	}

	posts := []types.Post{}
	errors := []types.Error{}
	for range feeds {
		out := <-ch
		if out.Posts == nil {
			errors = append(errors, out.Err)
		} else {
			posts = append(posts, *out.Posts...)
		}
	}

	// sort posts by timestamp
	sort.SliceStable(posts, func(i, j int) bool {
		return posts[i].Timestamp > posts[j].Timestamp
	})

	resp, err := json.Marshal(getFeedsPostsResponse{
		Posts:  posts,
		Errors: errors,
	})
	if err != nil {
		return err
	}

	setGlobalHeaders(ctx)
	return ctx.String(http.StatusOK, string(resp))
}

// GetFeedsFeedID returns a feed by it's feedID
func (api *API) GetFeedsFeedID(ctx echo.Context, feedID string) error {
	feed := getFeedByID(feedID, &api.CFG.Feeds)
	if feed == nil {
		return ctx.String(http.StatusNotFound, "Could not find "+feedID)
	}

	resp, err := json.Marshal(feed)
	if err != nil {
		return err
	}

	setGlobalHeaders(ctx)
	return ctx.String(http.StatusOK, string(resp))
}

// GetFeedsFeedIDPosts returns all posts associated with a given feed ID
func (api *API) GetFeedsFeedIDPosts(ctx echo.Context, feedID string) error {
	feed := getFeedByID(feedID, &api.CFG.Feeds)
	if feed == nil {
		return ctx.String(http.StatusNotFound, "Could not find "+feedID)
	}

	posts, err := getPostsForFeed(api.Loaders, feed)
	if err != nil {
		return err
	}

	resp, err := json.Marshal(posts)
	if err != nil {
		return err
	}

	setGlobalHeaders(ctx)
	return ctx.String(http.StatusOK, string(resp))
}

// GetGroupGroupID returns a group by it's groupID
func (api *API) GetGroupGroupID(ctx echo.Context, groupID string) error {
	group := getGroupByID(groupID, &api.CFG.Groups)
	if group == nil {
		return ctx.String(http.StatusNotFound, "Could not find "+groupID)
	}

	resp, err := json.Marshal(group)
	if err != nil {
		return err
	}

	setGlobalHeaders(ctx)
	return ctx.String(http.StatusOK, string(resp))
}

// GetGroups returns all groups
func (api *API) GetGroups(ctx echo.Context) error {
	resp, err := json.Marshal(api.CFG.Groups)
	if err != nil {
		return err
	}

	setGlobalHeaders(ctx)
	return ctx.String(http.StatusOK, string(resp))
}

// GetGroupsGroupIDFeeds returns a list of feeds associated with a given group ID
func (api *API) GetGroupsGroupIDFeeds(ctx echo.Context, groupID string) error {
	feeds := getFeedsForGroupID(groupID, &api.CFG.Feeds)
	if feeds == nil {
		return ctx.String(http.StatusNotFound, "No feeds are available")
	}

	resp, err := json.Marshal(feeds)
	if err != nil {
		return err
	}

	setGlobalHeaders(ctx)
	return ctx.String(http.StatusOK, string(resp))
}

type getGroupGroupIDPostsResponse struct {
	Group  types.Group
	Posts  []types.Post
	Errors []types.Error
}

// GetGroupGroupIDPosts returns a list of posts associated with a given group
func (api *API) GetGroupGroupIDPosts(ctx echo.Context, groupID string) error {
	group := getGroupByID(groupID, &api.CFG.Groups)
	if group == nil {
		return ctx.String(http.StatusNotFound, fmt.Sprintf("Could not find group with %s", groupID))
	}

	feeds := getFeedsForGroupID(groupID, &api.CFG.Feeds)
	if feeds == nil {
		return ctx.String(http.StatusOK, "[]")

	}

	ch := make(chan feedErrors)
	for _, feed := range feeds {
		go func(feedToLoad types.Feed) {
			posts, err := getPostsForFeed(api.Loaders, &feedToLoad)

			var errorReturn types.Error
			if err != nil {
				ids := []string{feedToLoad.ID}
				errMessage := err.Error()
				errorReturn = types.Error{
					Code:       errorFeedLoadFail,
					Message:    errMessage,
					RelatedIDs: ids,
				}
			}

			ch <- feedErrors{
				Posts: posts,
				Err:   errorReturn,
			}
		}(feed)
	}

	posts := []types.Post{}
	errors := []types.Error{}
	for range feeds {
		out := <-ch
		posts = append(posts, *out.Posts...)
		errors = append(errors, out.Err)
	}

	resp, err := json.Marshal(getGroupGroupIDPostsResponse{
		*group,
		posts,
		errors,
	})
	if err != nil {
		return err
	}

	setGlobalHeaders(ctx)
	return ctx.String(http.StatusOK, string(resp))
}

func (api *API) GetRedditgalleryGalleryID(ctx echo.Context, galleryID string) error {
	client := &http.Client{}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://www.reddit.com/gallery/%s", galleryID), nil)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	req.Header.Add("User-Agent", "linux:metacomposite:v0.0.1 (by /u/dankweedhacker)")
	resp, err := client.Do(req)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	galleryURLs := []string{}
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists && strings.Contains(href, "preview.redd.it") {
			galleryURLs = append(galleryURLs, html.UnescapeString(href))
		}
	})

	jsonResp, err := json.Marshal(galleryURLs)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, err.Error())
	}

	setGlobalHeaders(ctx)
	return ctx.String(http.StatusOK, string(jsonResp))
}
