-- Create "app_settings" table
CREATE TABLE "app_settings" (
  "key" character varying(80) NOT NULL,
  "value" jsonb NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("key")
);
-- Create "settings_audit_logs" table
CREATE TABLE "settings_audit_logs" (
  "id" uuid NOT NULL,
  "changed_by_id" uuid NULL,
  "changed_by_username" character varying(100) NULL,
  "key" character varying(80) NULL,
  "enabled_from" boolean NULL,
  "enabled_to" boolean NULL,
  "message_from" character varying(500) NULL,
  "message_to" character varying(500) NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_settings_audit_logs_changed_by_id" to table: "settings_audit_logs"
CREATE INDEX "idx_settings_audit_logs_changed_by_id" ON "settings_audit_logs" ("changed_by_id");
-- Create index "idx_settings_audit_logs_key" to table: "settings_audit_logs"
CREATE INDEX "idx_settings_audit_logs_key" ON "settings_audit_logs" ("key");
