package zenmodel

import "github.com/zenmodel/zenmodel/internal/constants"

type Maintainer interface {
	Start()
	ShutDown()
	SendMessage(constants.Message)
}
