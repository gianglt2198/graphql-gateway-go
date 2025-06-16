package common

type KeyType string

const (
	KEY_REQUEST_ID     KeyType = "request_id"
	KEY_AUTH_USER_ID   KeyType = "user_id"
	KEY_CONTEXT_LOADER KeyType = "data_loader"
	KEY_TRACE_ID       KeyType = "trace_id"
	KEY_SPAN_ID        KeyType = "span_id"
)
