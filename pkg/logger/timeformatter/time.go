package timeformatter

import (
	"time"
)

const (
	GlobalTimeLayout = "2006-01-02T15:04:05.999Z"
	TimeLayout       = "02-01-2006 15:04:05.999Z"
)

func GlobalTimeFormatter(t time.Time) string {
	return t.UTC().Format(GlobalTimeLayout)
}

func TimeFormatter(t time.Time) string {
	return t.UTC().Format(TimeLayout)
}
