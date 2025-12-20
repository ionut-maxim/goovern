package ckan

import (
	"context"
	"net/url"
	"strconv"
)

type Package struct {
	Rating                 float64       `json:"rating"`
	LicenseTitle           string        `json:"license_title"`
	Maintainer             string        `json:"maintainer"`
	RelationshipsAsObject  []interface{} `json:"relationships_as_object"`
	Private                bool          `json:"private"`
	MaintainerEmail        interface{}   `json:"maintainer_email"`
	NumTags                int           `json:"num_tags"`
	Id                     string        `json:"id"`
	MetadataCreated        string        `json:"metadata_created"`
	MetadataModified       string        `json:"metadata_modified"`
	Author                 string        `json:"author"`
	AuthorEmail            string        `json:"author_email"`
	State                  string        `json:"state"`
	Version                string        `json:"version"`
	CreatorUserId          string        `json:"creator_user_id"`
	Type                   string        `json:"type"`
	Resources              []Resource    `json:"resources"`
	NumResources           int           `json:"num_resources"`
	Tags                   []Tag         `json:"tags"`
	Groups                 []interface{} `json:"groups"`
	LicenseId              string        `json:"license_id"`
	RelationshipsAsSubject []interface{} `json:"relationships_as_subject"`
	Organization           Organization  `json:"organization"`
	Name                   string        `json:"name"`
	Isopen                 bool          `json:"isopen"`
	Url                    interface{}   `json:"url"`
	Notes                  string        `json:"notes"`
	OwnerOrg               string        `json:"owner_org"`
	Extras                 []interface{} `json:"extras"`
	LicenseUrl             string        `json:"license_url"`
	RatingsCount           int           `json:"ratings_count"`
	Title                  string        `json:"title"`
	RevisionId             string        `json:"revision_id"`
}

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
