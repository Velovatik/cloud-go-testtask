package entity

import "time"

type Song struct {
	ID       int
	Title    string
	Duration time.Duration
	Artist   string
}
