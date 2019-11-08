package paypal

import (
	"strings"
	"time"
)

const (
	DateFormatNoDot      = "2006-01-02T15:04:05Z"
	DateFormat           = "2006-01-02T15:04:05.000Z"
	DateFormatWithOffset = "2006-01-02T15:04:05-07:00"
)

type PTime struct{ time.Time }

func (t *PTime) UnmarshalJSON(data []byte) error {
	sdata := string(data)
	if sdata == `null` || sdata == `""` || sdata == `"0"` {
		return nil
	}

	format := DateFormat
	noZ := !strings.Contains(sdata, "Z")
	noDot := !strings.Contains(sdata, ".")
	switch {
	case noDot && noZ:
		format = DateFormatWithOffset
	case !noDot && !noZ:
		format = DateFormat
	case !noZ && noDot:
		format = DateFormatNoDot

	}

	tm, err := time.Parse(format, sdata[1:len(sdata)-1])
	if err != nil {
		return err
	}
	t.Time = tm
	return nil
}

func (t *PTime) String() string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format(DateFormat)
}
