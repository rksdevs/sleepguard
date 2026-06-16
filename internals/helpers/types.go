package helpers

type Events struct {
	Timestamp int64  `json:"timestamp"` //we will create a timestamp helper that will convert timestamp into int64 based on a given epoch
	Type      string `json:"type"`
	Source    string `json:"source"`
	State     string `json:"state"`
}