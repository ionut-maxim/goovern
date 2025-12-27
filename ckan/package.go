package ckan

type Package struct {
	Rating                 float64      `json:"rating"`
	LicenseTitle           string       `json:"license_title"`
	Maintainer             string       `json:"maintainer"`
	RelationshipsAsObject  []any        `json:"relationships_as_object"`
	Private                bool         `json:"private"`
	MaintainerEmail        any          `json:"maintainer_email"`
	NumTags                int          `json:"num_tags"`
	Id                     string       `json:"id"`
	MetadataCreated        string       `json:"metadata_created"`
	MetadataModified       string       `json:"metadata_modified"`
	Author                 string       `json:"author"`
	AuthorEmail            string       `json:"author_email"`
	State                  string       `json:"state"`
	Version                string       `json:"version"`
	CreatorUserId          string       `json:"creator_user_id"`
	Type                   string       `json:"type"`
	Resources              []Resource   `json:"resources"`
	NumResources           int          `json:"num_resources"`
	Tags                   []Tag        `json:"tags"`
	Groups                 []any        `json:"groups"`
	LicenseId              string       `json:"license_id"`
	RelationshipsAsSubject []any        `json:"relationships_as_subject"`
	Organization           Organization `json:"organization"`
	Name                   string       `json:"name"`
	Isopen                 bool         `json:"isopen"`
	Url                    any          `json:"url"`
	Notes                  string       `json:"notes"`
	OwnerOrg               string       `json:"owner_org"`
	Extras                 []any        `json:"extras"`
	LicenseUrl             string       `json:"license_url"`
	RatingsCount           int          `json:"ratings_count"`
	Title                  string       `json:"title"`
	RevisionId             string       `json:"revision_id"`
}
