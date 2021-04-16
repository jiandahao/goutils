package logger

var (
	// DefaultLogger default logger
	DefaultLogger Logger = &nilLogger{}
)

// Logger logger interface
type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	Sync() error
}

type nilLogger struct{}

func (lg *nilLogger) Debug(args ...interface{})                   {}
func (lg *nilLogger) Info(args ...interface{})                    {}
func (lg *nilLogger) Warn(args ...interface{})                    {}
func (lg *nilLogger) Error(args ...interface{})                   {}
func (lg *nilLogger) Debugf(template string, args ...interface{}) {}
func (lg *nilLogger) Infof(template string, args ...interface{})  {}
func (lg *nilLogger) Warnf(template string, args ...interface{})  {}
func (lg *nilLogger) Errorf(template string, args ...interface{}) {}
func (lg *nilLogger) Sync() error                                 { return nil }
