-- Create "psi_password_reset_tokens" table
CREATE TABLE "psi_password_reset_tokens" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "psi_id" uuid NOT NULL,
  "token_hash" character varying(64) NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "used_at" timestamptz NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_psi_password_reset_tokens_psi" FOREIGN KEY ("psi_id") REFERENCES "psi_users" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_psi_password_reset_tokens_psi_id" to table: "psi_password_reset_tokens"
CREATE INDEX "idx_psi_password_reset_tokens_psi_id" ON "psi_password_reset_tokens" ("psi_id");
-- Create index "idx_reset_token_hash" to table: "psi_password_reset_tokens"
CREATE UNIQUE INDEX "idx_reset_token_hash" ON "psi_password_reset_tokens" ("token_hash");
