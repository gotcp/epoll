package epoll

/*
#cgo LDFLAGS: -lssl -lcrypto -ldl
#include <malloc.h>
#include <stdint.h>
#include <stdio.h>
#include <string.h>
#include <openssl/evp.h>
#include <openssl/rand.h>
#include <openssl/ssl.h>
#include <openssl/dh.h>
#include <openssl/err.h>
#include <openssl/crypto.h>
*/
import "C"
import (
	"errors"
	"time"
	"unsafe"

	"github.com/wuyongjia/hashmap"
	"github.com/wuyongjia/pool"
	"github.com/wuyongjia/threadpool"
)

const (
	DEFAULT_C_MALLOC_TRIM_INTERVAL = 5000
)

type SSL struct {
	Id  uint64
	SSL *C.SSL
}

type Conn struct {
	Id         uint64
	Fd         int
	SSL        *SSL
	Data       interface{}
	SequenceId int
	Timestamp  int64
	Status     int
}

type EP struct {
	Host               string
	Port               int
	Epfd               int
	Fd                 int
	Connections        *hashmap.HM
	SSLCtx             *C.SSL_CTX
	IsSSL              bool
	Threads            int
	QueueLength        int
	ReadBuffer         int
	WriteBuffer        int
	EpollEvents        int
	WaitTimeout        int
	ReadTimeout        int
	WriteTimeout       int
	KeepAlive          int
	ReuseAddr          int
	ReusePort          int
	bufferPool         *pool.Pool               // []byte pool, return *[]byte
	connPool           *pool.Pool               // Conn pool, return *Conn
	requestPool        *pool.Pool               // *Request pool, return *Request
	sslPool            *pool.Pool               // *C.SSL pool, return *C.SSL
	threadPoolSequence *threadpool.PoolSequence // thread pool sequence
	OnAccept           OnAcceptEvent
	OnReceive          OnReceiveEvent
	OnClose            OnCloseEvent
	OnError            OnErrorEvent
}

func (ep *EP) newSSLPool(length int) *pool.Pool {
	return pool.NewWithId(length, func(id uint64) interface{} {
		var ssl = &SSL{
			Id:  id,
			SSL: C.SSL_new(ep.SSLCtx),
		}
		return ssl
	})
}

func (ep *EP) getSSL() *SSL {
	var ssl, err = ep.sslPool.Get()
	if err == nil {
		return ssl.(*SSL)
	}
	return nil
}

func (ep *EP) putSSL(ssl *SSL) {
	C.SSL_clear(ssl.SSL)
	ep.sslPool.PutWithId(ssl, ssl.Id)
}

func newSSLCtx(certFile string, keyFile string) *C.SSL_CTX {
	var cret C.int = C.OPENSSL_init_ssl(C.OPENSSL_INIT_LOAD_SSL_STRINGS|C.OPENSSL_INIT_LOAD_CRYPTO_STRINGS, nil)
	if cret <= 0 {
		panic(errors.New("unable to init SSL"))
	}

	var method = C.SSLv23_server_method()
	var ctx = C.SSL_CTX_new(method)
	if ctx == nil {
		panic(errors.New("unable to create SSL context"))
	}

	var ccertp = C.CString(certFile)
	var ckeyp = C.CString(keyFile)

	defer C.free(unsafe.Pointer(ccertp))
	defer C.free(unsafe.Pointer(ckeyp))

	if C.SSL_CTX_use_certificate_file(ctx, ccertp, C.SSL_FILETYPE_PEM) <= 0 {
		panic(errors.New("unable to set certificate"))
	}

	if C.SSL_CTX_use_PrivateKey_file(ctx, ckeyp, C.SSL_FILETYPE_PEM) <= 0 {
		panic(errors.New("unable to set private key"))
	}

	return ctx
}

func (ep *EP) freeSSLCtx() {
	if ep.SSLCtx != nil {
		C.SSL_CTX_free(ep.SSLCtx)
		ep.SSLCtx = nil
	}
}

func (ep *EP) GetConnectionSSL(fd int) *SSL {
	var ssl *SSL
	ep.Connections.UpdateWithFunc(fd, func(value interface{}) {
		var c, ok = value.(*Conn)
		if ok {
			c.Timestamp = time.Now().Unix()
			ssl = c.SSL
		}
	})
	return ssl
}

func (ep *EP) GetConnectionSequenceIdAndSSL(fd int) (int, *SSL) {
	var ssl *SSL
	var sequenceId = -1
	ep.Connections.UpdateWithFunc(fd, func(value interface{}) {
		var c, ok = value.(*Conn)
		if ok {
			c.Timestamp = time.Now().Unix()
			sequenceId = c.SequenceId
			ssl = c.SSL
		}
	})
	return sequenceId, ssl
}

func (ep *EP) newSSL(fd int) *SSL {
	var ssl = ep.getSSL()
	if ssl == nil {
		return nil
	}
	if C.SSL_set_fd(ssl.SSL, (C.int)(fd)) <= 0 {
		ep.putSSL(ssl)
		return nil
	}
	if C.SSL_accept(ssl.SSL) <= 0 {
		ep.putSSL(ssl)
		return nil
	}
	return ssl
}

func (ep *EP) putConnSSL(conn *Conn) {
	if conn.SSL != nil {
		ep.putSSL(conn.SSL)
		conn.SSL = nil
	}
}

func (ep *EP) AddConnectionSSL(fd int, ssl *SSL, sequenceId int) {
	var conn = ep.getConn()
	conn.Fd = fd
	conn.SSL = ssl
	conn.SequenceId = sequenceId
	conn.Timestamp = time.Now().Unix()
	conn.Status = 0
	ep.Connections.Put(fd, conn)
}

func (ep *EP) setConnectionSSL(fd int, ssl *SSL) bool {
	if ssl == nil {
		return false
	}
	var ok bool
	ep.Connections.UpdateWithFunc(fd, func(value interface{}) {
		var c, ok = value.(*Conn)
		if ok {
			c.SSL = ssl
		}
	})
	return ok
}

func sslRead(ssl *C.SSL, buffer []byte, n int) int {
	return int(C.SSL_read(ssl, unsafe.Pointer(&buffer[0]), (C.int)(n)))
}

func sslWrite(ssl *C.SSL, buffer []byte, n int) bool {
	if C.SSL_write(ssl, unsafe.Pointer(&buffer[0]), (C.int)(n)) > 0 {
		return true
	}
	return false
}

func freeSSL(ssl *SSL) {
	if ssl.SSL != nil {
		C.SSL_free(ssl.SSL)
		ssl.SSL = nil
	}
}

func sslRecycleUpdate(ptr interface{}) {
	var ssl, ok = ptr.(*SSL)
	if ok && ssl != nil {
		freeSSL(ssl)
	}
}

func cMallocTrimLoop() {
	go func() {
		var timer = time.NewTicker(DEFAULT_C_MALLOC_TRIM_INTERVAL * time.Second)
		defer timer.Stop()
		for {
			<-timer.C
			cMallocTrim()
		}
	}()
}

func cMallocTrim() {
	C.malloc_trim(0)
}
