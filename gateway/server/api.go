package server

type IServer interface {
	Start() error
	Stop() error
}
