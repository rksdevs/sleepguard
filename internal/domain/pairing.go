package domain

import "time"

// PushSubscription is the browser push subscription payload.
type PushSubscription struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		P256dh string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
}

// PairingRequest registers a phone for push notifications.
type PairingRequest struct {
	DeviceID     string           `json:"device_id"`
	Name         string           `json:"name"`
	Subscription PushSubscription `json:"subscription"`
}

// PairedClient is a registered notification target.
type PairedClient struct {
	ID           string     `json:"id"`
	DeviceID     string     `json:"device_id"`
	Name         string     `json:"name"`
	NotifyOnRise bool       `json:"notify_on_rise"`
	CreatedAt    time.Time  `json:"created_at"`
	RevokedAt    *time.Time `json:"revoked_at,omitempty"`
}
