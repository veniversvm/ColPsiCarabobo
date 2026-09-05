-- Modify "user_admins" table
ALTER TABLE "user_admins" ALTER COLUMN "is_active" SET DEFAULT false, ADD COLUMN "can_manage_tickets" boolean NULL DEFAULT false;
-- Create "ticket_areas" table
CREATE TABLE "ticket_areas" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "create_by" character varying(255) NULL,
  "update_by" character varying(255) NULL,
  "create_by_id" uuid NULL,
  "update_by_id" uuid NULL,
  "name" character varying(120) NOT NULL,
  "description" character varying(500) NULL,
  "tickets_per_psi" bigint NOT NULL DEFAULT 5,
  PRIMARY KEY ("id")
);
-- Create index "idx_ticket_areas_deleted_at" to table: "ticket_areas"
CREATE INDEX "idx_ticket_areas_deleted_at" ON "ticket_areas" ("deleted_at");
-- Create "ticket_motivos" table
CREATE TABLE "ticket_motivos" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "create_by" character varying(255) NULL,
  "update_by" character varying(255) NULL,
  "create_by_id" uuid NULL,
  "update_by_id" uuid NULL,
  "area_id" bigint NOT NULL,
  "name" character varying(120) NOT NULL,
  "description" character varying(500) NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_ticket_areas_motivos" FOREIGN KEY ("area_id") REFERENCES "ticket_areas" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_ticket_motivos_area_id" to table: "ticket_motivos"
CREATE INDEX "idx_ticket_motivos_area_id" ON "ticket_motivos" ("area_id");
-- Create index "idx_ticket_motivos_deleted_at" to table: "ticket_motivos"
CREATE INDEX "idx_ticket_motivos_deleted_at" ON "ticket_motivos" ("deleted_at");
-- Create "ticket_estados" table
CREATE TABLE "ticket_estados" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "create_by" character varying(255) NULL,
  "update_by" character varying(255) NULL,
  "create_by_id" uuid NULL,
  "update_by_id" uuid NULL,
  "motivo_id" bigint NOT NULL,
  "name" character varying(60) NOT NULL,
  "order" bigint NOT NULL DEFAULT 0,
  "is_closed" boolean NOT NULL DEFAULT false,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_ticket_motivos_estados" FOREIGN KEY ("motivo_id") REFERENCES "ticket_motivos" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_ticket_estados_deleted_at" to table: "ticket_estados"
CREATE INDEX "idx_ticket_estados_deleted_at" ON "ticket_estados" ("deleted_at");
-- Create index "idx_ticket_estados_motivo_id" to table: "ticket_estados"
CREATE INDEX "idx_ticket_estados_motivo_id" ON "ticket_estados" ("motivo_id");
-- Create index "idx_ticket_estados_motivo_nombre" to table: "ticket_estados"
CREATE UNIQUE INDEX "idx_ticket_estados_motivo_nombre" ON "ticket_estados" ("motivo_id", "name");
-- Create "tickets" table
CREATE TABLE "tickets" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "create_by" character varying(255) NULL,
  "update_by" character varying(255) NULL,
  "create_by_id" uuid NULL,
  "update_by_id" uuid NULL,
  "psi_user_id" uuid NOT NULL,
  "area_id" bigint NOT NULL,
  "motivo_id" bigint NOT NULL,
  "estado_id" bigint NOT NULL,
  "title" character varying(200) NOT NULL,
  "description" text NULL,
  "close_reason" character varying(500) NULL,
  "closed_by_type" character varying(10) NULL,
  "closed_by_admin_id" uuid NULL,
  "closed_by_psi_id" uuid NULL,
  "closed_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_tickets_area" FOREIGN KEY ("area_id") REFERENCES "ticket_areas" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_tickets_estado" FOREIGN KEY ("estado_id") REFERENCES "ticket_estados" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_tickets_motivo" FOREIGN KEY ("motivo_id") REFERENCES "ticket_motivos" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_tickets_psi" FOREIGN KEY ("psi_user_id") REFERENCES "psi_users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_tickets_deleted_at" to table: "tickets"
CREATE INDEX "idx_tickets_deleted_at" ON "tickets" ("deleted_at");
-- Create index "idx_tickets_estado_id" to table: "tickets"
CREATE INDEX "idx_tickets_estado_id" ON "tickets" ("estado_id");
-- Create index "idx_tickets_motivo_id" to table: "tickets"
CREATE INDEX "idx_tickets_motivo_id" ON "tickets" ("motivo_id");
-- Create index "idx_tickets_psi_area" to table: "tickets"
CREATE INDEX "idx_tickets_psi_area" ON "tickets" ("psi_user_id", "area_id");
-- Create "ticket_mensajes" table
CREATE TABLE "ticket_mensajes" (
  "id" bigserial NOT NULL,
  "ticket_id" bigint NOT NULL,
  "author_type" character varying(10) NOT NULL,
  "author_admin_id" uuid NULL,
  "author_psi_id" uuid NULL,
  "message" text NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_ticket_mensajes_admin" FOREIGN KEY ("author_admin_id") REFERENCES "user_admins" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_ticket_mensajes_psi" FOREIGN KEY ("author_psi_id") REFERENCES "psi_users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_tickets_mensajes" FOREIGN KEY ("ticket_id") REFERENCES "tickets" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_ticket_mensajes_author_admin_id" to table: "ticket_mensajes"
CREATE INDEX "idx_ticket_mensajes_author_admin_id" ON "ticket_mensajes" ("author_admin_id");
-- Create index "idx_ticket_mensajes_author_psi_id" to table: "ticket_mensajes"
CREATE INDEX "idx_ticket_mensajes_author_psi_id" ON "ticket_mensajes" ("author_psi_id");
-- Create index "idx_ticket_mensajes_ticket_id" to table: "ticket_mensajes"
CREATE INDEX "idx_ticket_mensajes_ticket_id" ON "ticket_mensajes" ("ticket_id");
-- Create "ticket_adjuntos" table
CREATE TABLE "ticket_adjuntos" (
  "id" bigserial NOT NULL,
  "mensaje_id" bigint NOT NULL,
  "s3_key" character varying(512) NOT NULL,
  "original_name" character varying(255) NULL,
  "mime_type" character varying(100) NULL,
  "size_bytes" bigint NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_ticket_mensajes_adjuntos" FOREIGN KEY ("mensaje_id") REFERENCES "ticket_mensajes" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_ticket_adjuntos_mensaje_id" to table: "ticket_adjuntos"
CREATE INDEX "idx_ticket_adjuntos_mensaje_id" ON "ticket_adjuntos" ("mensaje_id");
-- Create "ticket_status_logs" table
CREATE TABLE "ticket_status_logs" (
  "id" bigserial NOT NULL,
  "ticket_id" bigint NOT NULL,
  "previous_state_id" bigint NULL,
  "new_state_id" bigint NOT NULL,
  "changed_by_type" character varying(10) NOT NULL,
  "changed_by_admin_id" uuid NULL,
  "changed_by_psi_id" uuid NULL,
  "reason" character varying(500) NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_ticket_status_logs_new_state" FOREIGN KEY ("new_state_id") REFERENCES "ticket_estados" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_ticket_status_logs_previous_state" FOREIGN KEY ("previous_state_id") REFERENCES "ticket_estados" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_tickets_status_logs" FOREIGN KEY ("ticket_id") REFERENCES "tickets" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_ticket_status_logs_new_state_id" to table: "ticket_status_logs"
CREATE INDEX "idx_ticket_status_logs_new_state_id" ON "ticket_status_logs" ("new_state_id");
-- Create index "idx_ticket_status_logs_previous_state_id" to table: "ticket_status_logs"
CREATE INDEX "idx_ticket_status_logs_previous_state_id" ON "ticket_status_logs" ("previous_state_id");
-- Create index "idx_ticket_status_logs_ticket_id" to table: "ticket_status_logs"
CREATE INDEX "idx_ticket_status_logs_ticket_id" ON "ticket_status_logs" ("ticket_id");
