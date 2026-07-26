-- Migración: FIX-03 — Hash SHA-256 de claves HMAC antes de almacenar en DB
-- Fecha: 2026-07-25
-- Descripción: Convierte UUIDs raw (36 chars) a hashes SHA-256 (64 chars hex)
--              en las columnas 'key' de user_admins y psi_user_models.
--              Esto previene que claves de firma JWT queden en texto plano.

-- ═══ Habilitar extensión pgcrypto (provee sha256) ═══
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ═══ Hashear keys de user_admins (solo UUIDs raw de 36 chars) ═══
UPDATE user_admins
SET key = encode(sha256(key::bytea), 'hex')
WHERE key != ''
  AND length(key) = 36;

-- ═══ Hashear keys de psi_user_models (solo UUIDs raw de 36 chars) ═══
UPDATE psi_user_models
SET key = encode(sha256(key::bytea), 'hex')
WHERE key != ''
  AND length(key) = 36;
