package calendar

import (
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
	events := GetEventsForDate(1403, 1, 1)
	if len(events) == 0 || !events[0].IsHoliday {
		t.Fatalf("Expected Nowruz holiday on 1 Farvardin, got %v", events)
	}
}
