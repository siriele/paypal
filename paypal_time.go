package paypal

import (
	"time"
)

const DateFormat = "2006-01-02T15:04:05Z"

type PTime struct{ time.Time }

func (t *PTime) UnmarshalJSON(data []byte) error {
	sdata := string(data)
	if sdata == `null` || sdata == `""` || sdata == `"0"` {
		return nil
	}

	tm, err := time.Parse(DateFormat, sdata[1:len(sdata)-1])
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
