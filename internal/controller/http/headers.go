package http

import (
	"net/textproto"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

func customIncomingHeaderMatcher(key string) (string, bool) {
	key = textproto.CanonicalMIMEHeaderKey(key)
	return runtime.DefaultHeaderMatcher(key)
}

func customOutgoingHeaderMatcher(key string) (string, bool) {
	switch key {
	case
		"accept",
		"accept-encoding",
		"accept-language",
		"content-length",
		"content-disposition",
		"content-type",
		"date",
		"origin",
		"x-request-id":
		return key, true
	default:
		return key, false
	}
}
