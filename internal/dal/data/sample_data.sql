-- Password hash for "password" using bcrypt (cost 10)
INSERT INTO users (id, email, password_hash, password_salt, display_name)
VALUES (1, 'test@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'salt', 'Test User');

INSERT INTO user_settings (user_id, setting_key, setting_value)
VALUES (1, 'collect_hotels_params', '{"locations":["Riviera Maya, Mexico","Cancun, Mexico","Punta Cana, Dominican Republic"],"adults":2,"children":1,"children_ages":"2","property_types":"12","amenities":"52"}');

INSERT INTO user_settings (user_id, setting_key, setting_value)
VALUES (1, 'collect_dates', '{"dates":[{"name":"Verano Inicio","check_in":"2027-01-10","check_out":"2027-01-20"},{"name":"Verano","check_in":"2027-01-16","check_out":"2027-01-26"},{"name":"Invierno","check_in":"2027-06-26","check_out":"2027-07-06"}]}');

INSERT INTO user_settings (user_id, setting_key, setting_value)
VALUES (1, 'notification_email', 'user@example.com');

INSERT INTO collection_schedules (user_id, cron_expression, is_active)
VALUES (1, '0 15 * * *', true);
