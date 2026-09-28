-- ============================================================================
-- Respaldo a nivel de base de datos contra duplicados en solicitudes de
-- inscripción pendientes. El servicio ya valida CI/FPV/correo en Submit y
-- UpdateFicha (contra psi_users y contra otras solicitudes pendientes); estos
-- índices cierran la ventana de carrera (TOCTOU) entre la validación y el
-- INSERT, igual que el índice parcial de cédula ya existente
-- (idx_inscription_requests_cedula_pending).
--   • FPV: 0 y NULL significan "sin FPV"; por eso el predicado es fpv > 0
--     (con "fpv IS NOT NULL" el índice bloquearía todas las fichas sin FPV).
--   • Correo: único case-insensitive, alineado a ExistsPendingEmail.
-- ============================================================================

-- FPV único entre solicitudes pendientes (backstop DB contra doble submit)
CREATE UNIQUE INDEX "idx_inscription_requests_fpv_pending"
  ON "psi_inscription_requests" ("fpv")
  WHERE ((status)::text = 'pending'::text) AND (fpv > 0) AND (deleted_at IS NULL);

-- Correo único (case-insensitive) entre solicitudes pendientes
CREATE UNIQUE INDEX "idx_inscription_requests_correo_pending"
  ON "psi_inscription_requests" (LOWER(correo))
  WHERE ((status)::text = 'pending'::text) AND ((correo)::text <> ''::text) AND (deleted_at IS NULL);