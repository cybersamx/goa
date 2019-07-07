package goa

import "log"

// Standard error printing
func logf(logger *log.Logger, format string, args ...interface{}) {
	logger.Printf(format, args)
}
