-- these are NOT public API so service ONLY
CREATE SCHEMA IF NOT EXISTS service;

CREATE TABLE IF NOT EXISTS users (
    email TEXT PRIMARY KEY NOT NULL,
    last_login TIMESTAMPTZ DEFAULT date_trunc('milliseconds', NOW() AT TIME ZONE 'UTC')
);

CREATE TABLE IF NOT EXISTS user_analytics (
    email TEXT PRIMARY KEY NOT NULL,
    analytics_status TEXT DEFAULT 'none',
    applications_count INT DEFAULT 0,
    applied_count INT DEFAULT 0,
    ghosted_count INT DEFAULT 0,
    rejected_count INT DEFAULT 0,
    screen_count INT DEFAULT 0,
    interviewing_count INT DEFAULT 0,
    offer_count INT DEFAULT 0,
    -- all complex analytics CAN be null
    application_velocity INT DEFAULT NULL,
    application_velocity_trend INT DEFAULT NULL,
    resume_effectiveness INT DEFAULT NULL,
    resume_effectiveness_trend INT DEFAULT NULL,
    interview_effectiveness INT DEFAULT NULL,
    interview_effectiveness_trend INT DEFAULT NULL,
    avg_first_response_time INT DEFAULT NULL,
    avg_first_response_time_trend INT DEFAULT NULL,
    yearly_trends JSONB DEFAULT '{}'::JSONB,
    CONSTRAINT fk_user_analytics_user
        FOREIGN KEY (email)
        REFERENCES users(email)
        ON DELETE CASCADE
);

-- reads are more frequqent than writes, so having two tables is better
-- one for the history of applications and one for the current state
-- both need different indexes for performance
CREATE TABLE IF NOT EXISTS user_applications (
    application_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL,
    applied_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- latest operation (add or edit or reverted), simply used to see if avg first response 
    -- analytics has to be recalculated (i.e. if latest operation was add)
    operation TEXT, 
    app_status TEXT,
    company TEXT,
    title TEXT,
    link TEXT,
    locations TEXT,
    CONSTRAINT fk_user_applications_user
        FOREIGN KEY (email)
        REFERENCES users(email)
        ON DELETE CASCADE
);

-- bit of redundant data here but just it's to prevent joining
-- so yeah writes might be a bit slower but reads and analytics are way faster
CREATE TABLE IF NOT EXISTS application_history (
    operation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL,
    email TEXT NOT NULL,
    applied_date TIMESTAMPTZ NOT NULL,
    event_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    app_status TEXT,
    operation TEXT,
    CONSTRAINT fk_application_history_app
        FOREIGN KEY (application_id)
        REFERENCES user_applications(application_id)
        ON DELETE CASCADE,
    CONSTRAINT fk_application_history_user
        FOREIGN KEY (email)
        REFERENCES users(email)
        ON DELETE CASCADE
);

-- user_applications index for quick search and dashboard results
CREATE INDEX IF NOT EXISTS idx_user_applications_email_applied_date_appid
    ON user_applications (email, applied_date DESC, application_id);

-- application_history index for timeline for quicker timeline retrieval
-- timeline is application_id specific and returns event_time descending
CREATE INDEX IF NOT EXISTS idx_application_history_email_appid_event_time
    ON application_history (email, application_id, event_time DESC);

-- application_history index for analytics so recalculations are faster
-- analytics scan ALL applications and ALL events per application but processes
-- only recent events which is why we index as such
CREATE INDEX IF NOT EXISTS idx_application_history_email_applied_date
    ON application_history (email, applied_date DESC)
    WHERE operation != 'revert';

CREATE INDEX IF NOT EXISTS idx_user_email
    ON users (email);

CREATE INDEX IF NOT EXISTS idx_user_analytics_email
    ON user_analytics (email);