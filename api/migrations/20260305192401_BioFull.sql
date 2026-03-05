-- Modify "psi_users" table
ALTER TABLE "psi_users" ALTER COLUMN "bio_text_id" TYPE uuid, ADD CONSTRAINT "fk_psi_users_full_bio" FOREIGN KEY ("bio_text_id") REFERENCES "text_models" ("id") ON UPDATE CASCADE ON DELETE SET NULL;
