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
}

// maximumConnection:最大连接数
func WithMaximumConnection(maximumConnection int32) ConnManagerOption {
	return func(options *connManageroptions) error {
		options.maximumConnection = &maximumConnection
		return nil
	}
}

// connectionTimedOut:连接超时时间
func WithConnectionTimedOut(connectionTimedOut int64) ConnManagerOption {
	return func(options *connManageroptions) error {
		options.connectionTimedOut = &connectionTimedOut
		return nil
	}
}

// transmissionTimeout:传输超时时间
func WithTransmissionTimeout(transmissionTimeout int64) ConnManagerOption {
	return func(options *connManageroptions) error {
		options.transmissionTimeout = &transmissionTimeout
		return nil
	}
}

// explorationCycle:探测周期
func WithExplorationCycle(explorationCycle int64) ConnManagerOption {
	return func(options *connManageroptions) error {
		options.explorationCycle = &explorationCycle
		return nil
	}
}

// detectionTimeout:探测超时时间 每个连接探测超时时间，用次参数来监控连接是否正常
func WithDetectionTimeout(detectionTimeout int64) ConnManagerOption {
	return func(options *connManageroptions) error {
		options.detectionTimeout = &detectionTimeout
		return nil
	}
}
