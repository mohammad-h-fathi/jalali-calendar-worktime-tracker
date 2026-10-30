package calendar

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Event represents a holiday or commemorative occasion on the Persian calendar.
type Event struct {
	Month       int    `json:"month"`
	Day         int    `json:"day"`
	Title       string `json:"title"`
	IsHoliday   bool   `json:"is_holiday"`
	Description string `json:"description"`
}

// IranIREvent matches the JSON schema from https://iran.ir/Calendar/GetEvents?year=...&month=...
type IranIREvent struct {
	Title      string `json:"title"`
	JalaliDay  int    `json:"jalali_day"`
	IsHoliday  bool   `json:"is_holiday"`
	DateString string `json:"date_string"`
}

// HTTPGetter defines the interface for making HTTP GET requests.
type HTTPGetter interface {
	Get(url string) (*http.Response, error)
}

var (
	eventsCache   = make(map[string][]Event)
	eventsCacheMu sync.RWMutex
	HTTPClient    HTTPGetter = &http.Client{Timeout: 3 * time.Second}
)

// FetchMonthEvents queries the official iran.ir Calendar API with local caching and offline fallback.
func FetchMonthEvents(year, month int) []Event {
	key := fmt.Sprintf("%d-%d", year, month)

	eventsCacheMu.RLock()
	cached, ok := eventsCache[key]
	eventsCacheMu.RUnlock()
	if ok {
		return cached
	}

	// Attempt remote fetch from https://iran.ir/Calendar/GetEvents?year=...&month=...
	url := fmt.Sprintf("https://iran.ir/Calendar/GetEvents?year=%d&month=%d", year, month)
	resp, err := HTTPClient.Get(url)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err == nil && resp.StatusCode == http.StatusOK {
		var remoteEvents []IranIREvent
		if jsonErr := json.NewDecoder(resp.Body).Decode(&remoteEvents); jsonErr == nil {
			var parsed []Event
			for _, r := range remoteEvents {
				parsed = append(parsed, Event{
					Month:       month,
					Day:         r.JalaliDay,
					Title:       r.Title,
					IsHoliday:   r.IsHoliday,
					Description: r.DateString,
				})
			}

			if len(parsed) > 0 {
				eventsCacheMu.Lock()
				eventsCache[key] = parsed
				eventsCacheMu.Unlock()
				return parsed
			}
		}
	}

	// Fallback to offline events database (do not cache failures permanently so online fetch can retry)
	return GetOfflineEvents(year, month)
}

// GetOfflineEvents extracts built-in solar and lunar events for a given year and month.
func GetOfflineEvents(year, month int) []Event {
	var list []Event
	for _, ev := range SolarEvents {
		if ev.Month == month {
			list = append(list, ev)
		}
	}
	if lunarList, exists := YearlyLunarEvents[year]; exists {
		for _, ev := range lunarList {
			if ev.Month == month {
				list = append(list, ev)
			}
		}
	}
	return list
}

// GetOfflineEventsForMonth extracts built-in solar events for a month.
func GetOfflineEventsForMonth(month int) []Event {
	return GetOfflineEvents(0, month)
}

// SolarEvents lists fixed Iranian solar holidays and cultural celebrations.
var SolarEvents = []Event{
	// Farvardin
	{Month: 1, Day: 1, Title: "جشن نوروز / آغاز سال نو", IsHoliday: true, Description: "آغاز سال نوی خورشیدی و جشن باستانی نوروز"},
	{Month: 1, Day: 2, Title: "عید نوروز", IsHoliday: true, Description: "تعطیلات رسمی نوروز"},
	{Month: 1, Day: 3, Title: "عید نوروز", IsHoliday: true, Description: "تعطیلات رسمی نوروز"},
	{Month: 1, Day: 4, Title: "عید نوروز", IsHoliday: true, Description: "تعطیلات رسمی نوروز"},
	{Month: 1, Day: 12, Title: "روز جمهوری اسلامی", IsHoliday: true, Description: "روز جمهوری اسلامی ایران"},
	{Month: 1, Day: 13, Title: "روز طبیعت (سیزده‌بدر)", IsHoliday: true, Description: "جشن باستانی سیزده‌بدر و گرامیداشت طبیعت"},
	{Month: 1, Day: 25, Title: "روز بزرگداشت عطار نیشابوری", IsHoliday: false, Description: "بزرگداشت شاعر و عارف نامدار ایرانی"},
	{Month: 1, Day: 29, Title: "روز ارتش جمهوری اسلامی", IsHoliday: false, Description: "روز ارتش و نیروی زمینی"},

	// Ordibehesht
	{Month: 2, Day: 1, Title: "روز بزرگداشت سعدی", IsHoliday: false, Description: "بزرگداشت استاد سخن سعدی شیرازی"},
	{Month: 2, Day: 10, Title: "روز ملی خلیج فارس", IsHoliday: false, Description: "سالروز اخراج پرتغالی‌ها از تنگه هرمز"},
	{Month: 2, Day: 12, Title: "روز معلم", IsHoliday: false, Description: "گرامیداشت مقام معلم و استاد"},
	{Month: 2, Day: 25, Title: "روز بزرگداشت فردوسی و زبان فارسی", IsHoliday: false, Description: "پاسداشت حکیم ابوالقاسم فردوسی"},
	{Month: 2, Day: 28, Title: "روز بزرگداشت حکیم عمر خیام", IsHoliday: false, Description: "بزرگداشت دانشمند، ریاضیدان و شاعر نامدار"},

	// Khordad
	{Month: 3, Day: 3, Title: "سالروز آزادسازی خرمشهر", IsHoliday: false, Description: "روز مقاومت، ایثار و پیروزی"},
	{Month: 3, Day: 14, Title: "رحلت امام خمینی", IsHoliday: true, Description: "تعطیل رسمی"},
	{Month: 3, Day: 15, Title: "قیام ۱۵ خرداد", IsHoliday: true, Description: "تعطیل رسمی"},

	// Tir
	{Month: 4, Day: 10, Title: "جشن نیلوفر / تیرگان", IsHoliday: false, Description: "جشن باستانی تیرگان و آرش کمانگیر"},
	{Month: 4, Day: 14, Title: "روز قلم", IsHoliday: false, Description: "روز گرامیداشت نویسندگان و اهل قلم"},

	// Mordad
	{Month: 5, Day: 7, Title: "جشن مردادگان", IsHoliday: false, Description: "جشن باستانی امردادگان"},
	{Month: 5, Day: 14, Title: "صدور فرمان مشروطیت", IsHoliday: false, Description: "سالروز امضای فرمان مشروطه"},
	{Month: 5, Day: 28, Title: "کودتای ۲۸ مرداد", IsHoliday: false, Description: "سالروز کودتای ۲۸ مرداد علیه دولت دکتر مصدق"},

	// Shahrivar
	{Month: 6, Day: 1, Title: "روز بزرگداشت ابوعلی سینا و روز پزشک", IsHoliday: false, Description: "بزرگداشت پورسینا و قدردانی از پزشکان"},
	{Month: 6, Day: 5, Title: "روز بزرگداشت زکریای رازی و روز داروساز", IsHoliday: false, Description: "بزرگداشت دانشمند بزرگ ایرانی"},
	{Month: 6, Day: 27, Title: "روز شعر و ادب فارسی / بزرگداشت استاد شهریار", IsHoliday: false, Description: "پاسداشت شعر و ادب ایرانی"},

	// Mehr
	{Month: 7, Day: 8, Title: "روز بزرگداشت مولوی (مولانا)", IsHoliday: false, Description: "بزرگداشت جلال‌الدین محمد بلخی"},
	{Month: 7, Day: 16, Title: "جشن مهرگان", IsHoliday: false, Description: "جشن کهن مهرگان"},
	{Month: 7, Day: 20, Title: "روز بزرگداشت حافظ شیرازی", IsHoliday: false, Description: "بزرگداشت لسان‌الغیب خواجه حافظ شیرازی"},

	// Aban
	{Month: 8, Day: 1, Title: "روز آمار و برنامه‌ریزی", IsHoliday: false, Description: "گرامیداشت علم آمار"},
	{Month: 8, Day: 7, Title: "سالروز ورود کوروش بزرگ به بابل", IsHoliday: false, Description: "پاسداشت پادشاه دادگستر هخامنشی در سال ۵۳۹ پیش از میلاد"},
	{Month: 8, Day: 8, Title: "روز ملی محیط‌بان", IsHoliday: false, Description: "گرامیداشت محیط‌بانان کشور"},
	{Month: 8, Day: 10, Title: "جشن آبانگان", IsHoliday: false, Description: "جشن کهن پاسداشت آب"},
	{Month: 8, Day: 24, Title: "روز کتاب و کتاب‌خوانی", IsHoliday: false, Description: "بزرگداشت علامه طباطبایی و ترویج مطالعه"},

	// Azar
	{Month: 9, Day: 16, Title: "روز دانشجو", IsHoliday: false, Description: "پاسداشت شهدای ۱۶ آذر ۱۳۳۲ دانشگاه تهران"},
	{Month: 9, Day: 30, Title: "جشن شب یلدا (چله)", IsHoliday: false, Description: "طولانی‌ترین شب سال و جشن کهن یلدا"},

	// Dey
	{Month: 10, Day: 1, Title: "جشن خرم‌روز / دی‌گان", IsHoliday: false, Description: "آغاز زمستان و جشن دیگان"},

	// Bahman
	{Month: 11, Day: 22, Title: "پیروزی انقلاب اسلامی", IsHoliday: true, Description: "تعطیل رسمی"},
	{Month: 11, Day: 29, Title: "جشن سپندارمذگان (روز عشق ایرانی)", IsHoliday: false, Description: "روز پاسداشت زن و زمین در ایران باستان"},

	// Esfand
	{Month: 12, Day: 5, Title: "روز بزرگداشت خواجه نصیرالدین طوسی و روز مهندس", IsHoliday: false, Description: "بزرگداشت دانشمند و ریاضیدان نامدار"},
	{Month: 12, Day: 29, Title: "روز ملی شدن صنعت نفت ایران", IsHoliday: true, Description: "تعطیل رسمی"},
}

// YearlyLunarEvents provides offline lunar holidays for recent and upcoming Jalali years.
var YearlyLunarEvents = map[int][]Event{
	1402: {
		{Month: 1, Day: 23, Title: "شهادت حضرت علی (ع)", IsHoliday: true, Description: "۲۱ رمضان"},
		{Month: 2, Day: 2, Title: "عید سعید فطر", IsHoliday: true, Description: "۱ شوال"},
		{Month: 2, Day: 3, Title: "تعطیل به مناسبت عید سعید فطر", IsHoliday: true, Description: "۲ شوال"},
		{Month: 2, Day: 26, Title: "شهادت امام جعفر صادق (ع)", IsHoliday: true, Description: "۲۵ شوال"},
		{Month: 4, Day: 8, Title: "عید سعید قربان", IsHoliday: true, Description: "۱۰ ذی‌الحجه"},
		{Month: 4, Day: 16, Title: "عید سعید غدیر خم", IsHoliday: true, Description: "۱۸ ذی‌الحجه"},
		{Month: 5, Day: 5, Title: "تاسوعای حسینی", IsHoliday: true, Description: "۹ محرم"},
		{Month: 5, Day: 6, Title: "عاشورای حسینی", IsHoliday: true, Description: "۱۰ محرم"},
		{Month: 6, Day: 15, Title: "اربعین حسینی", IsHoliday: true, Description: "۲۰ صفر"},
		{Month: 6, Day: 23, Title: "رحلت رسول اکرم (ص) و شهادت امام حسن مجتبی (ع)", IsHoliday: true, Description: "۲۸ صفر"},
		{Month: 6, Day: 25, Title: "شهادت امام رضا (ع)", IsHoliday: true, Description: "۳۰ صفر"},
		{Month: 7, Day: 2, Title: "شهادت امام حسن عسکری (ع)", IsHoliday: true, Description: "۸ ربیع‌الاول"},
		{Month: 7, Day: 11, Title: "میلاد رسول اکرم (ص) و امام جعفر صادق (ع)", IsHoliday: true, Description: "۱۷ ربیع‌الاول"},
		{Month: 9, Day: 26, Title: "شهادت حضرت فاطمه زهرا (س)", IsHoliday: true, Description: "۳ جمادی‌الثانی"},
		{Month: 11, Day: 5, Title: "ولادت امام علی (ع)", IsHoliday: true, Description: "۱۳ رجب"},
		{Month: 11, Day: 19, Title: "مبعث حضرت رسول اکرم (ص)", IsHoliday: true, Description: "۲۷ رجب"},
		{Month: 12, Day: 6, Title: "ولادت حضرت قائم (عج)", IsHoliday: true, Description: "۱۵ شعبان"},
	},
	1403: {
		{Month: 1, Day: 22, Title: "عید سعید فطر", IsHoliday: true, Description: "۱ شوال"},
		{Month: 1, Day: 23, Title: "تعطیل به مناسبت عید سعید فطر", IsHoliday: true, Description: "۲ شوال"},
		{Month: 2, Day: 15, Title: "شهادت امام جعفر صادق (ع)", IsHoliday: true, Description: "۲۵ شوال"},
		{Month: 3, Day: 28, Title: "عید سعید قربان", IsHoliday: true, Description: "۱۰ ذی‌الحجه"},
		{Month: 4, Day: 5, Title: "عید سعید غدیر خم", IsHoliday: true, Description: "۱۸ ذی‌الحجه"},
		{Month: 4, Day: 25, Title: "تاسوعای حسینی", IsHoliday: true, Description: "۹ محرم"},
		{Month: 4, Day: 26, Title: "عاشورای حسینی", IsHoliday: true, Description: "۱۰ محرم"},
		{Month: 6, Day: 4, Title: "اربعین حسینی", IsHoliday: true, Description: "۲۰ صفر"},
		{Month: 6, Day: 12, Title: "رحلت رسول اکرم (ص) و شهادت امام حسن مجتبی (ع)", IsHoliday: true, Description: "۲۸ صفر"},
		{Month: 6, Day: 14, Title: "شهادت امام رضا (ع)", IsHoliday: true, Description: "۳۰ صفر"},
		{Month: 6, Day: 22, Title: "شهادت امام حسن عسکری (ع)", IsHoliday: true, Description: "۸ ربیع‌الاول"},
		{Month: 6, Day: 31, Title: "میلاد رسول اکرم (ص) و امام جعفر صادق (ع)", IsHoliday: true, Description: "۱۷ ربیع‌الاول"},
		{Month: 9, Day: 15, Title: "شهادت حضرت فاطمه زهرا (س)", IsHoliday: true, Description: "۳ جمادی‌الثانی"},
		{Month: 10, Day: 25, Title: "ولادت امام علی (ع)", IsHoliday: true, Description: "۱۳ رجب"},
		{Month: 11, Day: 8, Title: "مبعث حضرت رسول اکرم (ص)", IsHoliday: true, Description: "۲۷ رجب"},
		{Month: 11, Day: 26, Title: "ولادت حضرت قائم (عج)", IsHoliday: true, Description: "۱۵ شعبان"},
	},
	1404: {
		{Month: 1, Day: 11, Title: "عید سعید فطر", IsHoliday: true, Description: "۱ شوال"},
		{Month: 1, Day: 12, Title: "تعطیل به مناسبت عید سعید فطر", IsHoliday: true, Description: "۲ شوال"},
		{Month: 2, Day: 4, Title: "شهادت امام جعفر صادق (ع)", IsHoliday: true, Description: "۲۵ شوال"},
		{Month: 3, Day: 17, Title: "عید سعید قربان", IsHoliday: true, Description: "۱۰ ذی‌الحجه"},
		{Month: 3, Day: 25, Title: "عید سعید غدیر خم", IsHoliday: true, Description: "۱۸ ذی‌الحجه"},
		{Month: 4, Day: 14, Title: "تاسوعای حسینی", IsHoliday: true, Description: "۹ محرم"},
		{Month: 4, Day: 15, Title: "عاشورای حسینی", IsHoliday: true, Description: "۱۰ محرم"},
		{Month: 5, Day: 24, Title: "اربعین حسینی", IsHoliday: true, Description: "۲۰ صفر"},
		{Month: 6, Day: 1, Title: "رحلت رسول اکرم (ص) و شهادت امام حسن مجتبی (ع)", IsHoliday: true, Description: "۲۸ صفر"},
		{Month: 6, Day: 3, Title: "شهادت امام رضا (ع)", IsHoliday: true, Description: "۳۰ صفر"},
		{Month: 6, Day: 11, Title: "شهادت امام حسن عسکری (ع)", IsHoliday: true, Description: "۸ ربیع‌الاول"},
		{Month: 6, Day: 20, Title: "میلاد رسول اکرم (ص) و امام جعفر صادق (ع)", IsHoliday: true, Description: "۱۷ ربیع‌الاول"},
		{Month: 9, Day: 4, Title: "شهادت حضرت فاطمه زهرا (س)", IsHoliday: true, Description: "۳ جمادی‌الثانی"},
		{Month: 10, Day: 14, Title: "ولادت امام علی (ع)", IsHoliday: true, Description: "۱۳ رجب"},
		{Month: 10, Day: 28, Title: "مبعث حضرت رسول اکرم (ص)", IsHoliday: true, Description: "۲۷ رجب"},
		{Month: 11, Day: 15, Title: "ولادت حضرت قائم (عج)", IsHoliday: true, Description: "۱۵ شعبان"},
		{Month: 12, Day: 20, Title: "شهادت حضرت علی (ع)", IsHoliday: true, Description: "۲۱ رمضان"},
	},
	1405: {
		{Month: 1, Day: 1, Title: "عید سعید فطر", IsHoliday: true, Description: "۱ شوال"},
		{Month: 1, Day: 2, Title: "تعطیل به مناسبت عید سعید فطر", IsHoliday: true, Description: "۲ شوال"},
		{Month: 1, Day: 23, Title: "شهادت امام جعفر صادق (ع)", IsHoliday: true, Description: "۲۵ شوال"},
		{Month: 3, Day: 6, Title: "عید سعید قربان", IsHoliday: true, Description: "۱۰ ذی‌الحجه"},
		{Month: 3, Day: 14, Title: "عید سعید غدیر خم", IsHoliday: true, Description: "۱۸ ذی‌الحجه"},
		{Month: 4, Day: 4, Title: "تاسوعای حسینی", IsHoliday: true, Description: "۹ محرم"},
		{Month: 4, Day: 5, Title: "عاشورای حسینی", IsHoliday: true, Description: "۱۰ محرم"},
		{Month: 5, Day: 14, Title: "اربعین حسینی", IsHoliday: true, Description: "۲۰ صفر"},
		{Month: 5, Day: 22, Title: "رحلت رسول اکرم (ص) و شهادت امام حسن مجتبی (ع)", IsHoliday: true, Description: "۲۸ صفر"},
		{Month: 5, Day: 24, Title: "شهادت امام رضا (ع)", IsHoliday: true, Description: "۳۰ صفر"},
		{Month: 6, Day: 1, Title: "شهادت امام حسن عسکری (ع)", IsHoliday: true, Description: "۸ ربیع‌الاول"},
		{Month: 6, Day: 10, Title: "میلاد رسول اکرم (ص) و امام جعفر صادق (ع)", IsHoliday: true, Description: "۱۷ ربیع‌الاول"},
		{Month: 8, Day: 24, Title: "شهادت حضرت فاطمه زهرا (س)", IsHoliday: true, Description: "۳ جمادی‌الثانی"},
		{Month: 10, Day: 3, Title: "ولادت امام علی (ع)", IsHoliday: true, Description: "۱۳ رجب"},
		{Month: 10, Day: 17, Title: "مبعث حضرت رسول اکرم (ص)", IsHoliday: true, Description: "۲۷ رجب"},
		{Month: 11, Day: 4, Title: "ولادت حضرت قائم (عج)", IsHoliday: true, Description: "۱۵ شعبان"},
		{Month: 12, Day: 10, Title: "شهادت حضرت علی (ع)", IsHoliday: true, Description: "۲۱ رمضان"},
		{Month: 12, Day: 30, Title: "عید سعید فطر", IsHoliday: true, Description: "۱ شوال"},
	},
	1406: {
		{Month: 1, Day: 1, Title: "عید سعید فطر", IsHoliday: true, Description: "۱ شوال"},
		{Month: 1, Day: 12, Title: "شهادت امام جعفر صادق (ع)", IsHoliday: true, Description: "۲۵ شوال"},
		{Month: 2, Day: 26, Title: "عید سعید قربان", IsHoliday: true, Description: "۱۰ ذی‌الحجه"},
		{Month: 3, Day: 3, Title: "عید سعید غدیر خم", IsHoliday: true, Description: "۱۸ ذی‌الحجه"},
		{Month: 3, Day: 23, Title: "تاسوعای حسینی", IsHoliday: true, Description: "۹ محرم"},
		{Month: 3, Day: 24, Title: "عاشورای حسینی", IsHoliday: true, Description: "۱۰ محرم"},
		{Month: 5, Day: 3, Title: "اربعین حسینی", IsHoliday: true, Description: "۲۰ صفر"},
		{Month: 5, Day: 11, Title: "رحلت رسول اکرم (ص) و شهادت امام حسن مجتبی (ع)", IsHoliday: true, Description: "۲۸ صفر"},
		{Month: 5, Day: 13, Title: "شهادت امام رضا (ع)", IsHoliday: true, Description: "۳۰ صفر"},
		{Month: 5, Day: 21, Title: "شهادت امام حسن عسکری (ع)", IsHoliday: true, Description: "۸ ربیع‌الاول"},
		{Month: 5, Day: 30, Title: "میلاد رسول اکرم (ص) و امام جعفر صادق (ع)", IsHoliday: true, Description: "۱۷ ربیع‌الاول"},
		{Month: 8, Day: 13, Title: "شهادت حضرت فاطمه زهرا (س)", IsHoliday: true, Description: "۳ جمادی‌الثانی"},
		{Month: 9, Day: 23, Title: "ولادت امام علی (ع)", IsHoliday: true, Description: "۱۳ رجب"},
		{Month: 10, Day: 7, Title: "مبعث حضرت رسول اکرم (ص)", IsHoliday: true, Description: "۲۷ رجب"},
		{Month: 10, Day: 24, Title: "ولادت حضرت قائم (عج)", IsHoliday: true, Description: "۱۵ شعبان"},
		{Month: 11, Day: 28, Title: "شهادت حضرت علی (ع)", IsHoliday: true, Description: "۲۱ رمضان"},
		{Month: 12, Day: 19, Title: "عید سعید فطر", IsHoliday: true, Description: "۱ شوال"},
		{Month: 12, Day: 20, Title: "تعطیل به مناسبت عید سعید فطر", IsHoliday: true, Description: "۲ شوال"},
	},
}

// GetEventsForDate returns all events matching the given Jalali month and day.
func GetEventsForDate(year, month, day int) []Event {
	monthEvents := FetchMonthEvents(year, month)
	var results []Event
	for _, ev := range monthEvents {
		if ev.Day == day {
			results = append(results, ev)
		}
	}
	return results
}

// IsHoliday returns true if the date is Friday or matches an official holiday.
func IsHoliday(date JalaliDate) (bool, string) {
	if date.Weekday == 6 { // Friday
		return true, "جمعه (تعطیل آخر هفته)"
	}
	events := GetEventsForDate(date.Year, date.Month, date.Day)
	for _, ev := range events {
		if ev.IsHoliday {
			return true, ev.Title
		}
	}
	return false, ""
}
