-- FIX-31: Agregar FK de áreas de trabajo a psi_users
-- Agrega primary_specialty_id y secondary_specialty_id como FK nullable,
-- migra datos existentes desde strings, y sincroniza las columnas legacy.

-- 1. Agregar columnas nullable
ALTER TABLE psi_users ADD COLUMN primary_specialty_id INTEGER;
ALTER TABLE psi_users ADD COLUMN secondary_specialty_id INTEGER;

-- 2. Migrar datos existentes (strings → IDs)
UPDATE psi_users p
SET primary_specialty_id = s.id
FROM psi_specialty_models s
WHERE p.primary_work_area = s.name;

UPDATE psi_users p
SET secondary_specialty_id = s.id
FROM psi_specialty_models s
WHERE p.secondary_work_area = s.name;

-- 3. FK constraints (ON DELETE SET NULL para no borrar usuarios si se borra una especialidad)
ALTER TABLE psi_users ADD CONSTRAINT fk_psi_users_primary_specialty
    FOREIGN KEY (primary_specialty_id) REFERENCES psi_specialty_models(id) ON DELETE SET NULL;
ALTER TABLE psi_users ADD CONSTRAINT fk_psi_users_secondary_specialty
    FOREIGN KEY (secondary_specialty_id) REFERENCES psi_specialty_models(id) ON DELETE SET NULL;
