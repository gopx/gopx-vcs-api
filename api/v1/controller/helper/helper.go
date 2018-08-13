package helper

import (
	"fmt"
	"time"
)

// FormatTime formats time to append date to any path.
func FormatTime(t time.Time) string {
	return fmt.Sprintf(
		"%d_%02d_%02dT%02d_%02d_%02d",
		t.Year(),
		t.Month(),
		t.Day(),
		t.Hour(),
		t.Minute(),
		t.Second(),
	)
}
