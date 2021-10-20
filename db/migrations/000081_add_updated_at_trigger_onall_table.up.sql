--- bank_accounts
CREATE TRIGGER bank_accounts_updated_at_set_timestamp
BEFORE UPDATE ON bank_accounts
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

--- block_list
CREATE TRIGGER block_list_updated_at_set_timestamp
BEFORE UPDATE ON block_list
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

--- chatroom_users
CREATE TRIGGER chatroom_users_updated_at_set_timestamp
BEFORE UPDATE ON chatroom_users
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

--- chatrooms
CREATE TRIGGER chatrooms_updated_at_set_timestamp
BEFORE UPDATE ON chatrooms
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

--- coin_orders 
CREATE TRIGGER coin_orders_updated_at_set_timestamp
BEFORE UPDATE ON coin_orders
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

--- images 
CREATE TRIGGER images_updated_at_set_timestamp
BEFORE UPDATE ON images
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

--- payments 
CREATE TRIGGER payments_updated_at_set_timestamp
BEFORE UPDATE ON payments
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

--- service_inquiries 
CREATE TRIGGER service_inquiries_updated_at_set_timestamp
BEFORE UPDATE ON service_inquiries 
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

--- service_options 
CREATE TRIGGER service_options_updated_at_set_timestamp
BEFORE UPDATE ON service_options 
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

--- service_qrcode 
CREATE TRIGGER service_qrcode_updated_at_set_timestamp
BEFORE UPDATE ON service_qrcode 
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

--- service_ratings 
CREATE TRIGGER service_ratings_updated_at_set_timestamp
BEFORE UPDATE ON service_ratings 
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

--- services 
CREATE TRIGGER services_updated_at_set_timestamp
BEFORE UPDATE ON services 
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

--- user_balance 
CREATE TRIGGER user_balance_updated_at_set_timestamp
BEFORE UPDATE ON user_balance 
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

--- user_refcodes 
CREATE TRIGGER user_refcodes_updated_at_set_timestamp
BEFORE UPDATE ON user_refcodes 
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

--- user_service_options 
CREATE TRIGGER user_service_options_updated_at_set_timestamp
BEFORE UPDATE ON user_service_options 
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

--- users
CREATE TRIGGER users_updated_at_set_timestamp
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();