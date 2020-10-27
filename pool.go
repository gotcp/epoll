package epoll

func bufferRecycleUpdate(ptr interface{}) {
	var buffer, ok = ptr.(*[]byte)
	if ok && buffer != nil {
		*buffer = (*buffer)[:0]
	}
}

func (ep *EP) GetBuffer() (*[]byte, error) {
	var iface, err = ep.bufferPool.Get()
	if err == nil {
		var buffer, ok = iface.(*[]byte)
		if ok {
			return buffer, nil
		} else {
			return nil, ErrorGetPoolBuffer
		}
	} else {
		return nil, err
	}
}

func (ep *EP) PutBuffer(buffer *[]byte) {
	ep.bufferPool.Put(buffer)
}
