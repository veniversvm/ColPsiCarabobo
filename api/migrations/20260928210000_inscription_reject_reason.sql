-- ============================================================================
-- Motivo de rechazo opcional en solicitudes de inscripción.
--
-- El rechazo ya no elimina la ficha: la solicitud pasa a status='rejected'
-- conservando fila, documentos y archivos S3 para su revisión posterior, y
-- este campo guarda el motivo (opcional) escrito por el administrador.
-- Las rechazadas no bloquean re-aplicaciones: los índices únicos vigentes
-- (CI/FPV/correo) aplican solo a solicitudes pendientes y los checks de
-- unicidad del servicio también filtran por status='pending'.
-- ============================================================================

ALTER TABLE "psi_inscription_requests"
  ADD COLUMN IF NOT EXISTS "reject_reason" text NOT NULL DEFAULT '';