package client

import "time"

const (
	secondsInDay          = 24 * 60 * 60
	tick                  = int64(time.Microsecond) / 10
	ticksPerSecond        = int64(time.Second) / tick
	ticksSinceEpoch       = (1969*365 + 1969/4 - 1969/100 + 1969/400) * secondsInDay * ticksPerSecond
	millisecondsPerSecond = int64(time.Second / time.Millisecond)
)

func timeFromTicks(ticks int64) time.Time {
	ticksSinceUnix := ticks - ticksSinceEpoch
	return time.Unix(ticksSinceUnix/ticksPerSecond,
		(ticksSinceUnix%ticksPerSecond)*tick)
}

func timeFromUnixMilliseconds(msSinceUnix int64) time.Time {
	return time.Unix(
		msSinceUnix/millisecondsPerSecond,
		(msSinceUnix%millisecondsPerSecond)*int64(time.Millisecond),
	)
}
