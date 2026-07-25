-- Migración: FIX-41 (GraduationYear string → int) + FIX-42 (UUID consistency) + FIX-47 (Email size)
-- Fecha: 2026-07-25
-- Nota: En desarrollo, no hay datos que migrar. Solo cambios de esquema.

-- ═══ FIX-42: Habilitar extensión pg_uuidv7 ═══
CREATE EXTENSION IF NOT EXISTS "pg_uuidv7";

-- ═══ FIX-42: Cambiar defaults de UUID en tablas de analytics ═══
ALTER TABLE login_events ALTER COLUMN id SET DEFAULT uuidv7();
ALTER TABLE active_sessions ALTER COLUMN id SET DEFAULT uuidv7();

-- ═══ FIX-41: Cambiar graduation_year de varchar(50) a bigint ═══
ALTER TABLE psi_user_post_grades ALTER COLUMN graduation_year TYPE bigint USING graduation_year::bigint;

-- ═══ FIX-47: Unificar tamaño de email en psi_users (de varchar(50) a varchar(255)) ═══
ALTER TABLE psi_users ALTER COLUMN email TYPE varchar(255);
