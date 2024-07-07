package sarfya

import "errors"

var ErrDictionaryEntryNotFound = errors.New("dictionary entry not found")
var ErrExampleNotFound = errors.New("example not found")
var ErrReadOnly = errors.New("modifications are not allowed")
