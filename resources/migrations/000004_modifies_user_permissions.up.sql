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
