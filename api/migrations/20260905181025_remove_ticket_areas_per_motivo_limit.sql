-- Modify "ticket_motivos" table
ALTER TABLE "ticket_motivos" DROP COLUMN "area_id", ADD COLUMN "tickets_per_psi" bigint NOT NULL DEFAULT 3;
-- Modify "tickets" table
ALTER TABLE "tickets" DROP COLUMN "area_id";
-- Create index "idx_tickets_psi_user" to table: "tickets"
CREATE INDEX "idx_tickets_psi_user" ON "tickets" ("psi_user_id");
-- Drop "ticket_areas" table
DROP TABLE "ticket_areas";
