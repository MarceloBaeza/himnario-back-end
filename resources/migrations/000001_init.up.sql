-- USERS
CREATE TABLE IF NOT EXISTS users (
  id            BIGSERIAL PRIMARY KEY,
  email         VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  name          VARCHAR(100),
  role          VARCHAR(20) NOT NULL DEFAULT 'viewer',
  created_at    TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMP
);

-- HYMNS
CREATE TABLE IF NOT EXISTS hymns (
  id         BIGSERIAL PRIMARY KEY,
  title      VARCHAR(255) NOT NULL,
  content    JSONB NOT NULL,
  author     VARCHAR(255),
  created_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP
);

-- CATEGORIES
CREATE TABLE IF NOT EXISTS categories (
  id   BIGSERIAL PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  slug VARCHAR(100) UNIQUE NOT NULL
);

-- MANY-TO-MANY
CREATE TABLE IF NOT EXISTS hymn_categories (
  hymn_id     BIGINT NOT NULL REFERENCES hymns(id) ON DELETE CASCADE,
  category_id BIGINT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
  PRIMARY KEY (hymn_id, category_id)
);

-- FAVORITES
CREATE TABLE IF NOT EXISTS favorites (
  user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  hymn_id    BIGINT NOT NULL REFERENCES hymns(id) ON DELETE CASCADE,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  PRIMARY KEY (user_id, hymn_id)
);

-- SETLISTS
CREATE TABLE IF NOT EXISTS setlists (
  id         BIGSERIAL PRIMARY KEY,
  name       VARCHAR(255) NOT NULL,
  date       DATE NOT NULL,
  time       TIME NOT NULL,
  created_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  UNIQUE (date, time)
);

CREATE TABLE IF NOT EXISTS setlist_hymns (
  setlist_id BIGINT NOT NULL REFERENCES setlists(id) ON DELETE CASCADE,
  hymn_id    BIGINT NOT NULL REFERENCES hymns(id) ON DELETE CASCADE,
  position   INTEGER NOT NULL,
  PRIMARY KEY (setlist_id, hymn_id)
);

-- Índices útiles
CREATE INDEX IF NOT EXISTS idx_hymns_title ON hymns(title);
CREATE INDEX IF NOT EXISTS idx_hymns_content_gin ON hymns USING GIN (content);


---SPS
CREATE OR REPLACE PROCEDURE sp_create_user(
  p_email TEXT,
  p_password_hash TEXT,
  p_name TEXT DEFAULT NULL,
  p_role TEXT DEFAULT 'viewer'
)
LANGUAGE plpgsql
AS $$
BEGIN
  p_email := lower(trim(p_email));

  INSERT INTO users (email, password_hash, name, role)
  VALUES (p_email, p_password_hash, p_name, p_role);

EXCEPTION
  WHEN unique_violation THEN
    RAISE EXCEPTION 'email already exists' USING ERRCODE = '23505';
END;
$$;

CREATE OR REPLACE FUNCTION sp_get_user_auth_by_email(p_email TEXT)
RETURNS TABLE (
  id BIGINT,
  email TEXT,
  password_hash TEXT,
  role TEXT,
  name TEXT
)
LANGUAGE sql
AS $$
  SELECT u.id, u.email, u.password_hash, u.role, u.name
  FROM users u
  WHERE u.email = lower(trim(p_email))
  LIMIT 1;
$$;

--SP para crear himno dado email del autor

CREATE OR REPLACE FUNCTION public.sp_create_hymn_by_author_email(
  p_title text,
  p_content jsonb,
  p_author_email text
)
RETURNS bigint
LANGUAGE plpgsql
AS $function$
DECLARE
  v_user_id bigint;
  v_hymn_id bigint;
  v_norm_title text;
BEGIN
  -- validar título
  p_title := trim(p_title);

  IF p_title IS NULL OR p_title = '' THEN
    RAISE EXCEPTION 'title is required' USING ERRCODE = '22023';
  END IF;

  IF p_content IS NULL THEN
    RAISE EXCEPTION 'content is required' USING ERRCODE = '22023';
  END IF;

  IF p_author_email IS NULL OR trim(p_author_email) = '' THEN
    RAISE EXCEPTION 'author email is required' USING ERRCODE = '22023';
  END IF;

  -- normalizar email
  p_author_email := lower(trim(p_author_email));

  -- buscar usuario
  SELECT id
    INTO v_user_id
  FROM users
  WHERE email = p_author_email
  LIMIT 1;

  IF v_user_id IS NULL THEN
    RAISE EXCEPTION 'author not found' USING ERRCODE = 'P0001';
  END IF;

  -- normalizar título para comparación (solo minúsculas + trim)
  v_norm_title := lower(p_title);

  -- validar duplicidad por título (case-insensitive)
  IF EXISTS (
    SELECT 1
    FROM hymns h
    WHERE lower(trim(h.title)) = v_norm_title
    LIMIT 1
  ) THEN
    RAISE EXCEPTION 'hymn title already exists' USING ERRCODE = '23505';
  END IF;

  -- insertar himno
  INSERT INTO hymns (
    title,
    content,
    created_by,
    created_at,
    updated_at
  )
  VALUES (
    p_title,
    p_content,
    v_user_id,
    NOW(),
    NOW()
  )
  RETURNING id INTO v_hymn_id;

  RETURN v_hymn_id;
END;
$function$;

-- get all hymns
CREATE OR REPLACE FUNCTION public.sp_list_hymns()
RETURNS TABLE (
  id bigint,
  title text,
  created_by bigint,
  created_at timestamp,
  updated_at timestamp
)
LANGUAGE sql
AS $$
  SELECT
    h.id,
    h.title,
    h.created_by,
    h.created_at,
    h.updated_at
  FROM hymns h
  ORDER BY h.title ASC;
$$;


CREATE OR REPLACE FUNCTION public.sp_get_hymn_by_id(
  p_hymn_id bigint
)
RETURNS TABLE (
  id bigint,
  title text,
  content jsonb,
  created_at timestamp
)
LANGUAGE sql
AS $$
  SELECT
    h.id,
    h.title,
    h.content,
    h.created_at
  FROM hymns h
  WHERE h.id = p_hymn_id
  LIMIT 1;
$$;
