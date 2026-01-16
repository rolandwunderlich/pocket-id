CREATE TABLE webauthn_cross_device_logins (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    code VARCHAR(64) NOT NULL UNIQUE,
    exchange_token CHAR(64) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    requester_ip VARCHAR(45) NOT NULL,
    requester_user_agent TEXT NOT NULL,
    session_id VARCHAR(36),
    user_id UUID,
    completed_at TIMESTAMP,
    consumed_at TIMESTAMP,
    authenticator_ip VARCHAR(45),
    authenticator_user_agent TEXT,
    CONSTRAINT fk_webauthn_cross_device_login_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_webauthn_cross_device_logins_expires_at ON webauthn_cross_device_logins (expires_at);
CREATE INDEX idx_webauthn_cross_device_logins_code_exchange ON webauthn_cross_device_logins (code, exchange_token);
