package parseproxy

type Proxy struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"`
	Server   string `yaml:"server"`
	Port     int    `yaml:"port"`
	Cipher   string `yaml:"cipher,omitempty"`
	Password string `yaml:"password,omitempty"`
	Username string `yaml:"username,omitempty"`

	// vmess / vless
	UUID     string   `yaml:"uuid,omitempty"`
	AlterID  int      `yaml:"alter-id,omitempty"`
	Security string   `yaml:"security,omitempty"`
	Network  string   `yaml:"network,omitempty"` // tcp / ws / grpc
	TLS      bool     `yaml:"tls,omitempty"`
	SNI      string   `yaml:"sni,omitempty"`
	ALPN     []string `yaml:"alpn,omitempty"`

	// ws
	WSOpts *WSOptions `yaml:"ws-opts,omitempty"`

	// grpc
	GRPCOpts *GRPCOptions `yaml:"grpc-opts,omitempty"`

	// 扩展字段（关键🔥）
	Raw         map[string]any `yaml:",inline,omitempty"`
	DialerProxy string         `yaml:"dialer-proxy"`
}

type WSOptions struct {
	Path    string            `yaml:"path,omitempty"`
	Headers map[string]string `yaml:"headers,omitempty"`
}

type GRPCOptions struct {
	ServiceName string `yaml:"grpc-service-name,omitempty"`
}
