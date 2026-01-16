package dto

import (
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	datatype "github.com/pocket-id/pocket-id/backend/internal/model/types"
)

type WebauthnCredentialDto struct {
	ID              string                            `json:"id"`
	Name            string                            `json:"name"`
	CredentialID    string                            `json:"credentialID"`
	AttestationType string                            `json:"attestationType"`
	Transport       []protocol.AuthenticatorTransport `json:"transport"`

	BackupEligible bool `json:"backupEligible"`
	BackupState    bool `json:"backupState"`

	CreatedAt datatype.DateTime `json:"createdAt"`
}

type WebauthnCredentialUpdateDto struct {
	Name string `json:"name" binding:"required,min=1,max=50"`
}

type WebauthnCrossDeviceStartResponseDto struct {
	Code          string            `json:"code"`
	AuthenticatorURL         string            `json:"authenticatorUrl"`
	ExchangeToken string            `json:"exchangeToken"`
	ExpiresAt     datatype.DateTime `json:"expiresAt"`
	RequesterIP   string            `json:"requesterIp"`
	RequesterUserAgent string           `json:"requesterUserAgent"`
}

type WebauthnCrossDeviceLoginStartResponseDto struct {
	Response      protocol.PublicKeyCredentialRequestOptions `json:"response"`
	SessionID     string                                     `json:"sessionId"`
	Timeout       time.Duration                              `json:"timeout"`
	RequesterIP   string                                     `json:"requesterIp"`
	RequesterUserAgent string                                    `json:"requesterUserAgent"`
	ExpiresAt     datatype.DateTime                          `json:"expiresAt"`
}

type WebauthnCrossDeviceStatusResponseDto struct {
	Status    string            `json:"status"`
	ExpiresAt datatype.DateTime `json:"expiresAt"`
	User      *UserDto          `json:"user,omitempty"`
}
