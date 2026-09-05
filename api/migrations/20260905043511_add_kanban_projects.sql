-- Create "kanban_projects" table
CREATE TABLE "kanban_projects" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "create_by" character varying(255) NULL,
  "update_by" character varying(255) NULL,
  "create_by_id" uuid NULL,
  "update_by_id" uuid NULL,
  "name" character varying(120) NOT NULL,
  "description" character varying(500) NULL,
  "owner_id" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_kanban_projects_owner" FOREIGN KEY ("owner_id") REFERENCES "user_admins" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_kanban_projects_deleted_at" to table: "kanban_projects"
CREATE INDEX "idx_kanban_projects_deleted_at" ON "kanban_projects" ("deleted_at");
-- Create index "idx_kanban_projects_owner_id" to table: "kanban_projects"
CREATE INDEX "idx_kanban_projects_owner_id" ON "kanban_projects" ("owner_id");
-- Create "kanban_columns" table
CREATE TABLE "kanban_columns" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "create_by" character varying(255) NULL,
  "update_by" character varying(255) NULL,
  "create_by_id" uuid NULL,
  "update_by_id" uuid NULL,
  "project_id" uuid NOT NULL,
  "title" character varying(120) NOT NULL,
  "position" bigint NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_kanban_projects_columns" FOREIGN KEY ("project_id") REFERENCES "kanban_projects" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_kanban_columns_deleted_at" to table: "kanban_columns"
CREATE INDEX "idx_kanban_columns_deleted_at" ON "kanban_columns" ("deleted_at");
-- Create index "idx_kanban_columns_project_id" to table: "kanban_columns"
CREATE INDEX "idx_kanban_columns_project_id" ON "kanban_columns" ("project_id");
-- Create "kanban_cards" table
CREATE TABLE "kanban_cards" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "create_by" character varying(255) NULL,
  "update_by" character varying(255) NULL,
  "create_by_id" uuid NULL,
  "update_by_id" uuid NULL,
  "project_id" uuid NOT NULL,
  "column_id" uuid NOT NULL,
  "title" character varying(200) NOT NULL,
  "description" character varying(2000) NULL,
  "position" bigint NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_kanban_columns_cards" FOREIGN KEY ("column_id") REFERENCES "kanban_columns" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_kanban_cards_column_id" to table: "kanban_cards"
CREATE INDEX "idx_kanban_cards_column_id" ON "kanban_cards" ("column_id");
-- Create index "idx_kanban_cards_deleted_at" to table: "kanban_cards"
CREATE INDEX "idx_kanban_cards_deleted_at" ON "kanban_cards" ("deleted_at");
-- Create index "idx_kanban_cards_project_id" to table: "kanban_cards"
CREATE INDEX "idx_kanban_cards_project_id" ON "kanban_cards" ("project_id");
-- Create "kanban_card_notes" table
CREATE TABLE "kanban_card_notes" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "create_by" character varying(255) NULL,
  "update_by" character varying(255) NULL,
  "create_by_id" uuid NULL,
  "update_by_id" uuid NULL,
  "card_id" uuid NOT NULL,
  "content" character varying(500) NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_kanban_cards_notes" FOREIGN KEY ("card_id") REFERENCES "kanban_cards" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_kanban_card_notes_card_id" to table: "kanban_card_notes"
CREATE INDEX "idx_kanban_card_notes_card_id" ON "kanban_card_notes" ("card_id");
-- Create index "idx_kanban_card_notes_deleted_at" to table: "kanban_card_notes"
CREATE INDEX "idx_kanban_card_notes_deleted_at" ON "kanban_card_notes" ("deleted_at");
-- Create "kanban_project_members" table
CREATE TABLE "kanban_project_members" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "create_by" character varying(255) NULL,
  "update_by" character varying(255) NULL,
  "create_by_id" uuid NULL,
  "update_by_id" uuid NULL,
  "project_id" uuid NOT NULL,
  "user_admin_id" uuid NOT NULL,
  "role" character varying(20) NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_kanban_project_members_user" FOREIGN KEY ("user_admin_id") REFERENCES "user_admins" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_kanban_projects_members" FOREIGN KEY ("project_id") REFERENCES "kanban_projects" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_kanban_member_unique" to table: "kanban_project_members"
CREATE UNIQUE INDEX "idx_kanban_member_unique" ON "kanban_project_members" ("project_id", "user_admin_id");
-- Create index "idx_kanban_project_members_deleted_at" to table: "kanban_project_members"
CREATE INDEX "idx_kanban_project_members_deleted_at" ON "kanban_project_members" ("deleted_at");
-- Create index "idx_kanban_project_members_project_id" to table: "kanban_project_members"
CREATE INDEX "idx_kanban_project_members_project_id" ON "kanban_project_members" ("project_id");
-- Create index "idx_kanban_project_members_user_admin_id" to table: "kanban_project_members"
CREATE INDEX "idx_kanban_project_members_user_admin_id" ON "kanban_project_members" ("user_admin_id");

-- Permiso master para proyectos en user_admins
ALTER TABLE "user_admins" ADD COLUMN IF NOT EXISTS "can_manage_projects" boolean NULL DEFAULT false;
