-- Los SPs pasan a ejecutarse con los permisos de su owner (hymns_migrator),
-- sin importar qué usuario los llame (hymns_app u otro).
-- Esto permite que hymns_app llame los SPs sin tener acceso directo a las tablas.

ALTER PROCEDURE sp_create_user(TEXT, TEXT, TEXT, TEXT) SECURITY DEFINER;
ALTER FUNCTION  sp_get_user_auth_by_email(TEXT)        SECURITY DEFINER;
ALTER FUNCTION  sp_create_hymn_by_author_email(TEXT, JSONB, TEXT) SECURITY DEFINER;
ALTER FUNCTION  sp_list_hymns()                        SECURITY DEFINER;
ALTER FUNCTION  sp_get_hymn_by_id(BIGINT)              SECURITY DEFINER;
