package types

type HandshakeMessage struct {
	TunnelID string  `json:"tunnelID"`
	APIKey   string  `json:"apiKey"`
	Bearer   *string `json:"bearer"`
}

type TunnelConfig struct {
	TunnelID string  `yaml:"tunnel_id"`
	Bearer   *string `yaml:"bearer,omitempty"`
}

type TunnelClientConfig struct {
	PersistConfig bool
	DisableAuth   bool
}
