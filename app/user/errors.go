package user

import "fmt"

var ErrBotsNotAllowed = fmt.Errorf("bots are restricted to use this service")
var ErrUserNotFound = fmt.Errorf("user not found")
var ErrUserNotCreated = fmt.Errorf("user not created")
var ErrPromptNotFound = fmt.Errorf("prompt not found")
var ErrPromptNotCreated = fmt.Errorf("prompt not created")
var ErrModelNotSet = fmt.Errorf("model not set")
