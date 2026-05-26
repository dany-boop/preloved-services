CREATE TRIGGER trigger_auth_users_updated_at
BEFORE UPDATE ON auth.users
FOR EACH ROW
EXECUTE FUNCTION auth.update_updated_at();

CREATE TRIGGER trigger_user_profiles_updated_at
BEFORE UPDATE ON users.user_profiles
FOR EACH ROW
EXECUTE FUNCTION auth.update_updated_at();