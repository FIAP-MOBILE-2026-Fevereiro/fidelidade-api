DROP TABLE IF EXISTS reward_redemptions;
DROP TABLE IF EXISTS stamps;
DROP TABLE IF EXISTS qr_codes;
DROP TABLE IF EXISTS programs;
DROP TRIGGER IF EXISTS trg_users_updated_at ON users;
DROP TABLE IF EXISTS users;
DROP FUNCTION IF EXISTS set_updated_at();
