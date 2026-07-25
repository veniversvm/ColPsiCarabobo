-- Atlas Migration: Documentar política de DELETE en Foreign Keys
-- Fecha: 2026-07-25
-- FIX-18: Documentar que los deletes son únicamente lógicos (soft delete)
--
-- POLÍTICA: Todos los deletes en el sistema son LÓGICOS (soft delete via deleted_at).
-- Las Foreign Keys con ON DELETE NO ACTION actúan como red de seguridad:
-- impiden hard deletes accidentales que dejarían registros huérfanos.
--
-- Si en el futuro se necesita purga de datos (hard delete), crear una función
-- SQL que primero haga soft delete de todos los hijos, luego del padre.

-- ═══════════════════════════════════════════════════════════════════════════════
-- POLÍTICA DE DELETE DOCUMENTADA POR FK
-- ═══════════════════════════════════════════════════════════════════════════════

COMMENT ON CONSTRAINT fk_psi_users_col_data ON psi_user_col_data IS
    'ON DELETE NO ACTION — Soft delete only. Hard delete requiere borrar hijos primero.';

COMMENT ON CONSTRAINT fk_psi_users_post_grades ON psi_user_post_grades IS
    'ON DELETE NO ACTION — Soft delete only. Hard delete requiere borrar hijos primero.';

COMMENT ON CONSTRAINT fk_psi_users_social_networks ON psi_user_social_networks IS
    'ON DELETE NO ACTION — Soft delete only. Hard delete requiere borrar hijos primero.';

COMMENT ON CONSTRAINT fk_psi_users_solvencies ON psi_user_solvency IS
    'ON DELETE NO ACTION — Soft delete only. Hard delete requiere borrar hijos primero.';

COMMENT ON CONSTRAINT fk_posts_text ON posts IS
    'ON DELETE NO ACTION — Soft delete only. Hard delete requiere borrar hijos primero.';

COMMENT ON CONSTRAINT fk_psi_users_full_bio ON psi_users IS
    'ON DELETE NO ACTION — Soft delete only. Bio es opcional, puede ser NULL.';
