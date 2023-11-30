package server

// ServerOption 服务器选项
type ServerOption func(options *serveroptions) error
type serveroptions struct {
	ip         *string
	port       *int64
	servername *string
}

// ip:ipv4地址
func WithIP(ip string) ServerOption {
	return func(options *serveroptions) error {
		options.ip = &ip
		return nil
	}
}

// port:端口号
func WithPort(port int64) ServerOption {
	return func(options *serveroptions) error {
		options.port = &port
		return nil
	}
}

// servername:服务器名称
func WithServerName(servername string) ServerOption {
	return func(options *serveroptions) error {
		options.servername = &servername
		return nil
	}
}
