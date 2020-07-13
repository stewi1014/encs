package encio

import (
	"io"
	"os"
)

// Warnings is where warnings are sent to.
// In many cases Encs will continue to operate with e.g. incorrectly implemented io.Readers or io.Writers,
// however I don't want to silently put up with things that seem worrying.
var Warnings io.Writer = os.Stderr
