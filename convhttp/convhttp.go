package convhttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"go.uber.org/zap"
)

// RequestOptions request options
type RequestOptions struct {
	Method  string      `json:"method,omitempty"`
	URL     string      `json:"url,omitempty"`
	Header  http.Header `json:"header,omitempty"`
	Query   url.Values  `json:"query,omitempty"`
	Request interface{} `json:"request,omitempty"`

	aborted       bool
	abortedReason string
}

// Abort abort current request.
func (opts *RequestOptions) Abort(reason string) {
	opts.aborted = true
	opts.abortedReason = reason
}

// Response response
type Response struct {
	*http.Response
	Request *RequestOptions
	Body    []byte
	err     error // deferred error for easy chaining
}

// Error returns the error
func (resp *Response) Error() error {
	return resp.err
}

// ShouldBindJSON is a shortcut for resp.ShouldBindWith(obj, &JSONBinder{}).
func (resp *Response) ShouldBindJSON(obj interface{}) error {
	return resp.ShouldBindWith(obj, &JSONBinder{})
}

// ShouldBindWith binds the passed struct pointer using the specified binding engine.
func (resp *Response) ShouldBindWith(obj interface{}, binder Binder) (err error) {
	if resp.err != nil {
		return resp.err
	}

	if binder == nil {
		return fmt.Errorf("invalid binder")
	}

	defer func() {
		if err != nil {
			resp.err = err
		}
	}()

	err = binder.Bind(resp.Body, obj)
	if err != nil {
		err = fmt.Errorf("failed to parse %v with binder %s", reflect.ValueOf(obj).Type(), binder.Name())
		return
	}

	return nil
}

// DefaultClient default client
var DefaultClient = newDefaultClient()

func newDefaultClient() *Client {
	return &Client{
		httpClient: http.DefaultClient,
		Logger:     nil,
	}
}

// Client client
type Client struct {
	httpClient *http.Client
	Logger     *zap.Logger

	requestInterceptors  []RequestInterceptor
	responseInterceptors []ResponseInterceptor
}

// NewClient creates a new client object.
func NewClient(options ...ClientOption) *Client {
	cc := &Client{
		httpClient: &http.Client{},
	}

	for _, option := range options {
		option(cc)
	}

	return cc
}

// Do sends an HTTP request based on RequestOptions.
func (c *Client) Do(opts *RequestOptions) (resp *Response) {
	defer func() {
		if r := recover(); r != nil {
			resp.err = fmt.Errorf("recover from panic: %v", r)
		}

		if c.Logger == nil {
			return
		}

		if resp.err != nil {
			c.Logger.Error("handle request", zap.Any("options", opts), zap.Error(resp.err))
			return
		}

		c.Logger.Info("handle request", zap.Any("options", opts))
	}()

	for _, interceptor := range c.requestInterceptors {
		interceptor(opts)
		if opts.aborted {
			return &Response{
				err: fmt.Errorf("request aborted, reason: %s", opts.abortedReason),
			}
		}
	}

	resp = c.do(opts)
	if resp.err != nil {
		return
	}

	for _, interceptor := range c.responseInterceptors {
		if err := interceptor(resp); err != nil {
			resp.err = err
			return resp
		}
	}

	return
}

func (c *Client) do(opts *RequestOptions) (resp *Response) {
	var err error
	resp = &Response{
		Request: opts,
	}

	defer func() {
		resp.err = err
	}()

	if opts == nil {
		err = fmt.Errorf("invalid request options")
		return
	}

	if err = opts.validate(); err != nil {
		return
	}

	req, err := opts.makeRequest()
	if err != nil {
		return
	}

	var res *http.Response
	res, err = c.httpClient.Do(req)
	if err != nil {
		return
	}

	defer res.Body.Close()

	// make response
	resp.Response = res
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	resp.Body = data

	return
}

func (opts *RequestOptions) validate() error {
	if opts.Header == nil {
		opts.Header = http.Header{}
	}

	if opts.Method == "" {
		opts.Method = http.MethodGet
	}

	if !(strings.HasPrefix(opts.URL, "http://") || strings.HasPrefix(opts.URL, "https://")) {
		return fmt.Errorf("http: invalid request url")
	}

	return nil
}

func (opts *RequestOptions) makeRequest() (*http.Request, error) {
	buffer, err := opts.makeRequestBuffer(opts.Request)
	if err != nil {
		return nil, nil
	}

	req, err := http.NewRequest(opts.Method, opts.URL, buffer)
	if err != nil {
		return nil, err
	}
	req.Header = opts.Header
	req.URL.RawQuery = opts.Query.Encode()
	return req, nil
}

// application/json , application/x-www-form-urlencoded , multipart/form-data

func (opts *RequestOptions) makeRequestBuffer(body interface{}) (io.Reader, error) {
	if body == nil {
		return nil, nil
	}

	switch body.(type) {
	case string:
		return bytes.NewBuffer([]byte(body.(string))), nil
	case []byte:
		return bytes.NewBuffer(body.([]byte)), nil
	default:
	}

	if _, ok := body.(*FormData); ok {
		return opts.makeFormDataBuffer()
	}

	// assume that body is json serializable
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body with errror: %s", err)
	}

	return bytes.NewBuffer(data), nil
}

func (opts *RequestOptions) makeFormDataBuffer() (io.Reader, error) {
	fd, ok := opts.Request.(*FormData)
	if !ok {
		return nil, fmt.Errorf("invalid form data")
	}

	if fd == nil {
		return nil, nil
	}

	if fd.file == nil {
		opts.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		return strings.NewReader(fd.Values.Encode()), nil
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for k, vs := range fd.Values {
		for index := range vs {
			writer.WriteField(k, vs[index])
		}
	}

	fw, err := writer.CreateFormFile(fd.file.fieldname, fd.file.filename)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(fw, bytes.NewBuffer(fd.file.data))
	if err != nil {
		return nil, err
	}

	writer.Close()

	opts.Header.Set("Content-Type", writer.FormDataContentType())
	return body, nil
}
