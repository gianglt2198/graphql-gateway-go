-- Create "products" table
CREATE TABLE "products" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "created_by" character varying NULL,
  "updated_at" timestamptz NOT NULL,
  "updated_by" character varying NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "price" double precision NOT NULL DEFAULT 0,
  "stock" bigint NOT NULL DEFAULT 0,
  PRIMARY KEY ("id")
);
-- Create "categories" table
CREATE TABLE "categories" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "created_by" character varying NULL,
  "updated_at" timestamptz NOT NULL,
  "updated_by" character varying NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  PRIMARY KEY ("id")
);
-- Create "product_categories" table
CREATE TABLE "product_categories" (
  "product_id" character varying NOT NULL,
  "category_id" character varying NOT NULL,
  PRIMARY KEY ("product_id", "category_id"),
  CONSTRAINT "product_categories_product_id" FOREIGN KEY ("product_id") REFERENCES "products" ("id") ON DELETE CASCADE,
  CONSTRAINT "product_categories_category_id" FOREIGN KEY ("category_id") REFERENCES "categories" ("id") ON DELETE CASCADE
);
