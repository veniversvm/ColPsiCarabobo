-- Create "page_views" table
CREATE TABLE "page_views" (
  "id" bigserial NOT NULL,
  "path" character varying(512) NULL,
  "method" character varying(10) NULL,
  "user_id" uuid NULL,
  "session_id" character varying(64) NULL,
  "ip" character varying(45) NULL,
  "referer" character varying(512) NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_page_views_created_at" to table: "page_views"
CREATE INDEX "idx_page_views_created_at" ON "page_views" ("created_at");
-- Create index "idx_page_views_path" to table: "page_views"
CREATE INDEX "idx_page_views_path" ON "page_views" ("path");
-- Create index "idx_page_views_session_id" to table: "page_views"
CREATE INDEX "idx_page_views_session_id" ON "page_views" ("session_id");
-- Create index "idx_page_views_user_id" to table: "page_views"
CREATE INDEX "idx_page_views_user_id" ON "page_views" ("user_id");
-- Create "admin_permission_logs" table
CREATE TABLE "admin_permission_logs" (
  "id" uuid NOT NULL,
  "target_admin_id" uuid NULL,
  "target_username" character varying(100) NULL,
  "action" character varying(50) NULL,
  "changed_by_id" uuid NULL,
  "changed_by_username" character varying(100) NULL,
  "permissions_changed" text NULL,
  "role_from" character varying(50) NULL,
  "role_to" character varying(50) NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_admin_permission_logs_changed_by_id" to table: "admin_permission_logs"
CREATE INDEX "idx_admin_permission_logs_changed_by_id" ON "admin_permission_logs" ("changed_by_id");
-- Create index "idx_admin_permission_logs_target_admin_id" to table: "admin_permission_logs"
CREATE INDEX "idx_admin_permission_logs_target_admin_id" ON "admin_permission_logs" ("target_admin_id");
-- Create "app_settings" table
CREATE TABLE "app_settings" (
  "key" character varying(80) NOT NULL,
  "value" jsonb NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("key")
);
-- Create "profile_views" table
CREATE TABLE "profile_views" (
  "id" bigserial NOT NULL,
  "psi_id" uuid NOT NULL,
  "viewer_id" uuid NULL,
  "session_id" character varying(64) NULL,
  "ip" character varying(45) NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_profile_views_created_at" to table: "profile_views"
CREATE INDEX "idx_profile_views_created_at" ON "profile_views" ("created_at");
-- Create index "idx_profile_views_psi_id" to table: "profile_views"
CREATE INDEX "idx_profile_views_psi_id" ON "profile_views" ("psi_id");
-- Create index "idx_profile_views_viewer_id" to table: "profile_views"
CREATE INDEX "idx_profile_views_viewer_id" ON "profile_views" ("viewer_id");
-- Create "psi_specialty_models" table
CREATE TABLE "psi_specialty_models" (
  "id" bigserial NOT NULL,
  "name" character varying(100) NOT NULL,
  "description" text NULL,
  "active" boolean NOT NULL DEFAULT true,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "create_by" character varying(255) NULL,
  "update_by" character varying(255) NULL,
  "create_by_id" uuid NULL,
  "update_by_id" uuid NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_psi_specialty_models_deleted_at" to table: "psi_specialty_models"
CREATE INDEX "idx_psi_specialty_models_deleted_at" ON "psi_specialty_models" ("deleted_at");
-- Create index "idx_psi_specialty_models_name" to table: "psi_specialty_models"
CREATE UNIQUE INDEX "idx_psi_specialty_models_name" ON "psi_specialty_models" ("name");
-- Create "search_events" table
CREATE TABLE "search_events" (
  "id" bigserial NOT NULL,
  "query" character varying(255) NULL,
  "specialty" character varying(255) NULL,
  "municipality" character varying(255) NULL,
  "state" character varying(255) NULL,
  "results_count" bigint NULL,
  "user_id" uuid NULL,
  "session_id" character varying(64) NULL,
  "ip" character varying(45) NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_search_events_created_at" to table: "search_events"
CREATE INDEX "idx_search_events_created_at" ON "search_events" ("created_at");
-- Create index "idx_search_events_municipality" to table: "search_events"
CREATE INDEX "idx_search_events_municipality" ON "search_events" ("municipality");
-- Create index "idx_search_events_specialty" to table: "search_events"
CREATE INDEX "idx_search_events_specialty" ON "search_events" ("specialty");
-- Create index "idx_search_events_state" to table: "search_events"
CREATE INDEX "idx_search_events_state" ON "search_events" ("state");
-- Create index "idx_search_events_user_id" to table: "search_events"
CREATE INDEX "idx_search_events_user_id" ON "search_events" ("user_id");
-- Create "login_events" table
CREATE TABLE "login_events" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "user_id" uuid NOT NULL,
  "username" character varying(100) NULL,
  "role" character varying(50) NULL,
  "ip" character varying(45) NULL,
  "user_agent" character varying(512) NULL,
  "created_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_login_events_created_at" to table: "login_events"
CREATE INDEX "idx_login_events_created_at" ON "login_events" ("created_at");
-- Create index "idx_login_events_deleted_at" to table: "login_events"
CREATE INDEX "idx_login_events_deleted_at" ON "login_events" ("deleted_at");
-- Create index "idx_login_events_user_id" to table: "login_events"
CREATE INDEX "idx_login_events_user_id" ON "login_events" ("user_id");
-- Create "active_sessions" table
CREATE TABLE "active_sessions" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "user_id" uuid NOT NULL,
  "username" character varying(100) NULL,
  "role" character varying(50) NULL,
  "ip" character varying(45) NULL,
  "last_seen" timestamptz NULL,
  "expires_at" timestamptz NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_active_sessions_expires_at" to table: "active_sessions"
CREATE INDEX "idx_active_sessions_expires_at" ON "active_sessions" ("expires_at");
-- Create index "idx_active_sessions_last_seen" to table: "active_sessions"
CREATE INDEX "idx_active_sessions_last_seen" ON "active_sessions" ("last_seen");
-- Create index "idx_active_sessions_user_id" to table: "active_sessions"
CREATE UNIQUE INDEX "idx_active_sessions_user_id" ON "active_sessions" ("user_id", "user_id");
-- Create "user_admins" table
CREATE TABLE "user_admins" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "create_by" character varying(255) NULL,
  "update_by" character varying(255) NULL,
  "create_by_id" uuid NULL,
  "update_by_id" uuid NULL,
  "username" character varying(25) NOT NULL,
  "email" character varying(255) NOT NULL,
  "password" character varying(512) NOT NULL,
  "key" character varying(512) NULL,
  "is_active" boolean NULL DEFAULT false,
  "must_change_password" boolean NULL DEFAULT false,
  "sudo" boolean NULL DEFAULT false,
  "role" character varying(50) NULL,
  "can_read_psi" boolean NULL DEFAULT false,
  "can_create_psi" boolean NULL DEFAULT false,
  "can_update_psi" boolean NULL DEFAULT false,
  "can_delete_psi" boolean NULL DEFAULT false,
  "can_create_admin" boolean NULL DEFAULT false,
  "can_update_admin" boolean NULL DEFAULT false,
  "can_delete_admin" boolean NULL DEFAULT false,
  "can_publish" boolean NULL DEFAULT false,
  "can_update_publish" boolean NULL DEFAULT false,
  "can_delete_publish" boolean NULL DEFAULT false,
  "can_send_notifications" boolean NULL DEFAULT false,
  "can_manage_notifications" boolean NULL DEFAULT false,
  "can_read_notifications" boolean NULL DEFAULT false,
  "can_create_tags" boolean NULL DEFAULT false,
  "can_edit_tags" boolean NULL DEFAULT false,
  "can_delete_tags" boolean NULL DEFAULT false,
  "can_manage_projects" boolean NULL DEFAULT false,
  "can_manage_tickets" boolean NULL DEFAULT false,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_user_admins_email" UNIQUE ("email"),
  CONSTRAINT "uni_user_admins_username" UNIQUE ("username")
);
-- Create index "idx_user_admins_deleted_at" to table: "user_admins"
CREATE INDEX "idx_user_admins_deleted_at" ON "user_admins" ("deleted_at");
-- Create "settings_audit_logs" table
CREATE TABLE "settings_audit_logs" (
  "id" uuid NOT NULL,
  "changed_by_id" uuid NULL,
  "changed_by_username" character varying(100) NULL,
  "key" character varying(80) NULL,
  "enabled_from" boolean NULL,
  "enabled_to" boolean NULL,
  "message_from" character varying(500) NULL,
  "message_to" character varying(500) NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_settings_audit_logs_changed_by_id" to table: "settings_audit_logs"
CREATE INDEX "idx_settings_audit_logs_changed_by_id" ON "settings_audit_logs" ("changed_by_id");
-- Create index "idx_settings_audit_logs_key" to table: "settings_audit_logs"
CREATE INDEX "idx_settings_audit_logs_key" ON "settings_audit_logs" ("key");
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
-- Create "notifications" table
CREATE TABLE "notifications" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "create_by" character varying(255) NULL,
  "update_by" character varying(255) NULL,
  "create_by_id" uuid NULL,
  "update_by_id" uuid NULL,
  "title" character varying(255) NOT NULL,
  "message" text NOT NULL,
  "target_type" character varying(20) NOT NULL,
  "sender_id" uuid NOT NULL,
  "send_email" boolean NULL DEFAULT false,
  "scheduled_at" timestamptz NULL,
  "sent_at" timestamptz NULL,
  "status" character varying(20) NOT NULL DEFAULT 'pending',
  PRIMARY KEY ("id")
);
-- Create index "idx_notifications_deleted_at" to table: "notifications"
CREATE INDEX "idx_notifications_deleted_at" ON "notifications" ("deleted_at");
-- Create "notification_attachments" table
CREATE TABLE "notification_attachments" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "create_by" character varying(255) NULL,
  "update_by" character varying(255) NULL,
  "create_by_id" uuid NULL,
  "update_by_id" uuid NULL,
  "notification_id" uuid NOT NULL,
  "name" character varying(255) NULL,
  "s3_key" character varying(512) NOT NULL,
  "content_type" character varying(100) NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_notifications_attachs" FOREIGN KEY ("notification_id") REFERENCES "notifications" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_notification_attachments_deleted_at" to table: "notification_attachments"
CREATE INDEX "idx_notification_attachments_deleted_at" ON "notification_attachments" ("deleted_at");
-- Create index "idx_notification_attachments_notification_id" to table: "notification_attachments"
CREATE INDEX "idx_notification_attachments_notification_id" ON "notification_attachments" ("notification_id");
-- Create "notification_filters" table
CREATE TABLE "notification_filters" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "notification_id" uuid NOT NULL,
  "municipality" character varying(255) NULL,
  "state" character varying(255) NULL,
  "genre" character varying(1) NULL,
  "specialty_id" bigint NULL,
  "solvent" boolean NULL,
  "target_user_ids" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_notifications_filters" FOREIGN KEY ("notification_id") REFERENCES "notifications" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_notification_filters_notification_id" to table: "notification_filters"
CREATE INDEX "idx_notification_filters_notification_id" ON "notification_filters" ("notification_id");
-- Create "text_models" table
CREATE TABLE "text_models" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "create_by" character varying(255) NULL,
  "update_by" character varying(255) NULL,
  "create_by_id" uuid NULL,
  "update_by_id" uuid NULL,
  "content" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_text_models_deleted_at" to table: "text_models"
CREATE INDEX "idx_text_models_deleted_at" ON "text_models" ("deleted_at");
-- Create "psi_users" table
CREATE TABLE "psi_users" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "create_by" character varying(255) NULL,
  "update_by" character varying(255) NULL,
  "create_by_id" uuid NULL,
  "update_by_id" uuid NULL,
  "username" character varying(25) NOT NULL,
  "email" character varying(255) NOT NULL,
  "password" character varying(512) NOT NULL,
  "key" character varying(512) NULL,
  "is_active" boolean NULL DEFAULT false,
  "must_change_password" boolean NULL DEFAULT false,
  "audio_book_shell_id" character varying(50) NULL,
  "first_name" character varying(255) NOT NULL,
  "second_name" character varying(255) NULL,
  "last_name" character varying(255) NOT NULL,
  "second_last_name" character varying(255) NULL,
  "fpv" bigint NOT NULL,
  "ci" bigint NOT NULL,
  "nationality" character varying(1) NOT NULL,
  "control_number" character varying(50) NULL,
  "born_date" date NOT NULL,
  "genre" character varying(1) NOT NULL,
  "solvent" boolean NULL DEFAULT false,
  "proof_of_life" boolean NULL,
  "profile_picture_s3_key" character varying(512) NULL,
  "contact_phone" character varying(255) NOT NULL,
  "contact_cell_phone" character varying(255) NOT NULL,
  "contact_email" character varying(255) NOT NULL,
  "show_contact_email" boolean NULL DEFAULT false,
  "service_address" character varying(255) NULL,
  "show_public_service_address" boolean NULL DEFAULT false,
  "service_modality_presencial" boolean NULL DEFAULT false,
  "service_modality_distance" boolean NULL DEFAULT false,
  "service_modality_telephone" boolean NULL DEFAULT false,
  "show_service_modality" boolean NULL DEFAULT false,
  "municipality_carabobo" character varying(255) NULL,
  "show_municipality_carabobo" boolean NULL DEFAULT false,
  "phone_carabobo" character varying(20) NULL DEFAULT '',
  "show_phone_carabobo" boolean NULL DEFAULT false,
  "cel_phone_carabobo" character varying(20) NULL,
  "show_cel_phone_carabobo" boolean NULL DEFAULT false,
  "state_outside" character varying(255) NULL,
  "show_state_outside" boolean NULL DEFAULT false,
  "municipality_out_side_carabobo" character varying(255) NULL,
  "show_municipality_out_side_carabobo" boolean NULL DEFAULT false,
  "phone_out_side_carabobo" character varying(20) NULL,
  "show_phone_out_side_carabobo" boolean NULL DEFAULT false,
  "cel_phone_out_side_carabobo" character varying(20) NULL,
  "show_cell_phone_out_side_carabobo" boolean NULL DEFAULT false,
  "service_address_out_side_carabobo" character varying(255) NULL,
  "show_public_service_address_out_side_carabobo" boolean NULL DEFAULT false,
  "country" character varying(255) NULL,
  "phone_out_side_venezuela" character varying(20) NULL,
  "show_phone_out_side_venezuela" boolean NULL DEFAULT false,
  "cell_phone_out_side_venezuela" character varying(20) NULL,
  "show_cell_phone_out_side_venezuela" boolean NULL DEFAULT false,
  "service_address_out_side_venezuela" character varying(255) NULL,
  "show_public_service_address_out_side_venezuela" boolean NULL DEFAULT false,
  "primary_work_area" character varying(50) NULL,
  "secondary_work_area" character varying(50) NULL,
  "primary_specialty_id" bigint NULL,
  "secondary_specialty_id" bigint NULL,
  "mini_bio" text NULL,
  "bio_text_id" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_psi_users_audio_book_shell_id" UNIQUE ("audio_book_shell_id"),
  CONSTRAINT "uni_psi_users_email" UNIQUE ("email"),
  CONSTRAINT "uni_psi_users_username" UNIQUE ("username"),
  CONSTRAINT "fk_psi_users_full_bio" FOREIGN KEY ("bio_text_id") REFERENCES "text_models" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_psi_users_ci" to table: "psi_users"
CREATE UNIQUE INDEX "idx_psi_users_ci" ON "psi_users" ("ci");
-- Create index "idx_psi_users_control_number" to table: "psi_users"
CREATE UNIQUE INDEX "idx_psi_users_control_number" ON "psi_users" ("control_number") WHERE (((control_number)::text <> ''::text) AND (deleted_at IS NULL));
-- Create index "idx_psi_users_deleted_at" to table: "psi_users"
CREATE INDEX "idx_psi_users_deleted_at" ON "psi_users" ("deleted_at");
-- Create index "idx_psi_users_fpv" to table: "psi_users"
CREATE UNIQUE INDEX "idx_psi_users_fpv" ON "psi_users" ("fpv");
-- Create "notification_targets" table
CREATE TABLE "notification_targets" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "notification_id" uuid NOT NULL,
  "psi_user_id" uuid NOT NULL,
  "is_read" boolean NULL DEFAULT false,
  "read_at" timestamptz NULL,
  "email_sent" boolean NULL DEFAULT false,
  "email_sent_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_notification_targets_psi_user" FOREIGN KEY ("psi_user_id") REFERENCES "psi_users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_notifications_targets" FOREIGN KEY ("notification_id") REFERENCES "notifications" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_notification_targets_notification_psi" to table: "notification_targets"
CREATE UNIQUE INDEX "idx_notification_targets_notification_psi" ON "notification_targets" ("notification_id", "psi_user_id");
-- Create index "idx_notification_targets_psi_read" to table: "notification_targets"
CREATE INDEX "idx_notification_targets_psi_read" ON "notification_targets" ("psi_user_id", "is_read");
-- Create "posts" table
CREATE TABLE "posts" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "create_by" character varying(255) NULL,
  "update_by" character varying(255) NULL,
  "create_by_id" uuid NULL,
  "update_by_id" uuid NULL,
  "title" character varying(100) NOT NULL,
  "short_description" character varying(250) NULL,
  "type" character varying(20) NULL DEFAULT 'public',
  "text_id" uuid NULL,
  "image_s3_key" character varying(512) NULL,
  "status" character varying(20) NOT NULL DEFAULT 'draft',
  "publish_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_posts_text" FOREIGN KEY ("text_id") REFERENCES "text_models" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_posts_deleted_at" to table: "posts"
CREATE INDEX "idx_posts_deleted_at" ON "posts" ("deleted_at");
-- Create "psi_deontologia" table
CREATE TABLE "psi_deontologia" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "create_by" character varying(255) NULL,
  "update_by" character varying(255) NULL,
  "create_by_id" uuid NULL,
  "update_by_id" uuid NULL,
  "psi_user_id" uuid NULL,
  "content" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_psi_users_deontologia" FOREIGN KEY ("psi_user_id") REFERENCES "psi_users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_psi_deontologia_deleted_at" to table: "psi_deontologia"
CREATE INDEX "idx_psi_deontologia_deleted_at" ON "psi_deontologia" ("deleted_at");
-- Create index "idx_psi_deontologia_psi_user_id" to table: "psi_deontologia"
CREATE INDEX "idx_psi_deontologia_psi_user_id" ON "psi_deontologia" ("psi_user_id");
-- Create "psi_inscription_requests" table
CREATE TABLE "psi_inscription_requests" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "create_by" character varying(255) NULL,
  "update_by" character varying(255) NULL,
  "create_by_id" uuid NULL,
  "update_by_id" uuid NULL,
  "cedula" bigint NOT NULL,
  "nacionalidad" character varying(1) NOT NULL,
  "nombres" character varying(255) NOT NULL,
  "apellidos" character varying(255) NOT NULL,
  "segundo_nombre" character varying(255) NULL,
  "segundo_apellido" character varying(255) NULL,
  "genero" character varying(1) NULL,
  "fpv" bigint NULL,
  "telefono" character varying(50) NULL,
  "correo" character varying(255) NOT NULL,
  "fecha_nacimiento" date NULL,
  "titulo_universidad" character varying(255) NULL,
  "titulo_fecha_graduacion" date NULL,
  "titulo_mencion" character varying(255) NULL,
  "titulo_registro_numero" character varying(100) NULL,
  "titulo_registro_estado" character varying(100) NULL,
  "titulo_registro_tomo" character varying(100) NULL,
  "titulo_registro_folio" character varying(100) NULL,
  "rif" character varying(50) NULL,
  "service_address" character varying(255) NULL,
  "municipality_carabobo" character varying(255) NULL,
  "state_outside" character varying(255) NULL,
  "municipality_out_side_carabobo" character varying(255) NULL,
  "country" character varying(255) NULL,
  "service_modality_presencial" boolean NULL DEFAULT false,
  "service_modality_distance" boolean NULL DEFAULT false,
  "service_modality_telephone" boolean NULL DEFAULT false,
  "primary_specialty_id" bigint NULL,
  "secondary_specialty_id" bigint NULL,
  "foto_s3_key" character varying(512) NULL,
  "comprobante_s3_key" character varying(512) NULL,
  "status" character varying(20) NULL DEFAULT 'pending',
  "control_number" character varying(50) NULL,
  "notes" text NULL,
  "psi_user_id" uuid NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_inscription_requests_cedula_pending" to table: "psi_inscription_requests"
CREATE UNIQUE INDEX "idx_inscription_requests_cedula_pending" ON "psi_inscription_requests" ("cedula") WHERE ((status)::text = 'pending'::text);
-- Create index "idx_inscription_requests_control_number" to table: "psi_inscription_requests"
CREATE UNIQUE INDEX "idx_inscription_requests_control_number" ON "psi_inscription_requests" ("control_number") WHERE (((control_number)::text <> ''::text) AND (control_number IS NOT NULL));
-- Create index "idx_inscription_requests_psi_user_id" to table: "psi_inscription_requests"
CREATE INDEX "idx_inscription_requests_psi_user_id" ON "psi_inscription_requests" ("psi_user_id");
-- Create index "idx_inscription_requests_status" to table: "psi_inscription_requests"
CREATE INDEX "idx_inscription_requests_status" ON "psi_inscription_requests" ("status");
-- Create index "idx_psi_inscription_requests_deleted_at" to table: "psi_inscription_requests"
CREATE INDEX "idx_psi_inscription_requests_deleted_at" ON "psi_inscription_requests" ("deleted_at");
-- Create "psi_inscription_documents" table
CREATE TABLE "psi_inscription_documents" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "create_by" character varying(255) NULL,
  "update_by" character varying(255) NULL,
  "create_by_id" uuid NULL,
  "update_by_id" uuid NULL,
  "inscription_request_id" uuid NOT NULL,
  "document_type" character varying(50) NOT NULL,
  "s3_key" character varying(512) NOT NULL,
  "title" character varying(255) NULL,
  "notes" text NULL,
  "original_filename" character varying(255) NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_psi_inscription_requests_documents" FOREIGN KEY ("inscription_request_id") REFERENCES "psi_inscription_requests" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_inscription_document_type" to table: "psi_inscription_documents"
CREATE UNIQUE INDEX "idx_inscription_document_type" ON "psi_inscription_documents" ("inscription_request_id", "document_type");
-- Create index "idx_psi_inscription_documents_deleted_at" to table: "psi_inscription_documents"
CREATE INDEX "idx_psi_inscription_documents_deleted_at" ON "psi_inscription_documents" ("deleted_at");
-- Create "psi_observations" table
CREATE TABLE "psi_observations" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "create_by" character varying(255) NULL,
  "update_by" character varying(255) NULL,
  "create_by_id" uuid NULL,
  "update_by_id" uuid NULL,
  "psi_user_id" uuid NULL,
  "content" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_psi_users_observations" FOREIGN KEY ("psi_user_id") REFERENCES "psi_users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_psi_observations_deleted_at" to table: "psi_observations"
CREATE INDEX "idx_psi_observations_deleted_at" ON "psi_observations" ("deleted_at");
-- Create index "idx_psi_observations_psi_user_id" to table: "psi_observations"
CREATE INDEX "idx_psi_observations_psi_user_id" ON "psi_observations" ("psi_user_id");
-- Create "psi_user_col_data" table
CREATE TABLE "psi_user_col_data" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "create_by" character varying(255) NULL,
  "update_by" character varying(255) NULL,
  "create_by_id" uuid NULL,
  "update_by_id" uuid NULL,
  "psi_user_model_id" uuid NULL,
  "guild_inscription_date" date NULL,
  "university_undergraduate" character varying(255) NULL,
  "show_university_undergraduate" boolean NULL DEFAULT false,
  "graduate_date" date NULL,
  "show_graduate_date" boolean NULL DEFAULT false,
  "mention_undergraduate" character varying(255) NULL,
  "show_mention_undergraduate" boolean NULL DEFAULT false,
  "title_image_one_s3_key" character varying(512) NULL,
  "title_image_two_s3_key" character varying(512) NULL,
  "title_image_three_s3_key" character varying(512) NULL,
  "register_title_state" character varying(255) NULL,
  "register_title_date" date NULL,
  "register_number" bigint NULL,
  "register_folio" character varying(255) NULL,
  "register_tome" character varying(255) NULL,
  "guild_director" boolean NULL DEFAULT false,
  "sixty_five_or_plus" boolean NULL DEFAULT false,
  "guild_collaborator" boolean NULL DEFAULT false,
  "public_employee" boolean NULL DEFAULT false,
  "discapacity" boolean NULL DEFAULT false,
  "university_professor" boolean NULL DEFAULT false,
  "birthday_notification" boolean NULL DEFAULT false,
  "date_of_last_solvency" date NULL,
  "double_guild" boolean NULL DEFAULT false,
  "double_guild_location" character varying(255) NULL,
  "cpsm" boolean NULL DEFAULT false,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_psi_users_col_data" FOREIGN KEY ("psi_user_model_id") REFERENCES "psi_users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_psi_user_col_data_deleted_at" to table: "psi_user_col_data"
CREATE INDEX "idx_psi_user_col_data_deleted_at" ON "psi_user_col_data" ("deleted_at");
-- Create index "idx_psi_user_col_data_psi_user_model_id" to table: "psi_user_col_data"
CREATE UNIQUE INDEX "idx_psi_user_col_data_psi_user_model_id" ON "psi_user_col_data" ("psi_user_model_id");
-- Create "psi_user_documents" table
CREATE TABLE "psi_user_documents" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "create_by" character varying(255) NULL,
  "update_by" character varying(255) NULL,
  "create_by_id" uuid NULL,
  "update_by_id" uuid NULL,
  "psi_user_id" uuid NULL,
  "document_type" character varying(50) NOT NULL DEFAULT 'otro',
  "title" character varying(255) NOT NULL,
  "notes" text NULL,
  "document_date" date NULL,
  "s3_key" character varying(512) NOT NULL,
  "filename" character varying(255) NULL,
  "mime_type" character varying(100) NULL,
  "size_bytes" bigint NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_psi_users_documents" FOREIGN KEY ("psi_user_id") REFERENCES "psi_users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_psi_user_documents_deleted_at" to table: "psi_user_documents"
CREATE INDEX "idx_psi_user_documents_deleted_at" ON "psi_user_documents" ("deleted_at");
-- Create index "idx_psi_user_documents_psi_user_id" to table: "psi_user_documents"
CREATE INDEX "idx_psi_user_documents_psi_user_id" ON "psi_user_documents" ("psi_user_id");
-- Create "psi_user_post_grades" table
CREATE TABLE "psi_user_post_grades" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "create_by" character varying(255) NULL,
  "update_by" character varying(255) NULL,
  "create_by_id" uuid NULL,
  "update_by_id" uuid NULL,
  "psi_user_id" uuid NULL,
  "type" character varying(50) NOT NULL,
  "title" character varying(255) NOT NULL,
  "university" character varying(255) NULL,
  "graduation_year" bigint NULL,
  "description" text NULL,
  "active" boolean NULL DEFAULT true,
  "pic_one_s3_key" character varying(512) NULL,
  "pic_two_s3_key" character varying(512) NULL,
  "pic_three_s3_key" character varying(512) NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_psi_users_post_grades" FOREIGN KEY ("psi_user_id") REFERENCES "psi_users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_psi_user_post_grades_deleted_at" to table: "psi_user_post_grades"
CREATE INDEX "idx_psi_user_post_grades_deleted_at" ON "psi_user_post_grades" ("deleted_at");
-- Create index "idx_psi_user_post_grades_psi_user_id" to table: "psi_user_post_grades"
CREATE INDEX "idx_psi_user_post_grades_psi_user_id" ON "psi_user_post_grades" ("psi_user_id");
-- Create "psi_user_social_networks" table
CREATE TABLE "psi_user_social_networks" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "create_by" character varying(255) NULL,
  "update_by" character varying(255) NULL,
  "create_by_id" uuid NULL,
  "update_by_id" uuid NULL,
  "psi_user_id" uuid NOT NULL,
  "name" character varying(50) NOT NULL,
  "url" character varying(512) NOT NULL,
  "is_active" boolean NOT NULL DEFAULT true,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_psi_users_social_networks" FOREIGN KEY ("psi_user_id") REFERENCES "psi_users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_psi_user_social_networks_deleted_at" to table: "psi_user_social_networks"
CREATE INDEX "idx_psi_user_social_networks_deleted_at" ON "psi_user_social_networks" ("deleted_at");
-- Create index "idx_psi_user_social_networks_psi_user_id" to table: "psi_user_social_networks"
CREATE INDEX "idx_psi_user_social_networks_psi_user_id" ON "psi_user_social_networks" ("psi_user_id");
-- Create "psi_user_solvency" table
CREATE TABLE "psi_user_solvency" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "create_by" character varying(255) NULL,
  "update_by" character varying(255) NULL,
  "create_by_id" uuid NULL,
  "update_by_id" uuid NULL,
  "psi_user_model_id" uuid NOT NULL,
  "date" date NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_psi_users_solvencies" FOREIGN KEY ("psi_user_model_id") REFERENCES "psi_users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_psi_solvency_unique" to table: "psi_user_solvency"
CREATE UNIQUE INDEX "idx_psi_solvency_unique" ON "psi_user_solvency" ("psi_user_model_id", "date");
-- Create index "idx_psi_user_solvency_deleted_at" to table: "psi_user_solvency"
CREATE INDEX "idx_psi_user_solvency_deleted_at" ON "psi_user_solvency" ("deleted_at");
-- Create "ticket_motivos" table
CREATE TABLE "ticket_motivos" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "create_by" character varying(255) NULL,
  "update_by" character varying(255) NULL,
  "create_by_id" uuid NULL,
  "update_by_id" uuid NULL,
  "name" character varying(120) NOT NULL,
  "description" character varying(500) NULL,
  "tickets_per_psi" bigint NOT NULL DEFAULT 3,
  PRIMARY KEY ("id")
);
-- Create index "idx_ticket_motivos_deleted_at" to table: "ticket_motivos"
CREATE INDEX "idx_ticket_motivos_deleted_at" ON "ticket_motivos" ("deleted_at");
-- Create "ticket_estados" table
CREATE TABLE "ticket_estados" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "create_by" character varying(255) NULL,
  "update_by" character varying(255) NULL,
  "create_by_id" uuid NULL,
  "update_by_id" uuid NULL,
  "motivo_id" bigint NOT NULL,
  "name" character varying(60) NOT NULL,
  "order" bigint NOT NULL DEFAULT 0,
  "is_closed" boolean NOT NULL DEFAULT false,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_ticket_motivos_estados" FOREIGN KEY ("motivo_id") REFERENCES "ticket_motivos" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_ticket_estados_deleted_at" to table: "ticket_estados"
CREATE INDEX "idx_ticket_estados_deleted_at" ON "ticket_estados" ("deleted_at");
-- Create index "idx_ticket_estados_motivo_id" to table: "ticket_estados"
CREATE INDEX "idx_ticket_estados_motivo_id" ON "ticket_estados" ("motivo_id");
-- Create index "idx_ticket_estados_motivo_nombre" to table: "ticket_estados"
CREATE UNIQUE INDEX "idx_ticket_estados_motivo_nombre" ON "ticket_estados" ("motivo_id", "name");
-- Create "tickets" table
CREATE TABLE "tickets" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "create_by" character varying(255) NULL,
  "update_by" character varying(255) NULL,
  "create_by_id" uuid NULL,
  "update_by_id" uuid NULL,
  "psi_user_id" uuid NOT NULL,
  "motivo_id" bigint NOT NULL,
  "estado_id" bigint NOT NULL,
  "title" character varying(200) NOT NULL,
  "description" text NULL,
  "close_reason" character varying(500) NULL,
  "closed_by_type" character varying(10) NULL,
  "closed_by_admin_id" uuid NULL,
  "closed_by_psi_id" uuid NULL,
  "closed_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_tickets_estado" FOREIGN KEY ("estado_id") REFERENCES "ticket_estados" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_tickets_motivo" FOREIGN KEY ("motivo_id") REFERENCES "ticket_motivos" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_tickets_psi" FOREIGN KEY ("psi_user_id") REFERENCES "psi_users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_tickets_deleted_at" to table: "tickets"
CREATE INDEX "idx_tickets_deleted_at" ON "tickets" ("deleted_at");
-- Create index "idx_tickets_estado_id" to table: "tickets"
CREATE INDEX "idx_tickets_estado_id" ON "tickets" ("estado_id", "id");
-- Create index "idx_tickets_motivo_id" to table: "tickets"
CREATE INDEX "idx_tickets_motivo_id" ON "tickets" ("motivo_id", "id");
-- Create index "idx_tickets_psi_user_id" to table: "tickets"
CREATE INDEX "idx_tickets_psi_user_id" ON "tickets" ("psi_user_id", "id");
-- Create "ticket_mensajes" table
CREATE TABLE "ticket_mensajes" (
  "id" bigserial NOT NULL,
  "ticket_id" bigint NOT NULL,
  "author_type" character varying(10) NOT NULL,
  "author_admin_id" uuid NULL,
  "author_psi_id" uuid NULL,
  "message" text NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_ticket_mensajes_admin" FOREIGN KEY ("author_admin_id") REFERENCES "user_admins" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_ticket_mensajes_psi" FOREIGN KEY ("author_psi_id") REFERENCES "psi_users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_tickets_mensajes" FOREIGN KEY ("ticket_id") REFERENCES "tickets" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_ticket_mensajes_author_admin_id" to table: "ticket_mensajes"
CREATE INDEX "idx_ticket_mensajes_author_admin_id" ON "ticket_mensajes" ("author_admin_id");
-- Create index "idx_ticket_mensajes_author_psi_id" to table: "ticket_mensajes"
CREATE INDEX "idx_ticket_mensajes_author_psi_id" ON "ticket_mensajes" ("author_psi_id");
-- Create index "idx_ticket_mensajes_ticket_id" to table: "ticket_mensajes"
CREATE INDEX "idx_ticket_mensajes_ticket_id" ON "ticket_mensajes" ("ticket_id");
-- Create "ticket_adjuntos" table
CREATE TABLE "ticket_adjuntos" (
  "id" bigserial NOT NULL,
  "mensaje_id" bigint NOT NULL,
  "s3_key" character varying(512) NOT NULL,
  "original_name" character varying(255) NULL,
  "mime_type" character varying(100) NULL,
  "size_bytes" bigint NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_ticket_mensajes_adjuntos" FOREIGN KEY ("mensaje_id") REFERENCES "ticket_mensajes" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_ticket_adjuntos_mensaje_id" to table: "ticket_adjuntos"
CREATE INDEX "idx_ticket_adjuntos_mensaje_id" ON "ticket_adjuntos" ("mensaje_id");
-- Create "ticket_status_logs" table
CREATE TABLE "ticket_status_logs" (
  "id" bigserial NOT NULL,
  "ticket_id" bigint NOT NULL,
  "previous_state_id" bigint NULL,
  "new_state_id" bigint NOT NULL,
  "changed_by_type" character varying(10) NOT NULL,
  "changed_by_admin_id" uuid NULL,
  "changed_by_psi_id" uuid NULL,
  "reason" character varying(500) NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_ticket_status_logs_new_state" FOREIGN KEY ("new_state_id") REFERENCES "ticket_estados" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_ticket_status_logs_previous_state" FOREIGN KEY ("previous_state_id") REFERENCES "ticket_estados" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_tickets_status_logs" FOREIGN KEY ("ticket_id") REFERENCES "tickets" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_ticket_status_logs_new_state_id" to table: "ticket_status_logs"
CREATE INDEX "idx_ticket_status_logs_new_state_id" ON "ticket_status_logs" ("new_state_id");
-- Create index "idx_ticket_status_logs_previous_state_id" to table: "ticket_status_logs"
CREATE INDEX "idx_ticket_status_logs_previous_state_id" ON "ticket_status_logs" ("previous_state_id");
-- Create index "idx_ticket_status_logs_ticket_id" to table: "ticket_status_logs"
CREATE INDEX "idx_ticket_status_logs_ticket_id" ON "ticket_status_logs" ("ticket_id");
