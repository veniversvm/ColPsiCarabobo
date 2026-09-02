-- Modify "psi_user_col_data" table
ALTER TABLE "psi_user_col_data" ADD COLUMN "birthday_notification" boolean NULL DEFAULT false;
-- Modify "psi_users" table
ALTER TABLE "psi_users" ADD COLUMN "service_modality_presencial" boolean NULL DEFAULT false, ADD COLUMN "service_modality_distance" boolean NULL DEFAULT false, ADD COLUMN "service_modality_telephone" boolean NULL DEFAULT false, ADD COLUMN "show_service_modality" boolean NULL DEFAULT false;
