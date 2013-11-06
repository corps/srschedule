package srschedule

import (
	"time"
)

type Schedule struct {
	LastAnswered time.Time 
	Due          time.Time
}

type IntervalSchedule struct {
	Schedule
	Interval time.Duration
}
