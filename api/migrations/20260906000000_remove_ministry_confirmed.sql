-- Remove "ministry_registration_confirmed" column from "psi_user_col_data".
-- El N° de FPV ya acredita la inscripción ministerial; se elimina el requisito
-- independiente de activación (Art. 5 Ley de Ejercicio + Art. 18 Estatutos FPV).
ALTER TABLE "psi_user_col_data" DROP COLUMN IF EXISTS "ministry_registration_confirmed";