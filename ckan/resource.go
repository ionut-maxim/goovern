package ckan

type Resource struct {
	CacheLastUpdated     any    `json:"cache_last_updated"`
	DatagovroDownloadUrl string `json:"datagovro_download_url"`
	PackageId            string `json:"package_id"`
	DatastoreActive      bool   `json:"datastore_active"`
	Id                   string `json:"id"`
	Size                 int    `json:"size"`
	State                string `json:"state"`
	Archiver             string `json:"archiver"`
	Hash                 string `json:"hash"`
	Description          string `json:"description"`
	Format               string `json:"format"`
	LastModified         string `json:"last_modified"`
	UrlType              string `json:"url_type"`
	Mimetype             string `json:"mimetype"`
	CacheUrl             any    `json:"cache_url"`
	Name                 string `json:"name"`
	Created              string `json:"created"`
	Url                  string `json:"url"`
	Qa                   string `json:"qa"`
	MimetypeInner        any    `json:"mimetype_inner"`
	Position             int    `json:"position"`
	RevisionId           string `json:"revision_id"`
	ResourceType         any    `json:"resource_type"`
}
