package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Month int

func CurrentMonth() Month {
	t := time.Now()
	var m Month
	m.SetTime(t)
	return m
}

func MonthFromTime(t time.Time) Month {
	var m Month
	m.SetTime(t)
	return m
}

func MonthFromDate(d Date) Month {
	var m Month
	m.Set(d.Year(), d.Month())
	return m
}

func MonthFromString(str string) (Month, error) {

	if str == "" {
		return CurrentMonth(), nil
	}

	t, err := time.Parse("2006-01", str)
	if err != nil {
		return 0, err
	}
	var m Month
	m.SetTime(t)
	return m, nil
}

func (m *Month) String() string {
	str := fmt.Sprintf("%04d-%02d", m.Year(), m.Month())
	return str
}

func (m *Month) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	var err error
	*m, err = MonthFromString(s)
	return err
}

func (m *Month) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.String())
}

func (m *Month) Year() int {
	year := int(*m) / 100
	return year
}

func (m *Month) Month() int {
	month := int(*m) - (m.Year() * 100)
	return month
}

func (m *Month) Set(year int, month int) {
	*m = Month(year*100 + month)
}

func (m *Month) SetTime(t time.Time) {
	m.Set(t.Year(), int(t.Month()))
}

func (m *Month) Time() time.Time {
	t, _ := time.Parse("2006-01", m.String())
	return t
}

func (m *Month) Prev() Month {
	var prev Month
	month := m.Month()
	year := m.Year()
	if month == 1 {
		prev.Set(year-1, 12)
	} else {
		prev.Set(year, month-1)
	}
	return prev
}

func (m *Month) Next() Month {
	var next Month
	month := m.Month()
	year := m.Year()
	if month == 12 {
		next.Set(year+1, 1)
	} else {
		next.Set(year, month+1)
	}
	return next
}

type MonthData interface {
	GetMonth() Month
	SetMonth(m Month)
}

func MonthFromId(id string) (Month, error) {

	if len(id) != 20 {
		return 0, errors.New("too short ID")
	}

	idTimeStr := id[:8]
	timeInt, err := strconv.ParseInt(idTimeStr, 16, 64)
	if err != nil {
		return 0, errors.New("invalid time in ID")
	}
	tm := time.Unix(timeInt, 0)

	m := MonthFromTime(tm)
	return m, nil
}

type MonthDataBase struct {
	Month Month `gorm:"primary_key;index" json:"month"`
}

func (w *MonthDataBase) InitMonth() {
	w.Month = CurrentMonth()
}

func (w *MonthDataBase) SetMonth(m Month) {
	w.Month = m
}

func (w *MonthDataBase) GetMonth() Month {
	return w.Month
}
