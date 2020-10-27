package epoll

import (
	"errors"
)

var (
	ErrorTemplateNotFound = "%d not found in the list"
)

var (
	ErrorGetPoolBuffer = errors.New("get pool buffer error")
)

var (
	ErrorSSLUnableCreate = errors.New("unable to create SSL connection")
	ErrorSSLUnknow       = errors.New("ssl error unknow")
	ErrorSSLWantWrite    = errors.New("ssl error want write")
	ErrorSSLWantRead     = errors.New("ssl error want read")
	ErrorSSLZeroReturn   = errors.New("ssl error zero return")
	ErrorSSLWantConnect  = errors.New("ssl error want connect")
	ErrorSSLTimeout      = errors.New("ssl error timeout")
	ErrorSSL             = errors.New("ssl error")
	ErrorSSLSyscall      = errors.New("ssl error syscall")
)
