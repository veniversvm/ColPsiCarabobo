-- Create "psi_user_documents" table
CREATE TABLE "psi_user_documents" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "create_by" character varying(255) NULL,
  "update_by" character varying(255) NULL,
  "create_by_id" uuid NULL,
  "update_by_id" uuid NULL,
  "psi_user_id" uuid NULL,
  "document_type" character varying(50) NOT NULL DEFAULT 'otro',
  "title" character varying(255) NOT NULL,
  "notes" text NULL,
  "document_date" date NULL,
  "s3_key" character varying(512) NOT NULL,
  "filename" character varying(255) NULL,
  "mime_type" character varying(100) NULL,
  "size_bytes" bigint NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_psi_users_documents" FOREIGN KEY ("psi_user_id") REFERENCES "psi_users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_psi_user_documents_deleted_at" to table: "psi_user_documents"
CREATE INDEX "idx_psi_user_documents_deleted_at" ON "psi_user_documents" ("deleted_at");
-- Create index "idx_psi_user_documents_psi_user_id" to table: "psi_user_documents"
CREATE INDEX "idx_psi_user_documents_psi_user_id" ON "psi_user_documents" ("psi_user_id");
