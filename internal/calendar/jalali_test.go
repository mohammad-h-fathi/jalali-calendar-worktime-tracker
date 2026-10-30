package calendar

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestGregorianToJalali(t *testing.T) {
	tests := []struct {
		name       string
		gy, gm, gd int
		expectedJY int
		expectedJM int
		expectedJD int
	}{
		{"Nowruz 1403", 2024, 3, 20, 1403, 1, 1},
		{"Nowruz 1404", 2025, 3, 21, 1404, 1, 1},
		{"Today 2026-10-26", 2026, 10, 26, 1405, 8, 4},
		{"End of Esfand 1402", 2024, 3, 19, 1402, 12, 29},
		{"Mid summer 1403", 2024, 7, 21, 1403, 4, 31},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jy, jm, jd := GregorianToJalali(tt.gy, tt.gm, tt.gd)
			if jy != tt.expectedJY || jm != tt.expectedJM || jd != tt.expectedJD {
				t.Errorf("GregorianToJalali(%d, %d, %d) = (%d, %d, %d); expected (%d, %d, %d)",
					tt.gy, tt.gm, tt.gd, jy, jm, jd, tt.expectedJY, tt.expectedJM, tt.expectedJD)
			}

			// Check round-trip conversion
			rgy, rgm, rgd := JalaliToGregorian(jy, jm, jd)
			if rgy != tt.gy || rgm != tt.gm || rgd != tt.gd {
				t.Errorf("Roundtrip JalaliToGregorian(%d, %d, %d) = (%d, %d, %d); expected (%d, %d, %d)",
					jy, jm, jd, rgy, rgm, rgd, tt.gy, tt.gm, tt.gd)
			}
		})
	}
}

func TestFromTime(t *testing.T) {
	sample := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)
	jDate := FromTime(sample)
	if jDate.Year != 1405 || jDate.Month != 8 || jDate.Day != 4 {
		t.Fatalf("Unexpected date: %v", jDate)
	}
	// Monday in Persian calendar is Doshanbeh (index 2)
	if jDate.Weekday != 2 {
		t.Fatalf("Expected weekday 2 (دوشنبه), got %d (%s)", jDate.Weekday, jDate.WeekdayName())
	}
}

func TestHolidays(t *testing.T) {
	// Use stub HTTPClient to test offline events without network call
	oldClient := HTTPClient
	defer func() { HTTPClient = oldClient }()
	HTTPClient = &mockErrorHTTPClient{}

	events := GetEventsForDate(1403, 1, 1)
	if len(events) == 0 || !events[0].IsHoliday {
		t.Fatalf("Expected Nowruz holiday on 1 Farvardin, got %v", events)
	}

	// Verify lunar holiday from offline data (Tasua on 25 Tir 1403)
	lunarEvents := GetEventsForDate(1403, 4, 25)
	foundTasua := false
	for _, ev := range lunarEvents {
		if ev.Title == "تاسوعای حسینی" && ev.IsHoliday {
			foundTasua = true
			break
		}
	}
	if !foundTasua {
		t.Fatalf("Expected Tasua lunar holiday on 25 Tir 1403, got %v", lunarEvents)
	}
}

type mockErrorHTTPClient struct{}

func (m *mockErrorHTTPClient) Get(url string) (*http.Response, error) {
	return nil, fmt.Errorf("network disabled in tests")
}

func TestIsLeapYear(t *testing.T) {
	tests := []struct {
		year     int
		isLeap   bool
		daysEsf  int
	}{
		{1399, true, 30},
		{1402, false, 29},
		{1403, true, 30},
		{1404, false, 29},
		{1436, true, 30},
		{1437, false, 29},
		{1469, true, 30},
		{1470, false, 29},
	}

	for _, tt := range tests {
		got := IsLeapYear(tt.year)
		if got != tt.isLeap {
			t.Errorf("IsLeapYear(%d) = %v; want %v", tt.year, got, tt.isLeap)
		}
		days := DaysInMonth(tt.year, 12)
		if days != tt.daysEsf {
			t.Errorf("DaysInMonth(%d, 12) = %d; want %d", tt.year, days, tt.daysEsf)
		}
	}
}

func TestParseISO(t *testing.T) {
	valid := []string{"1403-12-30", "1404-01-15", "1405/08/04", " 1405-01-01 "}
	for _, s := range valid {
		d, err := ParseISO(s)
		if err != nil {
			t.Errorf("ParseISO(%q) unexpected error: %v", s, err)
		}
		if d.Year == 0 || d.Month == 0 || d.Day == 0 {
			t.Errorf("ParseISO(%q) invalid output: %+v", s, d)
		}
	}

	invalid := []string{
		"1404-12-30", // 1404 is common year, Esfand has 29 days
		"1405-13-01", // month 13
		"1405-02-32", // day 32
		"1405-07-31", // Mehr only has 30 days
		"bad-string",
		"1405",
	}
	for _, s := range invalid {
		if _, err := ParseISO(s); err == nil {
			t.Errorf("ParseISO(%q) expected error for invalid date, got nil", s)
		}
	}
}
