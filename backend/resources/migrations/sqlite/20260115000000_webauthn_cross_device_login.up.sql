CREATE TABLE webauthn_cross_device_logins (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    code TEXT NOT NULL UNIQUE,
    exchange_token TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    requester_ip TEXT NOT NULL,
    requester_user_agent TEXT NOT NULL,
    session_id TEXT,
    user_id TEXT,
    completed_at TIMESTAMP,
    consumed_at TIMESTAMP,
    authenticator_ip TEXT,
    authenticator_user_agent TEXT,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_webauthn_cross_device_logins_expires_at ON webauthn_cross_device_logins (expires_at);
CREATE INDEX idx_webauthn_cross_device_logins_code_exchange ON webauthn_cross_device_logins (code, exchange_token);
