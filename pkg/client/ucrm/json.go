package ucrm

import (
	"encoding/json"
	"fmt"
	"time"
)

type jsonDate struct {
	time.Time
}

func (t jsonDate) MarshalJSON() ([]byte, error) {
	date := fmt.Sprintf("%sT00:00:00+0000", t.Format("2006-01-02"))
	return json.Marshal(date)
}

type jsonDateTime struct {
	time.Time
}

func (t jsonDateTime) MarshalJSON() ([]byte, error) {
	date := t.Format("2006-01-02T15:04:05") + "+0000"
	return json.Marshal(date)
}
