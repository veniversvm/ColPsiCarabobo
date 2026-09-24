-- Modify "user_admins" table
ALTER TABLE "user_admins" ADD COLUMN "can_view_logs" boolean NULL DEFAULT false, ADD COLUMN "can_export_logs" boolean NULL DEFAULT false;
-- Create "api_change_logs" table
CREATE TABLE "api_change_logs" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "entity" character varying(50) NOT NULL,
  "entity_id" character varying(64) NULL,
  "entity_label" character varying(255) NULL,
  "action" character varying(50) NULL,
  "actor_id" uuid NULL,
  "actor_role" character varying(50) NULL,
  "actor_username" character varying(100) NULL,
  "ip" character varying(64) NULL,
  "user_agent" character varying(255) NULL,
  "changes" jsonb NULL,
  "metadata" jsonb NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_api_change_logs_action" to table: "api_change_logs"
CREATE INDEX "idx_api_change_logs_action" ON "api_change_logs" ("action");
-- Create index "idx_api_change_logs_created_at" to table: "api_change_logs"
CREATE INDEX "idx_api_change_logs_created_at" ON "api_change_logs" ("created_at");
-- Create index "idx_audit_actor_created" to table: "api_change_logs"
CREATE INDEX "idx_audit_actor_created" ON "api_change_logs" ("actor_id", "created_at");
-- Create index "idx_audit_entity_action" to table: "api_change_logs"
CREATE INDEX "idx_audit_entity_action" ON "api_change_logs" ("entity", "action", "created_at");
-- Create index "idx_audit_entity_created" to table: "api_change_logs"
CREATE INDEX "idx_audit_entity_created" ON "api_change_logs" ("entity", "entity_id", "created_at");
