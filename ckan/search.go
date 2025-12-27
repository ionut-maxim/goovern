package ckan

import (
	"context"
	"net/url"
	"strconv"
)

type Search struct {
	Count  int    `json:"count"`
	Sort   string `json:"sort"`
	Facets struct {
	} `json:"facets"`
	Results      []Package `json:"results"`
	SearchFacets struct {
	} `json:"search_facets"`
}

func (c *Client) Search(ctx context.Context, organization string, limit int) (*Search, error) {
	u := c.url.JoinPath("package_search")
	u.RawQuery = url.Values{
		"fq":   []string{"organization:" + organization},
		"sort": []string{"metadata_modified desc"},
		"rows": []string{strconv.Itoa(limit)},
	}.Encode()

	return doRequest[Search](ctx, c, u)
}
