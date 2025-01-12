package core

type Executor interface {
	Exec(command string) (string, error)
	Test() bool
	Verify() error
	SendFile(localPath, remotePath string) error
}
