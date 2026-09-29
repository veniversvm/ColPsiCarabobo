-- Aceptación de los Términos y Condiciones — Parte II (agremiados).
-- Cada aceptación inserta una fila; la tabla es APPEND-ONLY: no hay UPDATE ni
-- DELETE en el repositorio, así que la fila es evidencia de por sí sola y no
-- se puede reescribir el pasado. Un agremiado que acepta de nuevo porque el
-- texto cambió genera una segunda fila con otra versión.
--
-- La versión vigente NO se guarda en una tabla: es la constante `TermsVersion`
-- del backend (internal/domain/psi_terms_acceptance.model.go), que el frontend
-- recibe por GET /psi/me/terms. Una sola fuente de verdad evita que un
-- despliegue desalineado de la web y la API muestre un texto distinto del que
-- el servidor cree vigente.
--
-- Escrito a mano siguiendo el formato que produce `atlas migrate diff --env gorm`
-- (ver 20260928132854_emergency_contacts.sql): AuditModel embebido + FK GORM
-- (fk_<tabla madre>_<tabla hija>) + índices por deleted_at y por la FK.
-- ON DELETE NO ACTION a propósito: la baja de agremiados es lógica
-- (psi_users.deleted_at) y el historial de aceptación debe sobrevivir a ella.
--
-- Create "psi_terms_acceptance" table
CREATE TABLE "psi_terms_acceptance" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "create_by" character varying(255) NULL,
  "update_by" character varying(255) NULL,
  "create_by_id" uuid NULL,
  "update_by_id" uuid NULL,
  "psi_user_id" uuid NOT NULL,
  "version" character varying(32) NOT NULL,
  "accepted_at" timestamptz NULL,
  "ip" character varying(45) NULL,
  "user_agent" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_psi_users_psi_terms_acceptance" FOREIGN KEY ("psi_user_id") REFERENCES "psi_users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_psi_terms_acceptance_deleted_at" to table: "psi_terms_acceptance"
CREATE INDEX "idx_psi_terms_acceptance_deleted_at" ON "psi_terms_acceptance" ("deleted_at");
-- Create index "idx_psi_terms_acceptance_psi_user_id" to table: "psi_terms_acceptance"
CREATE INDEX "idx_psi_terms_acceptance_psi_user_id" ON "psi_terms_acceptance" ("psi_user_id");
-- La pareja (agremiado, versión) identifica una aceptación: el servicio la
-- consulta para ser idempotente y el índice único cierra la carrera del doble
-- clic (dos POST simultáneos no pueden insertar dos filas de la misma versión).
CREATE UNIQUE INDEX "idx_psi_terms_acceptance_user_version" ON "psi_terms_acceptance" ("psi_user_id", "version");
