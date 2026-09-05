-- Create "psi_inscription_requests" table
CREATE TABLE "psi_inscription_requests" (
    "id" uuid NOT NULL DEFAULT uuidv7(),
    "cedula" bigint NOT NULL,
    "nacionalidad" character varying(1) NOT NULL,
    "nombres" character varying(255) NOT NULL,
    "apellidos" character varying(255) NOT NULL,
    "fpv" bigint NULL,
    "telefono" character varying(50) NULL,
    "correo" character varying(255) NOT NULL,
    "fecha_nacimiento" date NULL,
    "titulo_universidad" character varying(255) NULL,
    "titulo_fecha_graduacion" date NULL,
    "titulo_mencion" character varying(255) NULL,
    "titulo_registro_numero" character varying(100) NULL,
    "titulo_registro_estado" character varying(100) NULL,
    "rif" character varying(50) NULL,
    "foto_s3_key" character varying(512) NULL,
    "comprobante_s3_key" character varying(512) NULL,
    "status" character varying(20) NULL DEFAULT 'pending',
    "control_number" character varying(50) NULL,
    "notes" text NULL,
    "created_at" timestamptz NULL,
    "updated_at" timestamptz NULL,
    "deleted_at" timestamptz NULL,
    "create_by" character varying(255) NULL,
    "update_by" character varying(255) NULL,
    "create_by_id" uuid NULL,
    "update_by_id" uuid NULL,
    PRIMARY KEY ("id")
);

-- Unicidad de cédula solo entre solicitudes pendientes
CREATE UNIQUE INDEX "idx_inscription_requests_cedula_pending"
    ON "psi_inscription_requests" ("cedula")
    WHERE "status" = 'pending';

-- Unicidad de número de control (cuando no esté vacío)
CREATE UNIQUE INDEX "idx_inscription_requests_control_number"
    ON "psi_inscription_requests" ("control_number")
    WHERE "control_number" <> '' AND "control_number" IS NOT NULL;

-- Índice para búsquedas por status
CREATE INDEX "idx_inscription_requests_status"
    ON "psi_inscription_requests" ("status");
