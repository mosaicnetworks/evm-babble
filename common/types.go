package common

type AccountMap map[string]struct {
	Code    string
	Storage map[string]string
	Balance string
}
