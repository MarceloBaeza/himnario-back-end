CREATE OR REPLACE FUNCTION public.sp_update_hymn(
  p_hymn_id bigint,
  p_title text,
  p_content jsonb
)
RETURNS void
LANGUAGE plpgsql
AS $$
BEGIN

  -- verificar que exista
  IF NOT EXISTS (
    SELECT 1
    FROM hymns
    WHERE id = p_hymn_id
  ) THEN
    RAISE EXCEPTION 'hymn not found' USING ERRCODE = 'P0001';
  END IF;

  -- actualizar
  UPDATE hymns
  SET
    title = p_title,
    content = p_content,
    updated_at = NOW()
  WHERE id = p_hymn_id;

END;
$$;


CREATE OR REPLACE FUNCTION public.sp_update_user_password(
  p_email text,
  p_password_hash text
)
RETURNS void
LANGUAGE plpgsql
AS $$
BEGIN

  -- verificar usuario
  IF NOT EXISTS (
    SELECT 1
    FROM users
    WHERE email = p_email
  ) THEN
    RAISE EXCEPTION 'user not found' USING ERRCODE = 'P0001';
  END IF;

  UPDATE users
  SET
    password_hash = p_password_hash,
    updated_at = NOW()
  WHERE email = p_email;

END;
$$;


-- Permisos sobre schema
GRANT USAGE ON SCHEMA public TO hymns_app;

-- Permisos sobre tablas
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE hymns TO hymns_app;

GRANT SELECT, UPDATE ON TABLE users TO hymns_app;

-- Permisos sobre secuencias (si usas bigserial)
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO hymns_app;

-- Para futuras tablas
ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO hymns_app;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT USAGE, SELECT ON SEQUENCES TO hymns_app;
