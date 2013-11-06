package srschedule

import (
	"github.com/stretchr/testify/assert"
	"time"
	"testing"
)

type ConstVariance struct {
	Result float64
}

func (c ConstVariance) Float64() float64 {
	return c.Result
}

type testCase struct {
	scheduler *IntervalFactorScheduler
	now time.Time
	factor float64
	answered time.Time
	due time.Time
	interval time.Duration
}

func (t *testCase) currentSchedule() IntervalSchedule {
	return IntervalSchedule{Interval: t.interval, Schedule: Schedule{Due: t.due, LastAnswered: t.due.Add(-t.interval)}}
}

func (t *testCase) nextSchedule() IntervalSchedule {
	return t.scheduler.NextByFactor(t.currentSchedule(), t.answered, t.factor)
}

func (t *testCase) nextInterval() time.Duration {
	return t.nextSchedule().Interval
}

func newTestCase() *testCase {
	now := time.Now()
	return &testCase{scheduler: &IntervalFactorScheduler{MinimumInterval: time.Hour, VarianceFactor: 0.4},
		now: now, factor: 2.0, answered: now, due: now, interval: time.Hour}
}

func TestFactorVariance(t *testing.T) {
	testCase := newTestCase()
	testCase.scheduler.Variance = ConstVariance{Result: 1.0}
	high := testCase.nextInterval()

	testCase.scheduler.Variance = ConstVariance{Result: 0.0}
	low := testCase.nextInterval()

	assert.Equal(t, float64(testCase.scheduler.MinimumInterval.Nanoseconds()) * 
		float64(testCase.scheduler.VarianceFactor), float64(high.Nanoseconds()) - float64(low.Nanoseconds()))
}

func TestNewIntervalFactorScheduler(t *testing.T) {
	scheduler := NewIntervalFactorScheduler(time.Minute * 3, 0.3)
	assert.Equal(t, time.Minute * 3, scheduler.MinimumInterval)
	assert.Equal(t, 0.3, scheduler.VarianceFactor)
	assert.NotNil(t, scheduler.Variance)
}

func TestNextByFactorScheduleAttributes(t *testing.T) {
	testCase := newTestCase()
	testCase.scheduler.Variance = &ConstVariance{Result: 0.5}

	testCase.answered = testCase.now.Add(time.Hour * 35)
	testCase.interval = testCase.scheduler.MinimumInterval * 2
	testCase.due = testCase.now.Add(-time.Hour * 3)

	assert.Equal(t, testCase.answered.Add(testCase.nextInterval()), testCase.nextSchedule().Due)
	assert.Equal(t, testCase.answered, testCase.nextSchedule().LastAnswered)
}

func TestNextByFactorIntervals(t *testing.T) {
	testCase := newTestCase()
	testCase.scheduler.Variance = &ConstVariance{Result: 0.5}

	// with a factor < 1
	testCase.factor = 0.6
	{
		// when the interval is the minimum
		testCase.interval = testCase.scheduler.MinimumInterval
		{
			// answering right on the due date
			testCase.answered = testCase.due
			assert.Equal(t, testCase.scheduler.MinimumInterval, testCase.nextInterval())

			// answering after only 1/3 of the time between its last answered and due date
			testCase.answered = testCase.currentSchedule().LastAnswered.Add(testCase.interval / 3)
			assert.Equal(t, testCase.scheduler.MinimumInterval, testCase.nextInterval())

			// answering long after the due date
			testCase.answered = testCase.due.Add(time.Hour * 300)
			assert.Equal(t, testCase.scheduler.MinimumInterval, testCase.nextInterval())
		}	

		// when the interval is the minimum times two
		testCase.interval = testCase.scheduler.MinimumInterval * 2
		{
			// answering right on the due date
			testCase.answered = testCase.due
			assert.Equal(t, testCase.scheduler.MinimumInterval + testCase.scheduler.MinimumInterval / 5, testCase.nextInterval())

			// answering after only 1/3 of the time between its last answered and due date
			testCase.answered = testCase.currentSchedule().LastAnswered.Add(testCase.interval / 3)
			assert.Equal(t, testCase.scheduler.MinimumInterval + testCase.scheduler.MinimumInterval / 5, testCase.nextInterval())

			// answering long after the due date
			testCase.answered = testCase.due.Add(time.Hour * 300)
			assert.Equal(t, testCase.scheduler.MinimumInterval + testCase.scheduler.MinimumInterval / 5, testCase.nextInterval())

		}
	}

	// with a factor > 1
	testCase.factor = 2.0
	{
		// when the interval is the minimum
		testCase.interval = testCase.scheduler.MinimumInterval
		{
			// answering right on the due date
			testCase.answered = testCase.due
			assert.Equal(t, testCase.scheduler.MinimumInterval * 2, testCase.nextInterval())

			// answering after only 1/3 of the time between its last answered and due date
			testCase.answered = testCase.currentSchedule().LastAnswered.Add(testCase.interval / 3)
			assert.Equal(t, testCase.scheduler.MinimumInterval + testCase.scheduler.MinimumInterval / 3, testCase.nextInterval())

			// answering long after the due date
			testCase.answered = testCase.due.Add(time.Hour * 300)
			assert.Equal(t, testCase.scheduler.MinimumInterval * 2, testCase.nextInterval())
		}

		// when the interval is the minimum times two
		testCase.interval = testCase.scheduler.MinimumInterval * 2
		{
			// answering right on the due date
			testCase.answered = testCase.due
			assert.Equal(t, testCase.scheduler.MinimumInterval * 4, testCase.nextInterval())

			// answering after only 1/3 of the time between its last answered and due date
			testCase.answered = testCase.currentSchedule().LastAnswered.Add(testCase.interval / 3)
			assert.Equal(t, testCase.scheduler.MinimumInterval + testCase.scheduler.MinimumInterval  * 5 / 3, testCase.nextInterval())

			// answering long after the due date
			testCase.answered = testCase.due.Add(time.Hour * 300)
			assert.Equal(t, testCase.scheduler.MinimumInterval * 4, testCase.nextInterval())
		}
	}
}