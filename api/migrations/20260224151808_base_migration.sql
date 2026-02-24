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
-- Create "text_models" table
CREATE TABLE "text_models" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
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
-- Create "user_admins" table
CREATE TABLE "user_admins" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
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
  "is_active" boolean NULL DEFAULT true,
  "key" character varying(512) NULL,
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
-- Create "posts" table
CREATE TABLE "posts" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
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
  "is_active" boolean NULL DEFAULT true,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_posts_text" FOREIGN KEY ("text_id") REFERENCES "text_models" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_posts_deleted_at" to table: "posts"
CREATE INDEX "idx_posts_deleted_at" ON "posts" ("deleted_at");
-- Create "psi_users" table
CREATE TABLE "psi_users" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
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
  "first_name" character varying(255) NOT NULL,
  "second_name" character varying(255) NULL,
  "last_name" character varying(255) NOT NULL,
  "second_last_name" character varying(255) NULL,
  "fpv" bigint NOT NULL,
  "ci" bigint NOT NULL,
  "nationality" character varying(1) NOT NULL,
  "born_date" date NOT NULL,
  "genre" character varying(1) NOT NULL,
  "contact_email" character varying(255) NOT NULL,
  "show_contact_email" boolean NULL DEFAULT false,
  "public_phone" character varying(20) NULL,
  "show_public_phone" boolean NULL DEFAULT false,
  "service_address" character varying(255) NULL,
  "show_public_service_address" boolean NULL DEFAULT false,
  "solvent" boolean NULL DEFAULT false,
  "proof_of_life" boolean NULL DEFAULT false,
  "profile_picture_s3_key" character varying(512) NULL,
  "municipality_carabobo" character varying(255) NULL,
  "phone_carabobo" character varying(20) NULL,
  "cel_phone_carabobo" character varying(20) NULL,
  "state_outside" character varying(255) NULL,
  "municipality_out_side_carabobo" character varying(255) NULL,
  "phone_out_side_carabobo" character varying(20) NULL,
  "cel_phone_out_side_carabobo" character varying(20) NULL,
  "primary_specialty" character varying(50) NULL,
  "secondary_specialty" character varying(50) NULL,
  "mini_bio" text NULL,
  "bio_text_id" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_psi_users_email" UNIQUE ("email"),
  CONSTRAINT "uni_psi_users_username" UNIQUE ("username")
);
-- Create index "idx_psi_users_ci" to table: "psi_users"
CREATE UNIQUE INDEX "idx_psi_users_ci" ON "psi_users" ("ci");
-- Create index "idx_psi_users_deleted_at" to table: "psi_users"
CREATE INDEX "idx_psi_users_deleted_at" ON "psi_users" ("deleted_at");
-- Create index "idx_psi_users_fpv" to table: "psi_users"
CREATE UNIQUE INDEX "idx_psi_users_fpv" ON "psi_users" ("fpv");
-- Create "psi_user_col_data" table
CREATE TABLE "psi_user_col_data" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "create_by" character varying(255) NULL,
  "update_by" character varying(255) NULL,
  "create_by_id" uuid NULL,
  "update_by_id" uuid NULL,
  "psi_user_model_id" uuid NULL,
  "university_undergraduate" character varying(255) NULL,
  "show_university_undergraduate" boolean NULL DEFAULT false,
  "graduate_date" date NULL,
  "show_graduate_date" boolean NULL DEFAULT false,
  "mention_undergraduate" character varying(255) NULL,
  "show_mention_undergraduate" boolean NULL DEFAULT false,
  "register_title_state" character varying(255) NULL,
  "register_title_date" date NULL,
  "register_number" bigint NULL,
  "register_folio" character varying(255) NULL,
  "register_tome" character varying(255) NULL,
  "guild_director" boolean NULL DEFAULT false,
  "sixty_five_or_plus" boolean NULL DEFAULT false,
  "guild_collaborator" boolean NULL DEFAULT false,
  "public_employee" boolean NULL DEFAULT false,
  "university_professor" boolean NULL DEFAULT false,
  "date_of_last_solvency" date NULL,
  "double_guild" boolean NULL DEFAULT false,
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
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "create_by" character varying(255) NULL,
  "update_by" character varying(255) NULL,
  "create_by_id" uuid NULL,
  "update_by_id" uuid NULL,
  "psi_user_id" uuid NULL,
  "title" character varying(255) NOT NULL,
  "university" character varying(255) NOT NULL,
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
  "id" uuid NOT NULL,
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
-- Create index "idx_psi_user_social_networks_id" to table: "psi_user_social_networks"
CREATE INDEX "idx_psi_user_social_networks_id" ON "psi_user_social_networks" ("id");
-- Create index "idx_psi_user_social_networks_psi_user_id" to table: "psi_user_social_networks"
CREATE INDEX "idx_psi_user_social_networks_psi_user_id" ON "psi_user_social_networks" ("psi_user_id");
