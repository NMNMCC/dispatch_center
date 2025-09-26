-- create "users" table
CREATE TABLE "users" ("id" uuid NOT NULL, "email" character varying NOT NULL, "password" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "users_email_key" to table: "users"
CREATE UNIQUE INDEX "users_email_key" ON "users" ("email");
-- create "keys" table
CREATE TABLE "keys" ("id" uuid NOT NULL, "body" character varying NOT NULL, "permissions" jsonb NOT NULL, "created_at" timestamptz NOT NULL, "revoked_at" timestamptz NOT NULL, "key_user" uuid NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "keys_users_user" FOREIGN KEY ("key_user") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
