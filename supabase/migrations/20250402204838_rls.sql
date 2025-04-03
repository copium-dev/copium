-- -- all requests are authorized backend so no real need for RLS but why not 
-- -- note that USING is for reads and WITH CHECK is for writes
-- ALTER TABLE service.users ENABLE ROW LEVEL SECURITY;
-- ALTER TABLE service.user_analytics ENABLE ROW LEVEL SECURITY;
-- ALTER TABLE service.user_applications ENABLE ROW LEVEL SECURITY;
-- ALTER TABLE service.application_history ENABLE ROW LEVEL SECURITY;

-- -- RLS policies for the users table
-- CREATE POLICY "users_select_policy" 
--     ON service.users
--     FOR SELECT 
--     USING (email = auth.email());

-- CREATE POLICY "users_insert_policy" 
--     ON service.users
--     FOR INSERT 
--     WITH CHECK (email = auth.email());

-- CREATE POLICY "users_update_policy" 
--     ON service.users
--     FOR UPDATE 
--     USING (email = auth.email());

-- CREATE POLICY "users_delete_policy" 
--     ON service.users
--     FOR DELETE 
--     USING (email = auth.email());

-- -- RLS policies for the user_analytics table
-- CREATE POLICY "user_analytics_select_policy" 
--     ON service.user_analytics
--     FOR SELECT 
--     USING (email = auth.email());

-- CREATE POLICY "user_analytics_insert_policy" 
--     ON service.user_analytics
--     FOR INSERT 
--     WITH CHECK (email = auth.email());

-- CREATE POLICY "user_analytics_update_policy" 
--     ON service.user_analytics
--     FOR UPDATE 
--     USING (email = auth.email());

-- CREATE POLICY "user_analytics_delete_policy" 
--     ON service.user_analytics
--     FOR DELETE 
--     USING (email = auth.email());

-- -- RLS policies for the user_applications table
-- CREATE POLICY "user_applications_select_policy" 
--     ON service.user_applications
--     FOR SELECT 
--     USING (email = auth.email());

-- CREATE POLICY "user_applications_insert_policy" 
--     ON service.user_applications
--     FOR INSERT 
--     WITH CHECK (email = auth.email());

-- CREATE POLICY "user_applications_update_policy" 
--     ON service.user_applications
--     FOR UPDATE 
--     USING (email = auth.email());

-- CREATE POLICY "user_applications_delete_policy" 
--     ON service.user_applications
--     FOR DELETE 
--     USING (email = auth.email());

-- -- RLS policies for the application_history table
-- CREATE POLICY "application_history_select_policy" 
--     ON service.application_history
--     FOR SELECT 
--     USING (email = auth.email());

-- CREATE POLICY "application_history_insert_policy" 
--     ON service.application_history
--     FOR INSERT 
--     WITH CHECK (email = auth.email());

-- CREATE POLICY "application_history_update_policy" 
--     ON service.application_history
--     FOR UPDATE 
--     USING (email = auth.email());

-- CREATE POLICY "application_history_delete_policy" 
--     ON service.application_history
--     FOR DELETE 
--     USING (email = auth.email());