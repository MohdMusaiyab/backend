-- We use ALTER TABLE to safely add a new column without touching existing data.
-- We do NOT add "NOT NULL" because old notifications (and instant ones) will not have this value.
ALTER TABLE notifications 
ADD COLUMN send_at TIMESTAMP WITH TIME ZONE;
