-- Add optional identity/registration fields to psi_inscription_requests
ALTER TABLE "psi_inscription_requests" ADD COLUMN "segundo_nombre" character varying(255) NULL;
ALTER TABLE "psi_inscription_requests" ADD COLUMN "segundo_apellido" character varying(255) NULL;
ALTER TABLE "psi_inscription_requests" ADD COLUMN "genero" character varying(1) NULL;
ALTER TABLE "psi_inscription_requests" ADD COLUMN "titulo_registro_tomo" character varying(100) NULL;
ALTER TABLE "psi_inscription_requests" ADD COLUMN "titulo_registro_folio" character varying(100) NULL;
ALTER TABLE "psi_inscription_requests" ADD COLUMN "psi_user_id" uuid NULL;
CREATE INDEX "idx_inscription_requests_psi_user_id" ON "psi_inscription_requests" ("psi_user_id");
