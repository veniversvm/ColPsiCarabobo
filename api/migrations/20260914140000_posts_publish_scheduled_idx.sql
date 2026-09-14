-- Create index "idx_posts_status_publish_at" to table: "posts"
CREATE INDEX "idx_posts_status_publish_at" ON "posts" ("status", "publish_at") WHERE "deleted_at" IS NULL;