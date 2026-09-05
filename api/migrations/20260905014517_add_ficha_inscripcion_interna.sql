-- Modify "psi_inscription_requests" table
ALTER TABLE "psi_inscription_requests" ADD COLUMN "service_address" character varying(255) NULL, ADD COLUMN "municipality_carabobo" character varying(255) NULL, ADD COLUMN "state_outside" character varying(255) NULL, ADD COLUMN "municipality_out_side_carabobo" character varying(255) NULL, ADD COLUMN "country" character varying(255) NULL, ADD COLUMN "service_modality_presencial" boolean NULL DEFAULT false, ADD COLUMN "service_modality_distance" boolean NULL DEFAULT false, ADD COLUMN "service_modality_telephone" boolean NULL DEFAULT false, ADD COLUMN "primary_specialty_id" bigint NULL, ADD COLUMN "secondary_specialty_id" bigint NULL;
-- Create index "idx_psi_inscription_requests_deleted_at" to table: "psi_inscription_requests"
CREATE INDEX "idx_psi_inscription_requests_deleted_at" ON "psi_inscription_requests" ("deleted_at");
-- Create "psi_inscription_documents" table
CREATE TABLE "psi_inscription_documents" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "create_by" character varying(255) NULL,
  "update_by" character varying(255) NULL,
  "create_by_id" uuid NULL,
  "update_by_id" uuid NULL,
  "inscription_request_id" uuid NOT NULL,
  "document_type" character varying(50) NOT NULL,
  "s3_key" character varying(512) NOT NULL,
  "title" character varying(255) NULL,
  "notes" text NULL,
  "original_filename" character varying(255) NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_psi_inscription_requests_documents" FOREIGN KEY ("inscription_request_id") REFERENCES "psi_inscription_requests" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_inscription_document_type" to table: "psi_inscription_documents"
CREATE UNIQUE INDEX "idx_inscription_document_type" ON "psi_inscription_documents" ("inscription_request_id", "document_type");
-- Create index "idx_psi_inscription_documents_deleted_at" to table: "psi_inscription_documents"
CREATE INDEX "idx_psi_inscription_documents_deleted_at" ON "psi_inscription_documents" ("deleted_at");