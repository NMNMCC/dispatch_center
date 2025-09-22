-- create "tags" table
CREATE TABLE "tags" ("id" uuid NOT NULL, "name" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "tags_name_key" to table: "tags"
CREATE UNIQUE INDEX "tags_name_key" ON "tags" ("name");
-- create "tasks" table
CREATE TABLE "tasks" ("id" uuid NOT NULL, "status" character varying NOT NULL DEFAULT 'pending', "body" jsonb NOT NULL, PRIMARY KEY ("id"));
-- create "tag_tasks" table
CREATE TABLE "tag_tasks" ("tag_id" uuid NOT NULL, "task_id" uuid NOT NULL, PRIMARY KEY ("tag_id", "task_id"), CONSTRAINT "tag_tasks_tag_id" FOREIGN KEY ("tag_id") REFERENCES "tags" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "tag_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
