package ckan

type Tag struct {
	VocabularyId any    `json:"vocabulary_id"`
	State        string `json:"state"`
	DisplayName  string `json:"display_name"`
	Id           string `json:"id"`
	Name         string `json:"name"`
}
