-- create "tags" table
CREATE TABLE "tags" ("id" uuid NOT NULL, "name" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "tags_name_key" to table: "tags"
CREATE UNIQUE INDEX "tags_name_key" ON "tags" ("name");
-- create "tasks" table
CREATE TABLE "tasks" ("id" uuid NOT NULL, "status" character varying NOT NULL DEFAULT 'pending', "body" jsonb NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, PRIMARY KEY ("id"));
-- create "tag_tasks" table
CREATE TABLE "tag_tasks" ("tag_id" uuid NOT NULL, "task_id" uuid NOT NULL, PRIMARY KEY ("tag_id", "task_id"), CONSTRAINT "tag_tasks_tag_id" FOREIGN KEY ("tag_id") REFERENCES "tags" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "tag_tasks_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create "workers" table
CREATE TABLE "workers" ("id" uuid NOT NULL, "end_of_life" timestamptz NOT NULL, "registered_at" timestamptz NOT NULL, "task_worker" uuid NULL, PRIMARY KEY ("id"), CONSTRAINT "workers_tasks_worker" FOREIGN KEY ("task_worker") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "workers_task_worker_key" to table: "workers"
CREATE UNIQUE INDEX "workers_task_worker_key" ON "workers" ("task_worker");
