package calendar

import (
	"fmt"
	"strings"
	"time"
)

// MonthNames lists the 12 Persian Solar Hijri month names.
var MonthNames = [12]string{
	"فروردین", "اردیبهشت", "خرداد",
	"تیر", "مرداد", "شهریور",
	"مهر", "آبان", "آذر",
	"دی", "بهمن", "اسفند",
}

// WeekdayNames lists the days of the week in Persian starting from Saturday.
var WeekdayNames = [7]string{
	"شنبه", "یکشنبه", "دوشنبه", "سه‌شنبه", "چهارشنبه", "پنج‌شنبه", "جمعه",
}

// WeekdayShortNames lists short names for calendar headers.
var WeekdayShortNames = [7]string{
	"ش", "ی", "د", "س", "چ", "پ", "ج",
}

// JalaliDate represents a date in the Persian Solar Hijri calendar.
type JalaliDate struct {
	Year    int `json:"year"`
	Month   int `json:"month"`   // 1 - 12
	Day     int `json:"day"`     // 1 - 31
	Weekday int `json:"weekday"` // 0: شنبه, ..., 6: جمعه
}

// IsLeapYear returns true if the specified Jalali year is a leap year.
func IsLeapYear(jy int) bool {
	// Standard astronomical / Birashk 33-year cycle approximation
	rem := (jy + 38) * 31 % 128
	return rem < 31
}

// DaysInMonth returns number of days in the specified Jalali year and month.
func DaysInMonth(jy, jm int) int {
	if jm < 1 || jm > 12 {
		return 0
	}
	if jm <= 6 {
		return 31
	}
	if jm <= 11 {
		return 30
	}
	if IsLeapYear(jy) {
		return 30
	}
	return 29
}

// GregorianToJalali converts Gregorian (gy, gm, gd) to Jalali (jy, jm, jd).
// Algorithm by Roozbeh Pournader and Mohammad Toossi.
func GregorianToJalali(gy, gm, gd int) (jy, jm, jd int) {
	gDaysInMonths := [12]int{0, 31, 59, 90, 120, 151, 181, 212, 243, 273, 304, 334}
	var gy2 int
	if gm > 2 {
		gy2 = gy + 1
	} else {
		gy2 = gy
	}

	days := 355666 + (365 * gy) + (gy2+3)/4 - (gy2+99)/100 + (gy2+399)/400 + gd + gDaysInMonths[gm-1]
	jy = -1595 + (33 * (days / 12053))
	days %= 12053
	jy += 4 * (days / 1461)
	days %= 1461

	if days > 365 {
		jy += (days - 1) / 365
		days = (days - 1) % 365
	}

	if days < 186 {
		jm = 1 + days/31
		jd = 1 + (days % 31)
	} else {
		jm = 7 + (days-186)/30
		jd = 1 + ((days - 186) % 30)
	}

	return jy, jm, jd
}

// JalaliToGregorian converts Jalali (jy, jm, jd) to Gregorian (gy, gm, gd).
// Algorithm by Roozbeh Pournader and Mohammad Toossi.
func JalaliToGregorian(jy, jm, jd int) (gy, gm, gd int) {
	jy += 1595
	var daysBeforeMonth int
	if jm < 7 {
		daysBeforeMonth = (jm - 1) * 31
	} else {
		daysBeforeMonth = (jm-7)*30 + 186
	}

	days := -355668 + (365 * jy) + (jy/33)*8 + ((jy%33)+3)/4 + jd + daysBeforeMonth
	gy = 400 * (days / 146097)
	days %= 146097

	if days > 36524 {
		days--
		gy += 100 * (days / 36524)
		days %= 36524
		if days >= 365 {
			days++
		}
	}

	gy += 4 * (days / 1461)
	days %= 1461

	if days > 365 {
		gy += (days - 1) / 365
		days = (days - 1) % 365
	}

	gd = days + 1

	isLeap := (gy%4 == 0 && gy%100 != 0) || (gy%400 == 0)
	salA := [13]int{0, 31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	if isLeap {
		salA[2] = 29
	}

	for gm = 1; gm < 13 && gd > salA[gm]; gm++ {
		gd -= salA[gm]
	}

	return gy, gm, gd
}

// FromTime converts a standard time.Time to JalaliDate.
func FromTime(t time.Time) JalaliDate {
	jy, jm, jd := GregorianToJalali(t.Year(), int(t.Month()), t.Day())
	// Go's t.Weekday(): 0 is Sunday, 1 is Monday ... 6 is Saturday
	// In Persian calendar: 0 is Saturday, 1 is Sunday ... 6 is Friday
	weekday := (int(t.Weekday()) + 1) % 7

	return JalaliDate{
		Year:    jy,
		Month:   jm,
		Day:     jd,
		Weekday: weekday,
	}
}

// ToTime converts a JalaliDate back to a time.Time in the given timezone.
func (j JalaliDate) ToTime(loc *time.Location) time.Time {
	if loc == nil {
		loc = time.Local
	}
	gy, gm, gd := JalaliToGregorian(j.Year, j.Month, j.Day)
	return time.Date(gy, time.Month(gm), gd, 0, 0, 0, 0, loc)
}

// WeekdayName returns the Persian name of the weekday.
func (j JalaliDate) WeekdayName() string {
	if j.Weekday >= 0 && j.Weekday < 7 {
		return WeekdayNames[j.Weekday]
	}
	return ""
}

// MonthName returns the Persian name of the month.
func (j JalaliDate) MonthName() string {
	if j.Month >= 1 && j.Month <= 12 {
		return MonthNames[j.Month-1]
	}
	return ""
}

// Format returns a formatted date string, e.g. "۵ آبان ۱۴۰۵"
func (j JalaliDate) Format() string {
	return fmt.Sprintf("%s %s %s", ToPersianDigits(fmt.Sprintf("%d", j.Day)), j.MonthName(), ToPersianDigits(fmt.Sprintf("%d", j.Year)))
}

// FormatISO returns an ISO-like Jalali representation: "1405-08-04"
func (j JalaliDate) FormatISO() string {
	return fmt.Sprintf("%04d-%02d-%02d", j.Year, j.Month, j.Day)
}

// ToPersianDigits converts ASCII digits ('0'-'9') to Persian digits ('۰'-'۹').
func ToPersianDigits(s string) string {
	persianDigits := []string{"۰", "۱", "۲", "۳", "۴", "۵", "۶", "۷", "۸", "۹"}
	var sb strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			sb.WriteString(persianDigits[r-'0'])
		} else {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}
