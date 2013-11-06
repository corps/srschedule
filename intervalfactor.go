package srschedule

import (
	"math"
	"math/rand"
	"time"
)

type IntervalFactorScheduler struct {
	MinimumInterval time.Duration
	Variance VarianceProvider
	VarianceFactor float64
}

type VarianceProvider interface {
	Float64() float64
}

func NewIntervalFactorScheduler(minimumInterval time.Duration, variance float64) *IntervalFactorScheduler {
	return &IntervalFactorScheduler{MinimumInterval: minimumInterval, VarianceFactor: variance, 
		Variance: rand.New(rand.NewSource(int64(time.Now().Nanosecond())))}
}

func (scheduler *IntervalFactorScheduler) NextByFactor(schedule IntervalSchedule, answered time.Time, factor float64) IntervalSchedule {
	answeredInterval := answered.Sub(schedule.LastAnswered)

	baseFactor := math.Min(factor, 1.0)
	bonusFactor := math.Max(0.0, factor - 1.0)
	randomFactor := scheduler.Variance.Float64() * scheduler.VarianceFactor + (1.0 - scheduler.VarianceFactor / 2.0)
	earlyAnswerMultiplier := float64(0.0)
	if schedule.Interval.Nanoseconds() != 0 {
		earlyAnswerMultiplier = math.Min(1.0, answeredInterval.Seconds() / schedule.Interval.Seconds())
	}

	effectiveFactor := baseFactor + (bonusFactor * earlyAnswerMultiplier * randomFactor)

	nextInterval := time.Duration(int64(float64(time.Nanosecond) * (float64(schedule.Interval.Nanoseconds()) * effectiveFactor)))
	if nextInterval.Nanoseconds() < scheduler.MinimumInterval.Nanoseconds() {
		nextInterval = scheduler.MinimumInterval
	}

	return IntervalSchedule{Schedule: Schedule{LastAnswered: answered, Due: answered.Add(nextInterval)}, Interval: nextInterval}
}
