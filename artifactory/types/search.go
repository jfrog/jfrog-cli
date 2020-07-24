package types

type SearchResult struct {
	Path     string              `json:"path,omitempty"`
	Type     string              `json:"type,omitempty"`
	Size     int64               `json:"size,omitempty"`
	Created  string              `json:"created,omitempty"`
	Modified string              `json:"modified,omitempty"`
	Sha1     string              `json:"sha1,omitempty"`
	Md5      string              `json:"md5,omitempty"`
	Props    map[string][]string `json:"props,omitempty"`
}
