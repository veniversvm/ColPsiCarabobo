-- Create "admin_permission_logs" table
CREATE TABLE "admin_permission_logs" (
  "id" uuid NOT NULL,
  "target_admin_id" uuid NULL,
  "target_username" character varying(100) NULL,
  "action" character varying(50) NULL,
  "changed_by_id" uuid NULL,
  "changed_by_username" character varying(100) NULL,
  "permissions_changed" text NULL,
  "role_from" character varying(50) NULL,
  "role_to" character varying(50) NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_admin_permission_logs_changed_by_id" to table: "admin_permission_logs"
CREATE INDEX "idx_admin_permission_logs_changed_by_id" ON "admin_permission_logs" ("changed_by_id");
-- Create index "idx_admin_permission_logs_target_admin_id" to table: "admin_permission_logs"
CREATE INDEX "idx_admin_permission_logs_target_admin_id" ON "admin_permission_logs" ("target_admin_id");