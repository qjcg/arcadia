package blocks

import (
	"time"
)

type TimeStamp struct {
	Fmt string
	time.Time
}

var (
	// Normal mode.
	DefaultTimeStamp = TimeStamp{
		Fmt: "Mon Jan 2, 3:04pm",
	}

	// Test mode (more frequent updates).
	TestTimeStamp = TimeStamp{
		Fmt: "Mon Jan 2, 3:04:05pm MST",
	}
)

// String implements the fmt.Stringer interface.
func (ts TimeStamp) String() string {
	return time.Now().Format(ts.Fmt)
}
