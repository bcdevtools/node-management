package types

type PrivKey struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}
type PubKey struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}
type PrivValidatorKey struct {
	PrivKey *PrivKey `json:"priv_key"`
	PubKey  *PubKey  `json:"pub_key"`
	Address string   `json:"address"`
}
