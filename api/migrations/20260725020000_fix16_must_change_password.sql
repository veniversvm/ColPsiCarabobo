-- FIX-16: Agregar flag de cambio de contraseña obligatorio
-- Fecha: 2026-07-25
-- Nota: En desarrollo, no hay datos que migrar. Solo cambios de esquema.

-- ═══ Agregar columna must_change_password a user_admins ═══
ALTER TABLE user_admins ADD COLUMN IF NOT EXISTS must_change_password BOOLEAN DEFAULT false;

-- ═══ Agregar columna must_change_password a psi_users ═══
ALTER TABLE psi_users ADD COLUMN IF NOT EXISTS must_change_password BOOLEAN DEFAULT false;
