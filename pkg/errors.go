package pkg

import (
	"errors"
	"fmt"
)

// ErrPkg generic package error.
var ErrPkg = errors.New("workerpool")

// ErrInit error.
var ErrInit = fmt.Errorf("%w: worker init error", ErrPkg)

// ErrClose error.
var ErrClose = fmt.Errorf("%w: worker close error", ErrPkg)

// ErrAction error.
var ErrAction = fmt.Errorf("%w: worker action error", ErrPkg)