-- Modify "user_admins" table
ALTER TABLE "user_admins" ADD COLUMN "role" character varying(50) NULL, ADD COLUMN "can_read_psi" boolean NULL DEFAULT false;