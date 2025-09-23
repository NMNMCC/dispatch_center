-- reverse: create index "workers_task_worker_key" to table: "workers"
DROP INDEX "workers_task_worker_key";
-- reverse: create "workers" table
DROP TABLE "workers";
-- reverse: create "tag_tasks" table
DROP TABLE "tag_tasks";
-- reverse: create "tasks" table
DROP TABLE "tasks";
-- reverse: create index "tags_name_key" to table: "tags"
DROP INDEX "tags_name_key";
-- reverse: create "tags" table
DROP TABLE "tags";
