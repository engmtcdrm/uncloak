package task

type Status int

const (
	NotStarted Status = iota
	Started
	Finished
	Warning
	Error
)
