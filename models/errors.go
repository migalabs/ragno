package models

type connectionError string

const (
	TimeoutError connectionError = "time-out"
)

func (c connectionError) String() string {
	return string(c)
}
