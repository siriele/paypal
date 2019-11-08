package paypal

import (
	"encoding/json"
	"testing"
	"time"
)

func TestPTime(t *testing.T) {
	type TestDate struct {
		Date PTime `json:"date"`
	}
	// "2018-08-15T19:14:04.543Z"
	// 2018-08-15T12:13:29-07:00"

	t.Run("Regular date should work", func(t *testing.T) {
		dateObject := new(TestDate)
		data := []byte(`{"date":"2018-08-15T19:14:04.543Z"}`)
		if jerr := json.Unmarshal(data, dateObject); jerr != nil {
			t.Fatalf("Failed to parse test date %s", jerr.Error())
		}
		expected := time.Date(2018, time.August, 15, 19, 14, 4, int(543*time.Millisecond), time.UTC)
		if !expected.Equal(dateObject.Date.Time) {
			t.Fatalf("Extpected %v to equal %v", expected, dateObject.Date.Time.String())
		}
	})

	t.Run("Should choose alternative date when there is no Z", func(t *testing.T) {
		dateObject := new(TestDate)
		data := []byte(`{"date":"2018-08-15T12:13:29-07:00"}`)
		if jerr := json.Unmarshal(data, dateObject); jerr != nil {
			t.Fatalf("Failed to parse test date %s", jerr.Error())
		}
		if _, offset := dateObject.Date.Time.Zone(); offset != -7*3600 {
			t.Fatalf("Extpected %v but got %v", -7*3600, offset)
		}
	})

	t.Run("Regular date should work with not . ", func(t *testing.T) {
		dateObject := new(TestDate)
		data := []byte(`{"date":"2018-08-15T19:14:04Z"}`)
		if jerr := json.Unmarshal(data, dateObject); jerr != nil {
			t.Fatalf("Failed to parse test date %s", jerr.Error())
		}
		expected := time.Date(2018, time.August, 15, 19, 14, 4, 0, time.UTC)
		if !expected.Equal(dateObject.Date.Time) {
			t.Fatalf("Extpected %v to equal %v", expected, dateObject.Date.Time.String())
		}
	})

}
