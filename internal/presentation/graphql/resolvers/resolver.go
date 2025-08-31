package resolvers

import (
	"github.com/terminator791/Layered-Architecture-graphql-GO/internal/business/service"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct{
	MessageService service.MessageService
	UserService    service.UserService
	RoomService    service.RoomService
}
