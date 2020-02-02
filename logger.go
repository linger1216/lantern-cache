package lantern_cache

import (
	"log"
	"os"
)

// Logger is invoked when `Config.Verbose=true`
type Logger interface {
	Printf(format string, v ...interface{})
}

// this is a safeguard, breaking on compile time in case
// `log.Logger` does not adhere to our `Logger` interface.
// see https://golang.org/doc/faq#guarantee_satisfies_interface
var _ Logger = &log.Logger{}

// DefaultLogger returns a `Logger` implementation
// backed by stdlib's log
func DefaultLogger() Logger {
	return log.New(os.Stdout, "", log.LstdFlags)
}

type noneLogger struct {
}

func (n *noneLogger) Printf(format string, v ...interface{}) {
}

func NoneLogger() Logger {
	return &noneLogger{}
}
