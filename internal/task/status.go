package task

type Status int

const (
	Started Status = iota
	Finished
	Warning
	Error
)
