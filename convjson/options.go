package convjson

// Option configures how we get the value.
type Option struct {
	delimiter string
}

func newOption() *Option {
	return &Option{
		delimiter: ".",
	}
}

func (opt *Option) load(options ...OptionFunc) {
	for _, o := range options {
		o(opt)
	}
}

// OptionFunc option func
type OptionFunc func(opt *Option)

// WithDelimiter let you set the delimiter, this determines how to split json path.
func WithDelimiter(d string) OptionFunc {
	return func(opt *Option) {
		opt.delimiter = d
	}
}
