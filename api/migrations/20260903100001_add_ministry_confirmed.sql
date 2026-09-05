-- Add "ministry_registration_confirmed" column to "psi_user_col_data"
ALTER TABLE "psi_user_col_data" ADD COLUMN "ministry_registration_confirmed" boolean NULL DEFAULT false;

-- Change default of "is_active" in "psi_users" from true to false.
-- Los psicólogos creados desde inscripciones aprobadas nacerán inactivos hasta
-- que la administración confirme los 3 requisitos legales.
ALTER TABLE "psi_users" ALTER COLUMN "is_active" SET DEFAULT false;
