-- FIX-37: Corregir tipo de columna phone_carabobo
-- Era: text DEFAULT 'false' (string literal "false")
-- Debe ser: varchar(20) DEFAULT '' (teléfono o vacío)

-- Limpiar datos basura: reemplazar el literal 'false' por cadena vacía
UPDATE psi_users SET phone_carabobo = '' WHERE phone_carabobo = 'false';

-- Corregir tipo y default
ALTER TABLE psi_users ALTER COLUMN phone_carabobo TYPE varchar(20);
ALTER TABLE psi_users ALTER COLUMN phone_carabobo SET DEFAULT '';
