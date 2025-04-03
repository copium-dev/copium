-- handle user login and update last login time 
CREATE OR REPLACE FUNCTION service.login(p_email TEXT)
RETURNS BOOLEAN AS $$
BEGIN
    -- insert into users table, update last login if user already exists
    INSERT INTO service.users (email, last_login)
    VALUES (p_email, date_trunc('milliseconds', NOW() AT TIME ZONE 'UTC'))
    ON CONFLICT (email) DO UPDATE
    SET last_login = date_trunc('milliseconds', NOW() AT TIME ZONE 'UTC');

    -- initialize user_analytics so user can see dashboard w/o adding anything
    -- schema already defaults everything to 0 so don't worry
    INSERT INTO service.user_analytics
        (email)
    VALUES (p_email)
    ON CONFLICT (email) DO NOTHING;

    RETURN TRUE;
EXCEPTION
    WHEN OTHERS THEN
        RAISE NOTICE 'Error in login function: %', SQLERRM;
        RETURN FALSE;
END;
$$ LANGUAGE plpgsql;

-- handle loading profile (just read all fields of user_analytics)
CREATE OR REPLACE FUNCTION service.profile(p_email TEXT)
RETURNS TABLE (
    applications_count INT,
    applied_count INT,
    ghosted_count INT,
    rejected_count INT,
    screen_count INT,
    interviewing_count INT,
    offer_count INT,
    application_velocity INT,
    application_velocity_trend INT,
    resume_effectiveness INT,
    resume_effectiveness_trend INT,
    interview_effectiveness INT,
    interview_effectiveness_trend INT,
    avg_first_response_time INT,
    prev_avg_first_response_time INT,
    avg_first_response_time_trend INT,
    yearly_trends JSONB
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        ua.applications_count,
        ua.applied_count,
        ua.ghosted_count,
        ua.rejected_count,
        ua.screen_count,
        ua.interviewing_count,
        ua.offer_count,
        ua.application_velocity AS application_velocity,
        ua.application_velocity_trend AS application_velocity_trend,
        ua.resume_effectiveness AS resume_effectiveness,
        ua.resume_effectiveness_trend AS resume_effectiveness_trend,
        ua.interview_effectiveness AS interview_effectiveness,
        ua.interview_effectiveness_trend AS interview_effectiveness_trend,
        ua.avg_first_response_time AS avg_first_response_time,
        ua.prev_avg_first_response_time AS prev_avg_first_response_time,
        ua.avg_first_response_time_trend AS avg_first_response_time_trend,
        COALESCE(ua.yearly_trends, '{}'::JSONB) AS yearly_trends
    FROM service.user_analytics ua
    WHERE ua.email = p_email;
END;
$$ LANGUAGE plpgsql;

-- handles both cases of not cached and cached page boundary searches
-- if cached (non-null p_last_applied_date), we have fast keyset pagination
-- if not cached (null p_last_applied_date), we have offset-based pagination
-- this is to get the best of both worlds: allow jumps but fast sequential navigation
CREATE OR REPLACE FUNCTION service.get_user_applications(
    p_email TEXT,
    p_query TEXT,
    p_limit INT,
    p_offset INT,                   -- fallback offset if p_last_applied_date is null
    p_last_applied_date TIMESTAMPTZ   -- if not cached or page == 1, this is null
)
RETURNS TABLE (
    application_id UUID,
    email TEXT,
    applied_date TIMESTAMPTZ,
    operation TEXT,
    app_status TEXT,
    company TEXT,
    title TEXT,
    link TEXT,
    locations TEXT,
    total_hits BIGINT
) AS $$
DECLARE
    v_total_hits BIGINT;
BEGIN
    -- get total hits first
    SELECT COUNT(DISTINCT ua.application_id) INTO v_total_hits
    FROM service.user_applications ua
    WHERE ua.email = p_email
        AND (
            p_query IS NULL
            OR p_query = ''
            OR to_tsvector('english', ua.title || ' ' || ua.company || ' ' || ua.locations)
                @@ websearch_to_tsquery('english', p_query || ':*')
        );
    
    IF p_last_applied_date IS NOT NULL THEN
        -- prev page boundary cached, just get everything before it up to limit
        -- this makes sequential navigation very fast. we cache page boundaries for every
        -- page so if we jump to page 90 we can do fast sequential navigation
        -- a potential optimization (if page numbers get crazy) is to cache surrounding pages too
        RETURN QUERY
        SELECT 
            ua.application_id,
            ua.email,
            ua.applied_date,
            ua.operation,
            ua.app_status,
            ua.company,
            ua.title,
            ua.link,
            ua.locations,
            v_total_hits
        FROM service.user_applications ua
        WHERE ua.email = p_email
            AND ua.applied_date < p_last_applied_date
            -- full text search on title OR company OR locations
            AND (
                p_query IS NULL
                OR p_query = ''
                OR to_tsvector('english', ua.title || ' ' || ua.company || ' ' || ua.locations)
                   @@ websearch_to_tsquery('english', p_query || ':*') -- allow prefix
            )
        ORDER BY ua.applied_date DESC
        LIMIT p_limit;
    ELSE
        -- offset-based if user jumps to page whose previous page has not been cached
        RETURN QUERY
        SELECT 
            ua.application_id,
            ua.email,
            ua.applied_date,
            ua.operation,
            ua.app_status,
            ua.company,
            ua.title,
            ua.link,
            ua.locations,
            v_total_hits
        FROM service.user_applications ua
        WHERE ua.email = p_email
            AND (
                p_query IS NULL
                OR p_query = ''
                OR to_tsvector('english', ua.title || ' ' || ua.company || ' ' || ua.locations)
                    @@ websearch_to_tsquery('english', p_query || ':*') -- allow prefix
            )
        ORDER BY ua.applied_date DESC
        OFFSET p_offset
        LIMIT p_limit;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- add application, updates applied_count, applications_count
CREATE OR REPLACE FUNCTION service.add_application(
    p_email TEXT,
    p_applied_date TIMESTAMPTZ,
    p_operation TEXT,
    p_app_status TEXT,
    p_company TEXT,
    p_title TEXT,
    p_link TEXT,
    p_locations TEXT
)
RETURNS UUID AS $$
DECLARE
    v_application_id UUID;
BEGIN
    INSERT INTO service.user_applications (
        application_id,
        email,
        applied_date,
        operation,
        app_status,
        company,
        title,
        link,
        locations
    )
    VALUES (
        gen_random_uuid(),
        p_email,
        p_applied_date,
        p_operation,
        p_app_status,
        p_company,
        p_title,
        p_link,
        p_locations
    )
    RETURNING application_id INTO v_application_id;
    
    INSERT INTO service.application_history (
        operation_id,
        application_id,
        email,
        applied_date,
        event_time,
        app_status,
        operation
    )
    VALUES (
        gen_random_uuid(),
        v_application_id,
        p_email,
        p_applied_date,
        date_trunc('milliseconds', NOW() AT TIME ZONE 'UTC'),
        p_app_status,
        p_operation
    );
    
    UPDATE service.user_analytics
    SET
        analytics_status = 'pending',
        applications_count = applications_count + 1,
        applied_count = applied_count + 1
    WHERE email = p_email;
    
    PERFORM service.full_recalculate_analytics(p_email);
    
    RETURN v_application_id;
END;
$$ LANGUAGE plpgsql;

-- full recalculation of analytics (only used on application deletes)
CREATE OR REPLACE FUNCTION service.delete_application(p_email TEXT, p_application_id UUID)
RETURNS BOOLEAN AS $$
DECLARE
    v_status TEXT;
BEGIN
    -- wait: before you recalculate you have to actually update counts in user_analytics lol
    -- either we can trust the client to send proper 'old status' or we can do safer way and
    -- just get the most recent status from application_history. but obviously this is +1 read
    -- BUT index should make this good enough
    SELECT app_status INTO v_status
    FROM service.application_history
    WHERE email = p_email 
        AND application_id = p_application_id
    ORDER BY event_time DESC
    LIMIT 1;

    UPDATE service.user_analytics
    SET 
        applications_count = GREATEST(0, applications_count - 1), 
        -- only decrement the status count that matches the current status
        ghosted_count = CASE WHEN v_status = 'Ghosted' THEN GREATEST(0, ghosted_count - 1) ELSE ghosted_count END,
        rejected_count = CASE WHEN v_status = 'Rejected' THEN GREATEST(0, rejected_count - 1) ELSE rejected_count END,
        screen_count = CASE WHEN v_status = 'Screen' THEN GREATEST(0, screen_count - 1) ELSE screen_count END,
        interviewing_count = CASE WHEN v_status = 'Interviewing' THEN GREATEST(0, interviewing_count - 1) ELSE interviewing_count END,
        offer_count = CASE WHEN v_status = 'Offer' THEN GREATEST(0, offer_count - 1) ELSE offer_count END,
        applied_count = CASE WHEN v_status = 'Applied' THEN GREATEST(0, applied_count - 1) ELSE applied_count END
    WHERE email = p_email;

    -- now delete the application from user_applications and application_history
    -- we already cascade deletes so just delete from user_applications
    DELETE FROM service.user_applications
    WHERE email = p_email
        AND application_id = p_application_id; 

    PERFORM service.full_recalculate_analytics(p_email);

    RETURN TRUE;
EXCEPTION
    WHEN OTHERS THEN
        RAISE NOTICE 'Error in full_recalculate_analytics function: %', SQLERRM;
        RETURN FALSE;
END;
$$ LANGUAGE plpgsql;

-- delete user (this is very easy since cascading deletes
-- will take care of everything else)
CREATE OR REPLACE FUNCTION service.delete_user(p_email TEXT)
RETURNS BOOLEAN AS $$
BEGIN
    -- delete user from users table
    -- all tables use cascading deletes on this primary key
    DELETE FROM service.users
    WHERE email = p_email;

    RETURN TRUE;
EXCEPTION
    WHEN OTHERS THEN
        RAISE NOTICE 'Error in delete_user function: %', SQLERRM;
        RETURN FALSE;
END;
$$ LANGUAGE plpgsql;

-- edit application, simple. just updates metadata, nothing in history and no analytics change
CREATE OR REPLACE FUNCTION service.edit_application(
    p_email TEXT,
    p_application_id UUID,
    p_company TEXT DEFAULT NULL,
    p_title TEXT DEFAULT NULL,
    p_link TEXT DEFAULT NULL,
    p_locations TEXT DEFAULT NULL
)
RETURNS BOOLEAN AS $$
BEGIN
    UPDATE service.user_applications
    SET 
        company = COALESCE(p_company, company),
        title = COALESCE(p_title, title),
        link = COALESCE(p_link, link),
        locations = COALESCE(p_locations, locations)
    WHERE email = p_email
    AND application_id = p_application_id;

    RETURN TRUE;
EXCEPTION
    WHEN OTHERS THEN
        RAISE NOTICE 'Error in edit_application function: %', SQLERRM;
        RETURN FALSE;
END;
$$ LANGUAGE plpgsql;

-- update edit status; this requires looking at most recent history and determining
-- which analytic to update. like delete, we COULD just trust the client
-- to send the correct previous status but to be safe we'll get it ourself
CREATE OR REPLACE FUNCTION service.update_application_status(
    p_email TEXT,
    p_application_id UUID,
    p_app_status TEXT   -- the new status
)
RETURNS BOOLEAN AS $$
DECLARE
    v_prev_status TEXT;
    v_applied_date TIMESTAMPTZ;
BEGIN
    -- get curr status and applied date of this application
    -- NOTE: applied date is put in history; its redundant but avoids joins
    SELECT app_status, applied_date
    INTO v_prev_status, v_applied_date
    FROM service.user_applications
    WHERE email = p_email
        AND application_id = p_application_id;

    IF v_prev_status IS NOT DISTINCT FROM p_app_status THEN
        RAISE NOTICE 'Status is same, no update needed';
        RETURN TRUE;
    END IF;

    -- atp confirmed to be a new event; update status and append to history
    UPDATE service.user_applications ua
    SET 
        app_status = p_app_status,
        operation = 'edit'
    WHERE ua.email = p_email
        AND ua.application_id = p_application_id;

    INSERT INTO service.application_history (
        operation_id,
        application_id,
        email,
        applied_date,
        event_time,
        app_status,
        operation
    )
    VALUES (
        gen_random_uuid(),
        p_application_id,
        p_email,
        v_applied_date,
        date_trunc('milliseconds', NOW() AT TIME ZONE 'UTC'),
        p_app_status,
        'edit'
    );
    
    -- update counts
    PERFORM service.update_application_counts(p_email, v_prev_status, p_app_status);

    -- update analytics status to pending
    PERFORM service.full_recalculate_analytics(p_email);
    
    RETURN TRUE;
EXCEPTION
    WHEN OTHERS THEN
        RAISE NOTICE 'Error in update_application_status function: %', SQLERRM;
        RETURN FALSE;
END;
$$ LANGUAGE plpgsql;

-- get all application history
CREATE OR REPLACE FUNCTION service.get_application_history(
    p_email TEXT,
    p_application_id UUID
)
RETURNS TABLE (
    operation_id UUID,
    application_id UUID,
    email TEXT,
    applied_date TIMESTAMPTZ,
    event_time TIMESTAMPTZ,
    app_status TEXT,
    operation TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        operation_id,
        application_id,
        email,
        applied_date,
        event_time,
        app_status,
        operation
    FROM service.application_history ah
    WHERE ah.email = p_email
      AND ah.application_id = p_application_id
    ORDER BY event_time DESC;
END;
$$ LANGUAGE plpgsql;

-- revert an operation; bit tricky since we have no joins, we have to check two cases:
-- 1. most recent operation: update the user_applications table to previous status (from history)
-- 2. not most recent: just delete most recent operation (soft delete; its flagged as reverted just in case)
-- in both cases need to update status counts and obviously full recalculation of analytics
CREATE OR REPLACE FUNCTION service.revert_operation(
    p_email TEXT,
    p_application_id UUID,
    p_operation_id UUID
)
RETURNS TEXT AS $$
DECLARE
    v_latest_op RECORD;
    v_operation_type TEXT;
    v_prev_status TEXT;
    v_is_latest BOOLEAN;
BEGIN
    SELECT operation
    INTO v_operation_type
    FROM service.application_history
    WHERE email = p_email
        AND application_id = p_application_id
        AND operation_id = p_operation_id;

    IF v_operation_type IN ('revert','add') THEN
        RAISE EXCEPTION 'Cannot revert: Operation is add';
        RETURN FALSE;
    END IF;

    -- get two latest operations. we have to handle case where operation latest or not latest
    -- 1. operation to revert is latest: use second latest operation's status and -= 1 current status, flag as reverted
    -- 2. operation to revert is not latest: just flag as reverted and -= 1 the status of that operation
    -- i would have used a CTE but in Postgres the result set does not extend past the query
    -- which very stupid if u ask me
    CREATE TEMP TABLE numbered_ops ON COMMIT DROP AS
    SELECT 
        operation_id,
        app_status,
        operation,
        applied_date,
        event_time,
        ROW_NUMBER() OVER (ORDER BY event_time DESC) AS rn
    FROM service.application_history
    WHERE email = p_email
        AND application_id = p_application_id
        AND operation != 'revert'
    ORDER BY event_time DESC
    LIMIT 2;

    SELECT 
        operation_id, app_status, operation, applied_date, event_time
    INTO v_latest_op
    FROM numbered_ops
    WHERE rn = 1;

    -- no latest op found (nothing besides revert)
    IF v_latest_op IS NULL THEN
        RAISE EXCEPTION 'Cannot revert: No valid operations found to revert';
        RETURN FALSE;
    END IF;

    v_is_latest := (v_latest_op.operation_id = p_operation_id);

    -- latest; use rn = 2 operation's status for user_applications
    -- otherwise do nothing special because it didn't revert latest 
    IF v_is_latest THEN
        SELECT app_status
        INTO v_prev_status
        FROM numbered_ops
        WHERE rn = 2;   -- the second row

        -- if no prev status (this shouldn't happen because 'add' is always first and cannot be reverted)
        -- then default prev status to 'Applied'. This is just a safety net for edge cases 
        IF v_prev_status IS NULL THEN
            v_prev_status := 'Applied';
        END IF;

        -- if you're confused, the counts in the user analytics only track latest status
        -- so we only update if this is the latest operation
        PERFORM service.update_application_counts(p_email, v_latest_op.app_status, v_prev_status);

        -- also update the status (not historical just latest)
        UPDATE service.user_applications
        SET app_status = v_prev_status
        WHERE email = p_email
            AND application_id = p_application_id;
    END IF;

    -- set as reverted in history regardless of latest or not
    UPDATE service.application_history
    SET operation = 'revert'
    WHERE email = p_email
        AND application_id = p_application_id
        AND operation_id = p_operation_id;
    
    -- recalculate (like usual)
    PERFORM service.full_recalculate_analytics(p_email);

    -- return previous status if latest, else null
    IF v_is_latest THEN
        RETURN v_prev_status;
    ELSE
        RETURN '';
    END IF;
END;
$$ LANGUAGE plpgsql;

-- update application counts based on previous and new status
CREATE OR REPLACE FUNCTION service.update_application_counts(
    p_email TEXT,
    p_prev_status TEXT,
    p_new_status TEXT
)
RETURNS BOOLEAN AS $$
BEGIN
    -- if status is same, raise notice
    IF p_prev_status = p_new_status THEN
        RAISE NOTICE 'Status is same, no update needed';
        RETURN TRUE;
    END IF;

    UPDATE service.user_analytics
    SET 
        applied_count = GREATEST(0, applied_count + 
            CASE 
                WHEN p_new_status = 'Applied' AND p_prev_status != 'Applied' THEN 1
                WHEN p_new_status != 'Applied' AND p_prev_status = 'Applied' THEN -1
                ELSE 0
            END),
        
        ghosted_count = GREATEST(0, ghosted_count + 
            CASE 
                WHEN p_new_status = 'Ghosted' AND p_prev_status != 'Ghosted' THEN 1
                WHEN p_new_status != 'Ghosted' AND p_prev_status = 'Ghosted' THEN -1
                ELSE 0
            END),
        
        rejected_count = GREATEST(0,rejected_count + 
            CASE 
                WHEN p_new_status = 'Rejected' AND p_prev_status != 'Rejected' THEN 1
                WHEN p_new_status != 'Rejected' AND p_prev_status = 'Rejected' THEN -1
                ELSE 0
            END),
        
        screen_count = GREATEST(0, screen_count + 
            CASE 
                WHEN p_new_status = 'Screen' AND p_prev_status != 'Screen' THEN 1
                WHEN p_new_status != 'Screen' AND p_prev_status = 'Screen' THEN -1
                ELSE 0
            END),
        
        interviewing_count = GREATEST(0, interviewing_count + 
            CASE 
                WHEN p_new_status = 'Interviewing' AND p_prev_status != 'Interviewing' THEN 1
                WHEN p_new_status != 'Interviewing' AND p_prev_status = 'Interviewing' THEN -1
                ELSE 0
            END),
        
        offer_count = GREATEST(0, offer_count + 
            CASE 
                WHEN p_new_status = 'Offer' AND p_prev_status != 'Offer' THEN 1
                WHEN p_new_status != 'Offer' AND p_prev_status = 'Offer' THEN -1
                ELSE 0
            END)
    WHERE email = p_email;

    RETURN TRUE;
EXCEPTION
    WHEN OTHERS THEN
        RAISE NOTICE 'Error in update_application_counts function: %', SQLERRM;
        RETURN FALSE;
END;
$$ LANGUAGE plpgsql;

-- get timeline
CREATE OR REPLACE FUNCTION service.get_application_timeline(
    p_email TEXT,
    p_application_id UUID
)
RETURNS TABLE (
    operation_id UUID,
    event_time TIMESTAMPTZ,
    app_status TEXT,
    operation TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        ah.operation_id,
        ah.event_time,
        ah.app_status,
        ah.operation
    FROM service.application_history ah
    WHERE ah.email = p_email
      AND ah.application_id = p_application_id
      AND ah.operation != 'revert'
    ORDER BY event_time DESC;
END;
$$ LANGUAGE plpgsql;