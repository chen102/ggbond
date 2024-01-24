package connmanage

// ConnManagerOption 连接管理器选项c
// 用于设置连接管理器的参数
type ConnManagerOption func(options *connManageroptions) error
type connManageroptions struct {
	maximumConnection   *int32 //最大连接数
	connectionTimedOut  *int64 //连接超时时间
	transmissionTimeout *int64 //传输超时时间
	explorationCycle    *int64 //探测周期
	detectionTimeout    *int64 //探测超时时间 每个连接探测超时时间，用次参数来监控连接是否正常
	readwriteTimeout    *int64 //读写超时时间
	readTimeout         *int64 //读超时时间
	writeTimeout        *int64 //写超时时间
	readbuffer          *int32 //读缓冲区大小
	writebuffer         *int32 //写缓冲区大小
}

// readbuffer:读缓冲区大小
func WithReadBuffer(readbuffer int32) ConnManagerOption {
	return func(options *connManageroptions) error {
		options.readbuffer = &readbuffer
		return nil
	}
}

// writebuffer:写缓冲区大小
func WithWriteBuffer(writebuffer int32) ConnManagerOption {
	return func(options *connManageroptions) error {
		options.writebuffer = &writebuffer
		return nil
	}
}

// maximumConnection:最大连接数
func WithMaximumConnection(maximumConnection int32) ConnManagerOption {
	return func(options *connManageroptions) error {
		options.maximumConnection = &maximumConnection
		return nil
	}
}

//

//读写超时时间
//若设置该参数WithReadTimeout、WithWriteTimeout将会被覆盖
func WithReadWriteTimeout(readwriteTimeout int64) ConnManagerOption {
	return func(options *connManageroptions) error {
		options.readwriteTimeout = &readwriteTimeout
		return nil
	}
}

//读超时时间
//若该连接readTimeout时间内没有读取到数据，将会被关闭
func WithReadTimeout(readTimeout int64) ConnManagerOption {
	return func(options *connManageroptions) error {
		options.readTimeout = &readTimeout
		return nil
	}
}

//写超时时间
//若该连接writeTimeout时间内没有写入数据，将会被关闭
func WithWriteTimeout(writeTimeout int64) ConnManagerOption {
	return func(options *connManageroptions) error {
		options.writeTimeout = &writeTimeout
		return nil
	}
}

// connectionTimedOut:连接超时时间
//没什么用，防止hook AfterConn()时间过长
func WithConnectionTimedOut(connectionTimedOut int64) ConnManagerOption {
	return func(options *connManageroptions) error {
		options.connectionTimedOut = &connectionTimedOut
		return nil
	}
}

// transmissionTimeout:传输超时时间
//检查数据包时间戳
func WithTransmissionTimeout(transmissionTimeout int64) ConnManagerOption {
	return func(options *connManageroptions) error {
		options.transmissionTimeout = &transmissionTimeout
		return nil
	}
}

// explorationCycle:探测周期
//业务心跳，检查间隔时间
func WithExplorationCycle(explorationCycle int64) ConnManagerOption {
	return func(options *connManageroptions) error {
		options.explorationCycle = &explorationCycle
		return nil
	}
}

// detectionTimeout:探测超时时间 每个连接探测超时时间，用次参数来监控连接是否正常
//业务心跳，检查超时时间
func WithDetectionTimeout(detectionTimeout int64) ConnManagerOption {
	return func(options *connManageroptions) error {
		options.detectionTimeout = &detectionTimeout
		return nil
	}
}
