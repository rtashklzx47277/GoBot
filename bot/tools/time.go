package tools

import (
	"fmt"
	"time"
)

type Time time.Time
type Duration time.Duration

func (t Time) InRange(hour int) bool {
	return time.Now().UTC().Sub(time.Time(t)) <= time.Duration(hour)*time.Hour
}

func (t Time) Stamp() int {
	return int(time.Time(t).Unix())
}

func (t Time) Format(layout string) string {
	return time.Time(t).Format(layout)
}

func (t1 Time) After(t2 Time) bool {
	return time.Time(t1).After(time.Time(t2))
}

func (t Time) UTC() Time {
	return Time(time.Time(t).UTC())
}

func (t Time) String(format ...string) string {
	if t == (Time{}) {
		return ""
	}

	var ts string

	if len(format) == 0 {
		ts = t.Format("2006-01-02 15:04:05")
	} else {
		switch format[0] {
		case "time":
			ts = fmt.Sprintf("<t:%d:T>", t.Stamp())
		case "date":
			ts = fmt.Sprintf("<t:%d:D>", t.Stamp())
		case "full":
			ts = fmt.Sprintf("<t:%d:f>", t.Stamp())
		case "relative":
			ts = fmt.Sprintf("<t:%d:R>", t.Stamp())
		}
	}

	return ts
}

func (d Duration) Seconds() int {
	return int(time.Duration(d).Seconds()) % 60
}

func (d Duration) Minutes() int {
	return int(time.Duration(d).Minutes()) % 60
}

func (d Duration) Hours() int {
	return int(time.Duration(d).Hours())
}

func (d Duration) String(format ...string) string {
	if d == Duration(0) {
		return "00:00:00"
	}

	var ds string

	if len(format) == 0 {
		ds = fmt.Sprintf("%02d:%02d:%02d", d.Hours(), d.Minutes(), d.Seconds())
	} else {
		ds = fmt.Sprintf("%02d時%02d分%02d秒", d.Hours(), d.Minutes(), d.Seconds())
	}

	return ds
}
