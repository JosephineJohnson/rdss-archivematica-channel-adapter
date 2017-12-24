package message

import (
	"time"
)

// Timestamp is similar to time.Time but implementing the formatting specifics
// described in the API: see https://git.io/v5obt for more details.
type Timestamp time.Time

func (t Timestamp) MarshalJSON() ([]byte, error) {
	ts := time.Time(t)
	if ts.IsZero() {
		return []byte("null"), nil
	}
	return time.Time(t).MarshalJSON()
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// The time is expected to be a quoted string in RFC 3339 format.
func (t *Timestamp) UnmarshalJSON(data []byte) error {
	str := string(data)
	// Ignore null, like in the main JSON package.
	if str == "null" {
		return nil
	}
	// Ignore emtpy string.
	if str == "\"\"" {
		return nil
	}
	// Fractional seconds are handled implicitly by Parse.
	ts, err := time.Parse(`"`+time.RFC3339+`"`, str)
	if err == nil {
		*t = Timestamp(ts)
		return nil
	}
	// Like time.RFC3339 without the second component which is accepted by RDSS.
	const RFC3339woSec = "2006-01-02T15:04Z07:00"
	ts, err = time.Parse(`"`+RFC3339woSec+`"`, str)
	if err == nil {
		*t = Timestamp(ts)
		return nil
	}
	return err
}

func (t Timestamp) String() string {
	return time.Time(t).String()
}
