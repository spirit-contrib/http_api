package http_json_api

import (
	"time"
)

var (
	DefaultTimeout time.Duration = 30 * time.Second

	DefaultApiHeader          = "X-Api"
	DefaultApiTimeoutHeader   = "X-Api-Call-Timeout"
	DefaultApiMultiCallHeader = "X-Api-Multi-Call"
)

const (
	HeaderForwardedPayload = "X-Forwarded-Payload"
)

var internalAllowHeaders = []string{
	"Origin",
	"Content-Type",
	"Authorization",
	"Accept",
	"X-Requested-With",
	"X-Forwarded-Payload",
}
