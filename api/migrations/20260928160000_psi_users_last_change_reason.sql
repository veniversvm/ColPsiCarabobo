-- Motivo del último cambio de expediente realizado por un admin.
--
-- Escrito a mano (ALTER TABLE simple): el modelo GORM se declara con AutoMigrate,
-- pero esta columna la escribe el servicio psi_user_admin_service.go solo cuando
-- el admin envía el motivo (PATCH /admin/psi/:id). Es un dato INTERNO: los DTOs
-- públicos (directorio, detalle público, sitemap) no lo exponen.
-- Add column "last_change_reason" to table: "psi_users"
ALTER TABLE "psi_users" ADD COLUMN "last_change_reason" character varying(500) NULL;