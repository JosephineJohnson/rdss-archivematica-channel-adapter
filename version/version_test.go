package version

import (
	"testing"
)

func TestAppVersion(t *testing.T) {
	VERSION = "v1.2.3"
	if have, want := AppVersion(), "RDSS Archivematica Channel Adapter v1.2.3"; have != want {
		t.Errorf("AppVersion() have %v want %v", have, want)
	}
}
