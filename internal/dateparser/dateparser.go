package dateparser

import (
	godateparser "github.com/markusmobius/go-dateparser"
	"strings"
	"time"
)

type DateParser struct {
	CurrentTimeFunc func() time.Time
}

func (d *DateParser) ParseDate(str string) (time.Time, error) {
	str = strings.TrimSpace(str)

	dt, err := godateparser.Parse(&godateparser.Configuration{
		CurrentTime: d.CurrentTimeFunc(),
	}, str)
	if err == nil {
		return dt.Time, nil
	}

	parts := strings.Split(str, " ")
	for len(parts) > 1 {
		newStr := strings.Join(parts, " ")
		dt, err = godateparser.Parse(nil, newStr)
		if err == nil {
			return dt.Time, err
		}
		parts = parts[1:]
	}

	return time.Time{}, err
}
