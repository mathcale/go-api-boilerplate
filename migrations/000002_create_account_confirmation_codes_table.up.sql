CREATE TABLE IF NOT EXISTS account_confirmation_codes (
  id UUID NOT NULL DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  code VARCHAR(64) NOT NULL,
  purpose VARCHAR(32) NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  used BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT account_confirmation_codes_pk PRIMARY KEY (id),
  CONSTRAINT account_confirmation_codes_user_id_fk FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS account_confirmation_codes_lookup_idx ON account_confirmation_codes (user_id, code, purpose);
