package types

type HandshakeMessage struct {
	TunnelID string `json:"tunnelID"`
	APIKey   string `json:"apiKey"`
	Bearer   string `json:"bearer"`
}
