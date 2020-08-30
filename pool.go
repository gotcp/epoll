package epoll

import (
	"errors"
)

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
