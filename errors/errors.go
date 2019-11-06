package errors

import "errors"

var (
	DbClientsNotInit = errors.New("you have to initialize the global dbClients before using it")
)
