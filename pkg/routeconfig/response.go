package routeconfig

import "time"

type ResponseItem struct {
	Path          string
	ContentType   string
	ContentLength int64
	LastModified  time.Time
	Content       []byte
	ContentPath   string
	ETag          string
}
