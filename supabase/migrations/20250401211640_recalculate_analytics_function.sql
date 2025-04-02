-- full recalculation for ANY kind of event
-- this is as optimized as i possibly could've made it (1 pass, no joins)
-- and in reality user got 1000 apps max; we need a 
-- a self-healing mechanism with three-stage 'pending' updates, as follows:
--  1. user adds an app
--  2. insert app, get a UUID
--  3. flush headers to user with this UUID for optimistic UI
--  4. set analytics status to 'pending' in user_analytics
--  5. try to run analytics update in the background
--      a) if success, set to 'fresh'
--      b) if fail, set to 'stale'
--  6. when user loads profile, check analytics:
--      a) if fresh, show it
--      b) if pending, show loading indicator and ping again in 5s
--      c) if stale, show warning and retry button
-- this is essentially the idea of making UX feel snappier without truly fast backend
-- it's because our demands are hard; need OLAP-style reads but OLTP-style writes...
-- well could do like read replicas but that is a ton of work and scale is not there yet 
-- ----------
-- the premature optimization strategy would be to do selective analytic updates
-- but in reality, 1000 apps is not a lot to recalculate on even with all the aggregations being done
-- also, when you have to scan the full history table anyway for ANY analytic, the cost
-- of adding more aggregations is negligible since that is the main bottleneck
-- and lowkey every 'selective' update isnt even that selective it still has to run monthly trends
-- ---------
-- the most naive way is to run all in one transaction. but this is bad for user 
-- ---------
-- yap sesh over, here's the code. a straight copy paste from previous bigquery implementation 
CREATE OR REPLACE FUNCTION service.full_recalculate_analytics(p_email TEXT)
RETURNS BOOLEAN AS $$
DECLARE
    v_status TEXT;
BEGIN
    WITH UserHistory AS (
        SELECT 
            application_id,
            email,
            event_time,
            applied_date,
            app_status,
            operation,
            (operation = 'add') AS is_application,
            (operation = 'edit' AND app_status IN ('Interviewing', 'Screen')) AS is_interview,
            (operation = 'edit' AND app_status = 'Offer') AS is_offer,
            (operation = 'edit' AND app_status IN ('Interviewing', 'Screen', 'Offer', 'Rejected', 'Ghosted')) AS is_response,
            (applied_date >= NOW() - INTERVAL '30 days') AS in_current_period,
            (applied_date >= NOW() - INTERVAL '60 days' AND applied_date < NOW() - INTERVAL '30 days') AS in_previous_period,
            TO_CHAR(applied_date, 'YYYY-MM') AS month
        FROM application_history
        WHERE email = p_email
            AND operation != 'revert'
            -- all operations in the past 365 days are relevant. technically this could be split
            -- into one 60-day lookback and one 365-day lookback but let's be real the amount of data
            -- from 60 days (prob like 50 rows) to 365 days (prob like 500 rows) is negligible
            AND applied_date >= NOW() - INTERVAL '365 days'
    ),

    -- avg time to first response
    ResponseMetrics AS (
        SELECT
            uh.application_id,
            uh.applied_date,
            MIN(CASE 
                WHEN uh.is_response AND uh.event_time > uh.applied_date
                THEN EXTRACT(EPOCH FROM (uh.event_time - uh.applied_date))/86400
                ELSE NULL
            END) AS days_to_response
        FROM UserHistory uh
        WHERE uh.is_application OR uh.is_response
        GROUP BY uh.application_id, uh.applied_date
    ),

    -- monthly trends in application status
    MonthlyTrends AS (
        SELECT
            month,
            SUM(CASE WHEN is_application THEN 1 ELSE 0 END) AS applications,
            COUNT(DISTINCT CASE WHEN is_interview THEN application_id END) AS interviews,
            COUNT(DISTINCT CASE WHEN is_offer THEN application_id END) AS offers
        FROM UserHistory
        WHERE applied_date >= NOW() - INTERVAL '365 days'
        GROUP BY month
        ORDER BY month
    ),

    Metrics AS (
        SELECT
            -- Application velocity metrics
            COALESCE(SUM(CASE WHEN in_current_period AND is_application THEN 1 ELSE 0 END), 0) AS current_30day_count,
            COALESCE(SUM(CASE WHEN in_previous_period AND is_application THEN 1 ELSE 0 END), 0) AS previous_30day_count,
            COALESCE(SUM(CASE WHEN in_current_period AND is_application THEN 1 ELSE 0 END), 0) - 
                COALESCE(SUM(CASE WHEN in_previous_period AND is_application THEN 1 ELSE 0 END), 0) AS application_velocity_trend,
            
            -- Resume effectiveness metrics
            COUNT(DISTINCT CASE WHEN event_time >= NOW() - INTERVAL '30 days' AND is_interview 
                        THEN application_id END) AS current_30day_interviews,
            COUNT(DISTINCT CASE WHEN event_time >= NOW() - INTERVAL '60 days' AND event_time < NOW() - INTERVAL '30 days' 
                        AND is_interview THEN application_id END) AS previous_30day_interviews,
            COUNT(DISTINCT CASE WHEN event_time >= NOW() - INTERVAL '30 days' AND is_interview 
                        THEN application_id END) -
            COUNT(DISTINCT CASE WHEN event_time >= NOW() - INTERVAL '60 days' AND event_time < NOW() - INTERVAL '30 days' 
                        AND is_interview THEN application_id END) AS resume_effectiveness_trend,
            
            -- Interview effectiveness metrics
            COUNT(DISTINCT CASE WHEN event_time >= NOW() - INTERVAL '30 days' AND is_offer 
                        THEN application_id END) AS current_30day_offers,
            COUNT(DISTINCT CASE WHEN event_time >= NOW() - INTERVAL '60 days' AND event_time < NOW() - INTERVAL '30 days' 
                        AND is_offer THEN application_id END) AS previous_30day_offers,
            COUNT(DISTINCT CASE WHEN event_time >= NOW() - INTERVAL '30 days' AND is_offer 
                        THEN application_id END) -
            COUNT(DISTINCT CASE WHEN event_time >= NOW() - INTERVAL '60 days' AND event_time < NOW() - INTERVAL '30 days' 
                        AND is_offer THEN application_id END) AS interview_effectiveness_trend,
            
            -- Response time metrics
            (SELECT COALESCE(AVG(CASE WHEN applied_date >= NOW() - INTERVAL '30 days' 
                            THEN days_to_response END), 0)::INT
             FROM ResponseMetrics) AS current_30day_avg_response_time,
            (SELECT COALESCE(AVG(CASE WHEN applied_date >= NOW() - INTERVAL '60 days' 
                            AND applied_date < NOW() - INTERVAL '30 days'
                            THEN days_to_response END), 0)::INT
             FROM ResponseMetrics) AS previous_30day_avg_response_time,
            (SELECT COALESCE(AVG(CASE WHEN applied_date >= NOW() - INTERVAL '30 days' 
                            THEN days_to_response END), 0)::INT -
                    COALESCE(AVG(CASE WHEN applied_date >= NOW() - INTERVAL '60 days' 
                                AND applied_date < NOW() - INTERVAL '30 days'
                                THEN days_to_response END), 0)::INT
             FROM ResponseMetrics) AS response_time_trend,
             
            -- Monthly trends 
            (SELECT jsonb_object_agg(month, jsonb_build_object(
                'applications', applications,
                'interviews', interviews,
                'offers', offers
            ))
            FROM MonthlyTrends) AS yearly_trends
        FROM UserHistory
    )
    
    -- FINALLY analytics can be updated
    UPDATE user_analytics ua
    SET 
        analytics_status = 'fresh',
        application_velocity = m.current_30day_count,
        application_velocity_trend = m.application_velocity_trend,
        resume_effectiveness = m.current_30day_interviews,
        resume_effectiveness_trend = m.resume_effectiveness_trend,
        interview_effectiveness = m.current_30day_offers,
        interview_effectiveness_trend = m.interview_effectiveness_trend,
        avg_first_response_time = m.current_30day_avg_response_time,
        avg_first_response_time_trend = m.response_time_trend,
        yearly_trends = COALESCE(m.yearly_trends, '{}'::JSONB)
    FROM Metrics m
    WHERE ua.email = p_email;

    RETURN TRUE;
EXCEPTION
    WHEN OTHERS THEN
        RAISE NOTICE 'Error in full_recalculate_analytics function: %', SQLERRM;
        RETURN FALSE;
END;
$$ LANGUAGE plpgsql;