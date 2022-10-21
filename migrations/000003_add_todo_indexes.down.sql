-- Filename: migrations/000003_add_todo_indexes.down.sql

DROP INDEX IF EXISTS todotbl_task_name_idx;
DROP INDEX IF EXISTS todotbl_priority_idx;
DROP INDEX IF EXISTS todotbl_status_idx;