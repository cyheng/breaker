package common

import (
	"errors"
	"time"
)

type Retry struct {
	NextDelay func(int) int
}

func Timed(total int, delay int) *Retry {
	return &Retry{
		NextDelay: func(attempt int) int {
			if attempt >= total {
				return -1
			}
			return delay
		}}
}
func (r *Retry) On(method func() error) error {
	attempt := 0
	for {
		err := method()
		if err == nil {
			return nil
		}
		delay := r.NextDelay(attempt)
		if delay < 0 {
			return errors.New("All retry attempts failed.")
		}
		<-time.After(time.Duration(delay) * time.Millisecond)
		attempt++
	}
}
