-- Create "sessions" table
CREATE TABLE "sessions" (
  "id" character varying NOT NULL,
  "user_id" character varying NOT NULL,
  "last_used_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "sessions_user_id_key" to table: "sessions"
CREATE UNIQUE INDEX "sessions_user_id_key" ON "sessions" ("user_id");
-- Create "users" table
CREATE TABLE "users" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "created_by" character varying NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "updated_by" character varying NULL,
  "deleted_at" timestamptz NULL,
  "deleted_by" character varying NULL,
  "username" character varying NOT NULL,
  "email" character varying NOT NULL,
  "password" character varying NOT NULL,
  "first_name" character varying NOT NULL,
  "last_name" character varying NOT NULL,
  "phone" character varying NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "users_username_key" to table: "users"
CREATE UNIQUE INDEX "users_username_key" ON "users" ("username");
-- Create index "users_email_key" to table: "users"
CREATE UNIQUE INDEX "users_email_key" ON "users" ("email");
