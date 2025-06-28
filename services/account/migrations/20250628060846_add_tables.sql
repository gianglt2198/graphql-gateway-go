-- Create "users" table
CREATE TABLE "users" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "created_by" character varying NULL,
  "updated_at" timestamptz NOT NULL,
  "updated_by" character varying NULL,
  "deleted_at" timestamptz NULL,
  "deleted_by" character varying NULL,
  "username" character varying NOT NULL,
  "email" character varying NOT NULL,
  "password" character varying NOT NULL,
  "first_name" character varying NULL,
  "last_name" character varying NULL,
  "phone" character varying NULL,
  PRIMARY KEY ("id")
);
-- Create index "users_username_key" to table: "users"
CREATE UNIQUE INDEX "users_username_key" ON "users" ("username");
-- Create index "users_email_key" to table: "users"
CREATE UNIQUE INDEX "users_email_key" ON "users" ("email");
-- Create "sessions" table
CREATE TABLE "sessions" (
  "id" character varying NOT NULL,
  "last_used_at" timestamptz NOT NULL,
  "user_sessions" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sessions_users_sessions" FOREIGN KEY ("user_sessions") REFERENCES "users" ("id") ON DELETE NO ACTION
);
-- Create index "sessions_user_sessions_key" to table: "sessions"
CREATE UNIQUE INDEX "sessions_user_sessions_key" ON "sessions" ("user_sessions");
