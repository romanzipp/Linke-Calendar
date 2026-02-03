package calendar

import (
	"time"

	"github.com/romanzipp/linke-calendar/internal/database"
)

type Month struct {
	Year      int
	Month     time.Month
	Weeks     []Week
	MonthName string
}

type Week struct {
	Days []Day
}

type Day struct {
	Date      time.Time
	Day       int
	IsToday   bool
	InMonth   bool
	Events    []*database.Event
}

func Generate(year int, month time.Month, events []*database.Event) *Month {
	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstDay.AddDate(0, 1, -1)

	cal := &Month{
		Year:      year,
		Month:     month,
		MonthName: getGermanMonthName(month),
		Weeks:     make([]Week, 0),
	}

	eventsByDate := groupEventsByDate(events)

	currentWeek := Week{Days: make([]Day, 0)}

	startWeekday := int(firstDay.Weekday())
	if startWeekday == 0 {
		startWeekday = 7
	}

	for i := 1; i < startWeekday; i++ {
		prevDate := firstDay.AddDate(0, 0, -(startWeekday - i))
		currentWeek.Days = append(currentWeek.Days, Day{
			Date:    prevDate,
			Day:     prevDate.Day(),
			IsToday: false,
			InMonth: false,
			Events:  nil,
		})
	}

	today := time.Now().UTC()
	todayStr := today.Format("2006-01-02")

	for day := 1; day <= lastDay.Day(); day++ {
		date := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
		dateStr := date.Format("2006-01-02")
		isToday := dateStr == todayStr

		currentWeek.Days = append(currentWeek.Days, Day{
			Date:    date,
			Day:     day,
			IsToday: isToday,
			InMonth: true,
			Events:  eventsByDate[dateStr],
		})

		if len(currentWeek.Days) == 7 {
			cal.Weeks = append(cal.Weeks, currentWeek)
			currentWeek = Week{Days: make([]Day, 0)}
		}
	}

	if len(currentWeek.Days) > 0 {
		nextMonthDay := 1
		for len(currentWeek.Days) < 7 {
			nextDate := lastDay.AddDate(0, 0, nextMonthDay)
			currentWeek.Days = append(currentWeek.Days, Day{
				Date:    nextDate,
				Day:     nextDate.Day(),
				IsToday: false,
				InMonth: false,
				Events:  nil,
			})
			nextMonthDay++
		}
		cal.Weeks = append(cal.Weeks, currentWeek)
	}

	return cal
}

func getGermanMonthName(month time.Month) string {
	germanMonths := map[time.Month]string{
		time.January:   "Januar",
		time.February:  "Februar",
		time.March:     "März",
		time.April:     "April",
		time.May:       "Mai",
		time.June:      "Juni",
		time.July:      "Juli",
		time.August:    "August",
		time.September: "September",
		time.October:   "Oktober",
		time.November:  "November",
		time.December:  "Dezember",
	}
	return germanMonths[month]
}

func groupEventsByDate(events []*database.Event) map[string][]*database.Event {
	result := make(map[string][]*database.Event)
	for _, event := range events {
		dateStr := event.DatetimeStart.Format("2006-01-02")
		result[dateStr] = append(result[dateStr], event)
	}
	return result
}
