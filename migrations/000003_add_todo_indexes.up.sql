-- Filename: migrations/000003_add_todo_indexes.up.sql

CREATE INDEX IF NOT EXISTS todotbl_task_name_idx ON todotbl USING GIN(to_tsvector('simple', task_name));
CREATE INDEX IF NOT EXISTS todotbl_priority_idx ON todotbl USING GIN(to_tsvector('simple', priority));
CREATE INDEX IF NOT EXISTS todotbl_status_idx ON todotbl USING GIN(status);