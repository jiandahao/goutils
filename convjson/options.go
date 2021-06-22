package convjson

// GetOption configures how we get the value.
type GetOption struct {
	delimiter string
}

func newGetOption() *GetOption {
	return &GetOption{
		delimiter: ".",
	}
}

func (opt *GetOption) load(options ...GetOptionFunc) {
	for _, o := range options {
		o(opt)
	}
}

// GetOptionFunc option func
type GetOptionFunc func(opt *GetOption)

// WithDelimiter let you set the delimiter, this determines how to split json path.
func WithDelimiter(d string) GetOptionFunc {
	return func(opt *GetOption) {
		opt.delimiter = d
	}
}
