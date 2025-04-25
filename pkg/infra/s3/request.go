package s3

import "time"

type PresignedUploadURL struct {
	Key           string
	ContentType   string
	ContentLength int64
	Expiry        time.Duration
}
