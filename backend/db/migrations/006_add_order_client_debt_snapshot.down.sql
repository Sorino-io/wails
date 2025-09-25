-- Down migration: cannot reliably remove column in SQLite without table rebuild.
-- We perform a no-op or you could recreate the table if strict rollback is required.
-- For simplicity, we leave it as a no-op.
SELECT 1;