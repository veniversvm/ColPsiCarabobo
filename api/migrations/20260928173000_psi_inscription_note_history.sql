-- Histórico de notas administrativas de inscripciones (1:N).
-- Cada guardado con cambio de texto agrega una versión (texto completo, autor y
-- fecha); la columna `notes` de psi_inscription_requests mantiene la nota actual.
-- El histórico nace vacío: solo cuenta lo que se guarde a partir del despliegue.
--
-- Escrito a mano siguiendo el formato que produce `atlas migrate diff --env gorm`
-- (ver 20260928132854_emergency_contacts.sql): AuditModel embebido + FK GORM
-- (fk_<tabla madre>_<tabla hija>) + índices por deleted_at y por la FK.
-- Create "psi_inscription_notes" table
CREATE TABLE "psi_inscription_notes" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "create_by" character varying(255) NULL,
  "update_by" character varying(255) NULL,
  "create_by_id" uuid NULL,
  "update_by_id" uuid NULL,
  "inscription_request_id" uuid NOT NULL,
  "notes" text NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_psi_inscription_requests_psi_inscription_notes" FOREIGN KEY ("inscription_request_id") REFERENCES "psi_inscription_requests" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_psi_inscription_notes_deleted_at" to table: "psi_inscription_notes"
CREATE INDEX "idx_psi_inscription_notes_deleted_at" ON "psi_inscription_notes" ("deleted_at");
-- Create index "idx_psi_inscription_notes_inscription_request_id" to table: "psi_inscription_notes"
CREATE INDEX "idx_psi_inscription_notes_inscription_request_id" ON "psi_inscription_notes" ("inscription_request_id");