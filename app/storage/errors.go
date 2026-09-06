package storage

import "fmt"

var ErrMessageNotDecoded = fmt.Errorf("failed to decode message")
var ErrMessageNotEncoded = fmt.Errorf("failed to encode message")
var ErrMessageNotFound = fmt.Errorf("message not found")
var ErrMessageNotCreated = fmt.Errorf("failed to create message")
