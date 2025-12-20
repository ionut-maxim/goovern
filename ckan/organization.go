package ckan

import (
	"context"
	"net/url"
)

type User struct {
	EmailHash                         string `json:"email_hash"`
	About                             string `json:"about"`
	Capacity                          string `json:"capacity"`
	Name                              string `json:"name"`
	Created                           string `json:"created"`
	Sysadmin                          bool   `json:"sysadmin"`
	ActivityStreamsEmailNotifications bool   `json:"activity_streams_email_notifications"`
	State                             string `json:"state"`
	NumberOfEdits                     int    `json:"number_of_edits"`
	DisplayName                       string `json:"display_name"`
	Fullname                          string `json:"fullname"`
	ID                                string `json:"id"`
	NumberCreatedPackages             int    `json:"number_created_packages"`
}

type Organization struct {
	Users           []User `json:"users"`
	DisplayName     string `json:"display_name"`
	Description     string `json:"description"`
	ImageDisplayURL string `json:"image_display_url"`
	PackageCount    int    `json:"package_count"`
	Created         string `json:"created"`
	Name            string `json:"name"`
	IsOrganization  bool   `json:"is_organization"`
	State           string `json:"state"`
	Extras          []any  `json:"extras"`
	ImageURL        string `json:"image_url"`
	Groups          []any  `json:"groups"`
	Type            string `json:"type"`
	Title           string `json:"title"`
	RevisionID      string `json:"revision_id"`
	NumFollowers    int    `json:"num_followers"`
	ID              string `json:"id"`
	Tags            []any  `json:"tags"`
	ApprovalStatus  string `json:"approval_status"`
}

func (c *Client) Organization(ctx context.Context, id string) (*Organization, error) {
	reqUrl := c.url.JoinPath("organization_show")
	reqUrl.RawQuery = url.Values{
		"id": []string{id},
	}.Encode()

	return doRequest[Organization](ctx, c, reqUrl)
}
