-- Modify "psi_user_col_data" table
ALTER TABLE "psi_user_col_data" ADD COLUMN "title_image_one_s3_key" character varying(512) NULL, ADD COLUMN "title_image_two_s3_key" character varying(512) NULL, ADD COLUMN "title_image_three_s3_key" character varying(512) NULL;
