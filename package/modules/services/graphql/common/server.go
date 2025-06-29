package common

type GraphqlServer interface {
	Start() error
	Stop() error
}
