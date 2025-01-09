package core

type Executor interface {
	Exec(command string) (string, error)
	Test() bool
	Verify() error
}
