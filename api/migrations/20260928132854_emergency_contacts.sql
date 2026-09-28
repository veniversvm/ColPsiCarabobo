-- Persona de contacto para emergencias (datos de un TERCERO: nunca públicos).
--
-- Generado con `atlas migrate diff emergency_contacts --env gorm` y ajustado a mano:
--   1. Se elimina el "DROP INDEX idx_posts_status_publish_at" que Atlas propone. Ese
--      índice se creó en una migración escrita a mano (20260914140000) y el generador
--      desde modelos GORM no lo conoce, así que el diff lo marcaría como huérfano.
--   2. Se añaden los CHECK que respetan la regla de negocio declarada por el Colegio
--      (nombre y parentesco siempre; al menos un canal de contacto). El service ya
--      la valida; esto es defensa en profundidad ante escrituras directas.
-- Create "psi_user_emergency_contacts" table
CREATE TABLE "psi_user_emergency_contacts" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "create_by" character varying(255) NULL,
  "update_by" character varying(255) NULL,
  "create_by_id" uuid NULL,
  "update_by_id" uuid NULL,
  "psi_user_id" uuid NOT NULL,
  "name" character varying(255) NOT NULL,
  "relationship" character varying(100) NOT NULL,
  "phone" character varying(20) NULL,
  "email" character varying(255) NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_psi_users_emergency_contacts" FOREIGN KEY ("psi_user_id") REFERENCES "psi_users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "chk_emergency_contact_identified" CHECK (btrim("name") <> '' AND btrim("relationship") <> ''),
  CONSTRAINT "chk_emergency_contact_channel" CHECK (COALESCE(btrim("phone"), '') <> '' OR COALESCE(btrim("email"), '') <> '')
);
-- Create index "idx_psi_user_emergency_contacts_deleted_at" to table: "psi_user_emergency_contacts"
CREATE INDEX "idx_psi_user_emergency_contacts_deleted_at" ON "psi_user_emergency_contacts" ("deleted_at");
-- Create index "idx_psi_user_emergency_contacts_psi_user_id" to table: "psi_user_emergency_contacts"
CREATE INDEX "idx_psi_user_emergency_contacts_psi_user_id" ON "psi_user_emergency_contacts" ("psi_user_id");
