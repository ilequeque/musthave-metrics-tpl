package repository

import "time"

func retry(fn func() error) error {
	delays := []time.Duration{
		1 * time.Second,
		3 * time.Second,
		5 * time.Second,
	}

	var err error
	for i := 0; i <= len(delays); i++ {
		if err = fn(); err == nil {
			return nil
		}
		if i < len(delays) {
			time.Sleep(delays[i])
		}
	}
	return err
}
