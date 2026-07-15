ALTER TABLE notifications ADD COLUMN idempotency_key VARCHAR(255) UNIQUE;
