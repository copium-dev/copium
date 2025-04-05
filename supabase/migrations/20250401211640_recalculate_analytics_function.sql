-- full recalculation for ANY kind of event
-- this is as optimized as i possibly could've made it (1 pass, no joins)
-- and in the most generous reality ever a user has 5000 historical events (this ain't shi....)
-- PREPARE FOR MASSIVE YAP SESH ABOUT WHY THIS WORKS. THIS COMES FROM THE FLAWS OF THE BIGQUERY IMPLEMENTATION
-- THAT WAS IN PROD FOR 2 WEEKS. TALK TO ME (SEAN) IF U HAVE SUGGESTIONS TO MAKE THIS BETTER
-- ----------
-- here's an example of async processing replicated all within postgres using triggers and pg_cron for cleanup
--  1. user adds an app or makes a status change
--  2. insert app into latest, add to history, get a UUID
--  3. increment basic counters in user_analytics (e.g. app count, interview count)
--     gotta be done in this transaction in case add (or edit) rollback; dont wanna manage consistency app-side
--  4. set analytics status to 'pending' in user_analytics
--  5. an AFTER trigger will push the UUID to a queue table
--     after this point, the app remains responsive. Since the trigger isn't analytics recalculation
--     and rather a simple insert, this is essentially the same as waiting for pushing to RabbitMQ
--  6. an AFTER trigger on the queue table will call full_recalculate_analytics on the UUID
--     a) succeeds, status is set to 'fresh'
--     b) fails, status is set to 'error'
--  7. pg_cron will run a cleanup job at 3AM every day to remove any entries older than 3 days
--     this is a safety net in case the queue table gets too big, does a batch delete in off-peak hours
--     to ensure speed during peak hours
--  8. user loads the profile page. 
--     a) if analytics status is 'pending', show a loading spinner in the last updated area, and have frontend re-poll every 5s
--     b) if analytics status is 'fresh', show the analytics and last updated
--     c) if analytics status is 'error', show a retry button that calls this function again
-- this is essentially the idea of making UX feel snappier without truly fast analytics
-- it's because our demands are hard; need OLAP-style reads but OLTP-style writes...
-- well could do like read replicas but that is a ton of work and scale is not there yet
-- but in general this is super minimal extra code but tons of good UX gains. literally a complete replica
-- of the old bigquery implementation but now all in postgres. postgres is king
-- ----------
-- the premature optimization strategy would be to do selective analytic updates
-- but in reality, 5000 rows is not a lot to recalculate on even with all the aggregations being done
-- also, when you have to scan the full history table anyway for ANY analytic, the cost
-- of adding more aggregations is negligible since the main bottleneck is the scan
-- and lowkey every 'selective' update isnt even that selective it still has to run monthly trends
-- plus it's annoying as hell to manage all the different cases and makes adding new analytics hard
-- ----------
-- the most naive way is to run all in one transaction. but this is bad for UX once app scales and we have network hops and stuff
-- ----------
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
        FROM service.application_history
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
            
            -- Resume effectiveness metrics
            COUNT(DISTINCT CASE WHEN event_time >= NOW() - INTERVAL '30 days' AND is_interview 
                        THEN application_id END) AS current_30day_interviews,
            COUNT(DISTINCT CASE WHEN event_time >= NOW() - INTERVAL '60 days' AND event_time < NOW() - INTERVAL '30 days' 
                        AND is_interview THEN application_id END) AS previous_30day_interviews,
            
            -- Interview effectiveness metrics
            COUNT(DISTINCT CASE WHEN event_time >= NOW() - INTERVAL '30 days' AND is_offer 
                        THEN application_id END) AS current_30day_offers,
            COUNT(DISTINCT CASE WHEN event_time >= NOW() - INTERVAL '60 days' AND event_time < NOW() - INTERVAL '30 days' 
                        AND is_offer THEN application_id END) AS previous_30day_offers,
            
            -- Response time metrics
            (SELECT AVG(CASE WHEN applied_date >= NOW() - INTERVAL '30 days' 
                            THEN days_to_response ELSE NULL END)
            FROM ResponseMetrics)::INT AS current_30day_avg_response_time,
                        
            (SELECT AVG(CASE WHEN applied_date >= NOW() - INTERVAL '60 days' 
                            AND applied_date < NOW() - INTERVAL '30 days'
                            THEN days_to_response ELSE NULL END)
            FROM ResponseMetrics)::INT AS previous_30day_avg_response_time,

             
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
    UPDATE service.user_analytics ua
    SET 
        analytics_status = 'fresh',
        application_velocity = m.current_30day_count,
        application_velocity_trend = m.current_30day_count - m.previous_30day_count,
        resume_effectiveness = m.current_30day_interviews,
        resume_effectiveness_trend = m.current_30day_interviews - m.previous_30day_interviews,
        interview_effectiveness = m.current_30day_offers,
        interview_effectiveness_trend = m.current_30day_offers - m.previous_30day_offers,
        avg_first_response_time = m.current_30day_avg_response_time,
        prev_avg_first_response_time = m.previous_30day_avg_response_time,
        avg_first_response_time_trend = m.current_30day_avg_response_time - m.previous_30day_avg_response_time,
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