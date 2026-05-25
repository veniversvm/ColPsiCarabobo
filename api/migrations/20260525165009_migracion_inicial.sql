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
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
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
-- Create "active_sessions" table
CREATE TABLE "active_sessions" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
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
  PRIMARY KEY ("id")
);
-- Create index "idx_psi_deontologia_deleted_at" to table: "psi_deontologia"
CREATE INDEX "idx_psi_deontologia_deleted_at" ON "psi_deontologia" ("deleted_at");
-- Create index "idx_psi_deontologia_psi_user_id" to table: "psi_deontologia"
CREATE INDEX "idx_psi_deontologia_psi_user_id" ON "psi_deontologia" ("psi_user_id");
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
  PRIMARY KEY ("id")
);
-- Create index "idx_psi_observations_deleted_at" to table: "psi_observations"
CREATE INDEX "idx_psi_observations_deleted_at" ON "psi_observations" ("deleted_at");
-- Create index "idx_psi_observations_psi_user_id" to table: "psi_observations"
CREATE INDEX "idx_psi_observations_psi_user_id" ON "psi_observations" ("psi_user_id");
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
  "is_active" boolean NULL DEFAULT true,
  "sudo" boolean NULL DEFAULT false,
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
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_user_admins_email" UNIQUE ("email"),
  CONSTRAINT "uni_user_admins_username" UNIQUE ("username")
);
-- Create index "idx_user_admins_deleted_at" to table: "user_admins"
CREATE INDEX "idx_user_admins_deleted_at" ON "user_admins" ("deleted_at");
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
  "is_active" boolean NULL,
  "first_name" character varying(255) NOT NULL,
  "second_name" character varying(255) NULL,
  "last_name" character varying(255) NOT NULL,
  "second_last_name" character varying(255) NULL,
  "fpv" bigint NOT NULL,
  "ci" bigint NOT NULL,
  "nationality" character varying(1) NOT NULL,
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
  "municipality_carabobo" character varying(255) NULL,
  "show_municipality_carabobo" boolean NULL,
  "phone_carabobo" text NULL DEFAULT 'false',
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
  "mini_bio" text NULL,
  "bio_text_id" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_psi_users_email" UNIQUE ("email"),
  CONSTRAINT "uni_psi_users_username" UNIQUE ("username"),
  CONSTRAINT "fk_psi_users_full_bio" FOREIGN KEY ("bio_text_id") REFERENCES "text_models" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_psi_users_ci" to table: "psi_users"
CREATE UNIQUE INDEX "idx_psi_users_ci" ON "psi_users" ("ci");
-- Create index "idx_psi_users_deleted_at" to table: "psi_users"
CREATE INDEX "idx_psi_users_deleted_at" ON "psi_users" ("deleted_at");
-- Create index "idx_psi_users_fpv" to table: "psi_users"
CREATE UNIQUE INDEX "idx_psi_users_fpv" ON "psi_users" ("fpv");
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
  "graduation_year" character varying(50) NULL,
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
