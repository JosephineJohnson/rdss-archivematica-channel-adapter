package version

import "fmt"

// VERSION is the app-global version string, which should be substituted with a
// real value during build.
var VERSION = "UNKNOWN"

// Name of this application.
const Name = "RDSS Archivematica Channel Adapter"

func AppVersion() string {
	return fmt.Sprintf("%s %s", Name, VERSION)
}
