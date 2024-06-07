package user

import "fmt"

var ErrBotsNotAllowed = fmt.Errorf("bots are restricted to use this service")
var ErrUserNotFound = fmt.Errorf("user not found")
var ErrUserNotCreated = fmt.Errorf("user not created")
