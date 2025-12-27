package ckan

import (
	"github.com/google/uuid"
)

type Resource struct {
	Id                   uuid.UUID     `json:"id" db:"id"`
	PackageId            uuid.NullUUID `json:"package_id,omitempty" db:"package_id"`
	Name                 string        `json:"name" db:"name"`
	Description          string        `json:"description" db:"description"`
	Url                  string        `json:"url" db:"url"`
	UrlType              string        `json:"url_type" db:"url_type"`
	Format               string        `json:"format" db:"format"`
	Mimetype             string        `json:"mimetype" db:"mimetype"`
	MimetypeInner        any           `json:"mimetype_inner" db:"mimetype_inner"`
	Size                 int           `json:"size" db:"size"`
	Hash                 string        `json:"hash" db:"hash"`
	State                string        `json:"state" db:"state"`
	Position             int           `json:"position" db:"position"`
	ResourceType         any           `json:"resource_type" db:"resource_type"`
	Created              Time          `json:"created" db:"created"`
	LastModified         Time          `json:"last_modified" db:"last_modified"`
	CacheLastUpdated     Time          `json:"cache_last_updated" db:"cache_last_updated"`
	CacheUrl             any           `json:"cache_url" db:"cache_url"`
	DatagovroDownloadUrl string        `json:"datagovro_download_url" db:"datagovro_download_url"`
	Qa                   string        `json:"qa" db:"-"`
	Archiver             string        `json:"archiver" db:"-"`
	DatastoreActive      bool          `json:"datastore_active" db:"datastore_active"`
	RevisionId           uuid.NullUUID `json:"revision_id,omitempty" db:"revision_id"`
}
