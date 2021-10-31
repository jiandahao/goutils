package convhttp

import (
	"net/http"

	"go.uber.org/zap"
)

// ClientOption client option.
type ClientOption func(c *Client)

// WithLogger returns a ClientOption that specifies the logger for logging.
func WithLogger(logger *zap.Logger) ClientOption {
	return func(c *Client) {
		c.Logger = logger
	}
}

// RequestInterceptor represents a request interceptor.
type RequestInterceptor func(opts *RequestOptions)

// WithRequestInterceptors returns a ClientOption that specifies the
// interceptors for request.
func WithRequestInterceptors(interceptors ...RequestInterceptor) ClientOption {
	return func(c *Client) {
		c.requestInterceptors = append(c.requestInterceptors, interceptors...)
	}
}

// ResponseInterceptor intercepts the response. ResponseInterceptor counld be specified as a
// customize error handler. Errors returned from ResponseInterceptor will be set into Response.err field.
type ResponseInterceptor func(resp *Response) error

// WithResponseInterceptors returns a ClientOption that specifies the
// interceptors for handling response right before sending response data to caller.
func WithResponseInterceptors(interceptors ...ResponseInterceptor) ClientOption {
	return func(c *Client) {
		c.responseInterceptors = append(c.responseInterceptors, interceptors...)
	}
}

// WithHTTPClient returns a ClientOption that specifies the http client for sending
// http request.
func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = hc
	}
}
