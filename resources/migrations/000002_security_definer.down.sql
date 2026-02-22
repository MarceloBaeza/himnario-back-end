ALTER PROCEDURE sp_create_user(TEXT, TEXT, TEXT, TEXT) SECURITY INVOKER;
ALTER FUNCTION  sp_get_user_auth_by_email(TEXT)        SECURITY INVOKER;
ALTER FUNCTION  sp_create_hymn_by_author_email(TEXT, JSONB, TEXT) SECURITY INVOKER;
ALTER FUNCTION  sp_list_hymns()                        SECURITY INVOKER;
ALTER FUNCTION  sp_get_hymn_by_id(BIGINT)              SECURITY INVOKER;
