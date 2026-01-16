package model

import datatype "github.com/pocket-id/pocket-id/backend/internal/model/types"

// WebauthnCrossDeviceLogin stores a short-lived cross-device login request initiated from a requester client.
// A secondary client (authenticator) scans the QR code to navigate to a page to perform a WebAuthn login;
// the requester later polls with an exchange token to receive the access token once the authenticator has
// completed authentication.
type WebauthnCrossDeviceLogin struct {
	Base

	Code          string `gorm:"uniqueIndex"`
	ExchangeToken string
	ExpiresAt     datatype.DateTime

	RequesterIP        string
	RequesterUserAgent string

	SessionID *string
	UserID    *string
	User      *User `gorm:"constraint:OnDelete:CASCADE"`

	CompletedAt         *datatype.DateTime
	ConsumedAt          *datatype.DateTime
	AuthenticatorIP        *string
	AuthenticatorUserAgent *string
}
