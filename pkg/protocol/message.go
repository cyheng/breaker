package protocol

type Login struct {
	Version      string `json:"version"`
	Hostname     string `json:"hostname"`
	Os           string `json:"os"`
	Arch         string `json:"arch"`
	User         string `json:"user"`
	PrivilegeKey string `json:"privilege_key"`
	Timestamp    int64  `json:"timestamp"`
	RunID        string `json:"run_id"`
}

type LoginResponse struct {
	Version       string `json:"version"`
	RunID         string `json:"run_id"`
	ServerUDPPort int    `json:"server_udp_port"`
	Error         string `json:"error"`
}

type NewProxy struct {
	ProxyName      string            `json:"proxy_name"`
	ProxyType      string            `json:"proxy_type"`
	UseEncryption  bool              `json:"use_encryption"`
	UseCompression bool              `json:"use_compression"`
	Metas          map[string]string `json:"metas"`
	// tcp and udp only
	RemotePort        int               `json:"remote_port"`
	HostHeaderRewrite string            `json:"host_header_rewrite"`
	Headers           map[string]string `json:"headers"`
}

type NewProxyResponse struct {
	ProxyName  string `json:"proxy_name"`
	RemoteAddr string `json:"remote_addr"`
	Error      string `json:"error"`
}

type CloseProxy struct {
	ProxyName string `json:"proxy_name"`
}

type NewWorkConn struct {
	RunID        string `json:"run_id"`
	PrivilegeKey string `json:"privilege_key"`
	Timestamp    int64  `json:"timestamp"`
}

type ReqWorkConn struct {
}

type StartWorkConn struct {
	ProxyName string `json:"proxy_name"`
	SrcAddr   string `json:"src_addr"`
	DstAddr   string `json:"dst_addr"`
	SrcPort   uint16 `json:"src_port"`
	DstPort   uint16 `json:"dst_port"`
	Error     string `json:"error"`
}
