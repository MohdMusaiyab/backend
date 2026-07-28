CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    phone VARCHAR(50),
    -- The magical JSONB column for infinite flexibility
    preferences JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- We use a COMPOSITE primary key (name, version) to enforce Immutable Versioning!
CREATE TABLE templates (
    name VARCHAR(100) NOT NULL,
    version VARCHAR(50) NOT NULL,
    subject_template TEXT NOT NULL,
    body_template TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (name, version)
);

-- =========================================================================
-- SEED DATA (For easy testing later)
-- =========================================================================

-- 1. A mock user who explicitly Opts-In to Email, but Opts-Out of SMS
INSERT INTO users (id, email, phone, preferences) VALUES (
    '11111111-1111-1111-1111-111111111111', 
    'john.doe@example.com', 
    '+1234567890', 
    '{"channels": {"email": true, "sms": false}}'::jsonb
);

-- 2. A mock template (v1) using Go's text/template syntax ({{.Variable}})
INSERT INTO templates (name, version, subject_template, body_template) VALUES (
    'welcome_email',
    'v1',
    'Welcome to our store, {{.FirstName}}!',
    'Hello {{.FirstName}}, thanks for joining! Your secret discount code is {{.Code}}.'
);
