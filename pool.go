package epoll

import (
	"errors"
)

func bufferRecycleUpdate(ptr interface{}) {
	var buffer, ok = ptr.(*[]byte)
	if ok && buffer != nil {
		*buffer = (*buffer)[:0]
	}
}

func (ep *EP) GetBufferPoolItem() (*[]byte, error) {
	var iface, err = ep.bufferPool.Get()
	if err == nil {
		var buffer, ok = iface.(*[]byte)
		if ok {
			return buffer, nil
		} else {
			return nil, errors.New("get pool buffer error")
		}
	} else {
		return nil, err
	}
}

func (ep *EP) PutBufferPoolItem(buffer *[]byte) {
	ep.bufferPool.Put(buffer)
}
