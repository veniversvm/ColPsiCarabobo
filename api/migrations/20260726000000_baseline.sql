-- ═══════════════════════════════════════════════════════════════════════
-- BASELINE: ColPsiCarabobo — Esquema completo consolidado
-- Fecha: 2026-07-26
-- Descripción: Migración única que reemplaza todas las migraciones
--              anteriores. Incluye el esquema final + índices optimizados.
-- ═══════════════════════════════════════════════════════════════════════

-- ── Extensiones ─────────────────────────────────────────────────────
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "pg_uuidv7";
CREATE EXTENSION IF NOT EXISTS "unaccent";

-- ═══════════════════════════════════════════════════════════════════════
-- TABLAS
-- ═══════════════════════════════════════════════════════════════════════

-- ── psi_specialty_models ────────────────────────────────────────────
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
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_psi_specialty_models_name" UNIQUE ("name")
);

-- ── text_models ─────────────────────────────────────────────────────
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

-- ── user_admins ─────────────────────────────────────────────────────
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
  "must_change_password" boolean NULL DEFAULT false,
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

-- ── psi_users ───────────────────────────────────────────────────────
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
  "must_change_password" boolean NULL DEFAULT false,
  "audio_book_shell_id" character varying(50) NULL,
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
  "primary_specialty_id" integer NULL,
  "secondary_specialty_id" integer NULL,
  "mini_bio" text NULL,
  "bio_text_id" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_psi_users_audio_book_shell_id" UNIQUE ("audio_book_shell_id"),
  CONSTRAINT "uni_psi_users_email" UNIQUE ("email"),
  CONSTRAINT "uni_psi_users_username" UNIQUE ("username"),
  CONSTRAINT "fk_psi_users_full_bio" FOREIGN KEY ("bio_text_id") REFERENCES "text_models" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_psi_users_primary_specialty" FOREIGN KEY ("primary_specialty_id") REFERENCES "psi_specialty_models" ("id") ON DELETE SET NULL,
  CONSTRAINT "fk_psi_users_secondary_specialty" FOREIGN KEY ("secondary_specialty_id") REFERENCES "psi_specialty_models" ("id") ON DELETE SET NULL
);

-- ── psi_user_col_data ───────────────────────────────────────────────
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

-- ── psi_user_post_grades ────────────────────────────────────────────
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

-- ── psi_user_social_networks ────────────────────────────────────────
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

-- ── psi_user_solvency ───────────────────────────────────────────────
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

-- ── psi_observations ────────────────────────────────────────────────
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

-- ── psi_deontologia ─────────────────────────────────────────────────
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

-- ── posts ───────────────────────────────────────────────────────────
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

-- ── search_events ───────────────────────────────────────────────────
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

-- ── page_views ──────────────────────────────────────────────────────
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

-- ── profile_views ───────────────────────────────────────────────────
CREATE TABLE "profile_views" (
  "id" bigserial NOT NULL,
  "psi_id" uuid NOT NULL,
  "viewer_id" uuid NULL,
  "session_id" character varying(64) NULL,
  "ip" character varying(45) NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id")
);

-- ── login_events ────────────────────────────────────────────────────
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

-- ── active_sessions ─────────────────────────────────────────────────
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


-- ═══════════════════════════════════════════════════════════════════════
-- ÍNDICES ÚNICOS Y DE RESTRICCIÓN
-- ═══════════════════════════════════════════════════════════════════════

-- user_admins: SUDO único globalmente (ignora soft-deleted)
CREATE UNIQUE INDEX "idx_user_admins_unique_sudo"
  ON "user_admins" ("sudo")
  WHERE "sudo" IS TRUE AND "deleted_at" IS NULL;

-- psi_users: Identificadores únicos
CREATE UNIQUE INDEX "idx_psi_users_ci" ON "psi_users" ("ci");
CREATE UNIQUE INDEX "idx_psi_users_fpv" ON "psi_users" ("fpv");

-- psi_user_col_data: Relación 1:1 con psi_users
CREATE UNIQUE INDEX "idx_psi_user_col_data_psi_user_model_id"
  ON "psi_user_col_data" ("psi_user_model_id");

-- psi_user_solvency: Unicidad compuesta (una solvencia por fecha por usuario)
CREATE UNIQUE INDEX "idx_psi_solvency_unique"
  ON "psi_user_solvency" ("psi_user_model_id", "date");

-- active_sessions: Una sesión activa por usuario (FIX: era ("user_id","user_id"))
CREATE UNIQUE INDEX "idx_active_sessions_user_id"
  ON "active_sessions" ("user_id");


-- ═══════════════════════════════════════════════════════════════════════
-- ÍNDICES GORM (Soft Delete + FK lookup)
-- ═══════════════════════════════════════════════════════════════════════

-- Soft delete (requeridos por GORM WHERE deleted_at IS NULL)
CREATE INDEX "idx_user_admins_deleted_at" ON "user_admins" ("deleted_at");
CREATE INDEX "idx_text_models_deleted_at" ON "text_models" ("deleted_at");
CREATE INDEX "idx_psi_specialty_models_deleted_at" ON "psi_specialty_models" ("deleted_at");
CREATE INDEX "idx_psi_user_col_data_deleted_at" ON "psi_user_col_data" ("deleted_at");
CREATE INDEX "idx_psi_user_post_grades_deleted_at" ON "psi_user_post_grades" ("deleted_at");
CREATE INDEX "idx_psi_user_social_networks_deleted_at" ON "psi_user_social_networks" ("deleted_at");
CREATE INDEX "idx_psi_user_solvency_deleted_at" ON "psi_user_solvency" ("deleted_at");
CREATE INDEX "idx_psi_observations_deleted_at" ON "psi_observations" ("deleted_at");
CREATE INDEX "idx_psi_deontologia_deleted_at" ON "psi_deontologia" ("deleted_at");
CREATE INDEX "idx_posts_deleted_at" ON "posts" ("deleted_at");
CREATE INDEX "idx_login_events_deleted_at" ON "login_events" ("deleted_at");

-- FK lookups en relaciones 1:N
CREATE INDEX "idx_psi_user_post_grades_psi_user_id" ON "psi_user_post_grades" ("psi_user_id");
CREATE INDEX "idx_psi_user_social_networks_psi_user_id" ON "psi_user_social_networks" ("psi_user_id");
CREATE INDEX "idx_psi_observations_psi_user_id" ON "psi_observations" ("psi_user_id");
CREATE INDEX "idx_psi_deontologia_psi_user_id" ON "psi_deontologia" ("psi_user_id");
CREATE INDEX "idx_login_events_user_id" ON "login_events" ("user_id");


-- ═══════════════════════════════════════════════════════════════════════
-- ÍNDICES DE RENDIMIENTO — psi_users (Directorio + Búsqueda)
-- ═══════════════════════════════════════════════════════════════════════

-- Filtrado por especialidad FK (agregada en FIX-31, sin index hasta ahora)
CREATE INDEX "idx_psi_users_primary_specialty_id" ON "psi_users" ("primary_specialty_id");
CREATE INDEX "idx_psi_users_secondary_specialty_id" ON "psi_users" ("secondary_specialty_id");

-- Filtrado por género (SearchAdmin)
CREATE INDEX "idx_psi_users_genre" ON "psi_users" ("genre");

-- Directorio público: filtro base + orden (solventes primero)
CREATE INDEX "idx_psi_users_active_solvent" ON "psi_users" ("is_active", "solvent");

-- Sitemap: WHERE is_active AND solvent
CREATE INDEX "idx_psi_users_active_solvent_pic" ON "psi_users" ("is_active", "solvent", "profile_picture_s3_key");

-- Búsqueda ILIKE por nombre con unaccent (SearchDirectory)
-- Expresión exacta que coincide con la query Go: COALESCE(first_name,'') || ' ' || COALESCE(second_name,'') || ' ' || COALESCE(last_name,'') || ' ' || COALESCE(second_last_name,'')
CREATE INDEX "idx_psi_users_unaccent_full_name" ON "psi_users" (
  unaccent(COALESCE("first_name", '') || ' ' || COALESCE("second_name", '') || ' ' || COALESCE("last_name", '') || ' ' || COALESCE("second_last_name", ''))
);

-- Búsqueda ILIKE por nombre con unaccent (SearchAdmin — campos individuales)
CREATE INDEX "idx_psi_users_unaccent_first_name" ON "psi_users" (unaccent("first_name"));
CREATE INDEX "idx_psi_users_unaccent_last_name" ON "psi_users" (unaccent("last_name"));

-- Búsqueda por CI/FPV como texto (CAST en queries Go)
CREATE INDEX "idx_psi_users_ci_text" ON "psi_users" (("ci")::text);
CREATE INDEX "idx_psi_users_fpv_text" ON "psi_users" (("fpv")::text);


-- ═══════════════════════════════════════════════════════════════════════
-- ÍNDICES DE RENDIMIENTO — posts (Listados + Scheduling)
-- ═══════════════════════════════════════════════════════════════════════

-- List(): WHERE status IN (...) ORDER BY created_at DESC
CREATE INDEX "idx_posts_status_created_at" ON "posts" ("status", "created_at" DESC);

-- PublishScheduled(): WHERE status = 'scheduled' AND publish_at <= now()
CREATE INDEX "idx_posts_publish_scheduled" ON "posts" ("publish_at")
  WHERE "status" = 'scheduled';

-- GetSitemapPosts(): WHERE status = 'published' AND type = 'public'
CREATE INDEX "idx_posts_sitemap" ON "posts" ("created_at" DESC)
  WHERE "status" = 'published' AND "type" = 'public';

-- Búsqueda ILIKE por título (PostFilter.Search)
CREATE INDEX "idx_posts_title" ON "posts" ("title");


-- ═══════════════════════════════════════════════════════════════════════
-- ÍNDICES DE RENDIMIENTO — Analytics (Dashboard)
-- ═══════════════════════════════════════════════════════════════════════

-- login_events: Dashboard trend (WHERE created_at >= X)
CREATE INDEX "idx_login_events_created_at" ON "login_events" ("created_at");

-- active_sessions: Heartbeat + cleanup
CREATE INDEX "idx_active_sessions_expires_at" ON "active_sessions" ("expires_at");
CREATE INDEX "idx_active_sessions_last_seen" ON "active_sessions" ("last_seen");

-- search_events: Dashboard aggregations
CREATE INDEX "idx_search_events_created_at" ON "search_events" ("created_at");
CREATE INDEX "idx_search_events_specialty_created" ON "search_events" ("specialty", "created_at")
  WHERE "specialty" != '';
CREATE INDEX "idx_search_events_municipality_created" ON "search_events" ("municipality", "created_at")
  WHERE "municipality" != '';
CREATE INDEX "idx_search_events_query_created" ON "search_events" ("query", "created_at")
  WHERE "query" != '';

-- page_views: Trend + UniqueVisitors
CREATE INDEX "idx_page_views_created_at" ON "page_views" ("created_at");
CREATE INDEX "idx_page_views_session_created" ON "page_views" ("session_id", "created_at");

-- profile_views: Top perfiles + trend
CREATE INDEX "idx_profile_views_created_at" ON "profile_views" ("created_at");
CREATE INDEX "idx_profile_views_psi_id" ON "profile_views" ("psi_id");
CREATE INDEX "idx_profile_views_psi_created" ON "profile_views" ("psi_id", "created_at");
CREATE INDEX "idx_profile_views_viewer_id" ON "profile_views" ("viewer_id");

-- psi_specialty_models: List ORDER BY name
CREATE INDEX "idx_psi_specialty_models_active_name" ON "psi_specialty_models" ("active", "name");


-- ═══════════════════════════════════════════════════════════════════════
-- COMENTARIOS DE POLÍTICA — FK DELETE
-- ═══════════════════════════════════════════════════════════════════════

COMMENT ON CONSTRAINT fk_psi_users_col_data ON psi_user_col_data IS
    'ON DELETE NO ACTION — Soft delete only. Hard delete requiere borrar hijos primero.';
COMMENT ON CONSTRAINT fk_psi_users_post_grades ON psi_user_post_grades IS
    'ON DELETE NO ACTION — Soft delete only. Hard delete requiere borrar hijos primero.';
COMMENT ON CONSTRAINT fk_psi_users_social_networks ON psi_user_social_networks IS
    'ON DELETE NO ACTION — Soft delete only. Hard delete requiere borrar hijos primero.';
COMMENT ON CONSTRAINT fk_psi_users_solvencies ON psi_user_solvency IS
    'ON DELETE NO ACTION — Soft delete only. Hard delete requiere borrar hijos primero.';
COMMENT ON CONSTRAINT fk_posts_text ON posts IS
    'ON DELETE NO ACTION — Soft delete only. Hard delete requiere borrar hijos primero.';
COMMENT ON CONSTRAINT fk_psi_users_full_bio ON psi_users IS
    'ON DELETE NO ACTION — Soft delete only. Bio es opcional, puede ser NULL.';
COMMENT ON CONSTRAINT fk_psi_users_primary_specialty ON psi_users IS
    'ON DELETE SET NULL — Si se borra la especialidad, se limpia la referencia.';
COMMENT ON CONSTRAINT fk_psi_users_secondary_specialty ON psi_users IS
    'ON DELETE SET NULL — Si se borra la especialidad, se limpia la referencia.';
