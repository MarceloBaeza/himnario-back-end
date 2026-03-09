# Runbook — Arrancar el sistema

## Prerequisitos

| Herramienta | Versión mínima | Para qué se usa |
|---|---|---|
| Go | 1.25.5 | Compilar y ejecutar la API |
| Docker + Docker Compose | cualquier versión reciente | Base de datos PostgreSQL |
| Make | cualquier | Atajos de comandos del proyecto |
| golangci-lint | cualquier | Linting (solo para desarrollo) |

Verificar instalaciones:

```bash
go version
docker --version
docker compose version
make --version
```

---

## 1. Variables de entorno

El proyecto carga variables desde un archivo `.env` en la raíz del repositorio. El `Makefile` lo incluye automáticamente con `include .env` + `export`.

Crear el archivo `.env` (si no existe) con el siguiente contenido, ajustando los valores según el entorno:

```env
# ── Base de datos: usuario de aplicación (runtime) ──────────────────────────
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=hymns_db
DATABASE_USER=hymns_app
DATABASE_PASSWORD=password_app
DATABASE_DRIVER=pgx
DATABASE_SSL_MODE=disable

# ── Base de datos: usuario de migraciones (DDL) ──────────────────────────────
DATABASE_MIGRATION_USER=hymns_migrator
DATABASE_MIGRATION_PASSWORD=password_migrations

# ── Pool de conexiones ────────────────────────────────────────────────────────
DATABASE_MAX_OPEN_CONNECTIONS=10
DATABASE_MIN_OPEN_CONNECTIONS=1
DATABASE_MAX_CONNECTION_LIFETIME=30s
DATABASE_MAX_CONNECTION_IDLE_TIME=30s
DATABASE_HEALTH_CHECK_INTERVAL=5s
DATABASE_STATEMENTS_CACHE_CAP=256

# ── JWT ───────────────────────────────────────────────────────────────────────
JWT_SECRET_KEY=cambia_esto_en_produccion
JWT_EXPIRATION_TIME=60

# ── Seeder: usuario admin inicial ─────────────────────────────────────────────
ADMIN_INITIAL_PASSWORD=adminpass123
```

> **Nota:** Los valores de contraseña anteriores corresponden a los definidos en `docker/postgres/initdb/01-users.sql`. Si los cambias ahí, cámbialos también en `.env`.

---

## 2. Base de datos

### 2.1 Levantar PostgreSQL con Docker

```bash
docker compose -f docker/docker-compose.yml up -d
```

Esto inicia el contenedor `hymns-postgres` (PostgreSQL 16 Alpine) en el puerto `5432` con:

- Base de datos: `hymns_db`
- Timezone: `UTC` (configurado en `docker/postgres/initdb/00-settings.sql`)
- Volumen persistente: `hymns_pgdata`

Verificar que el contenedor está corriendo:

```bash
docker compose -f docker/docker-compose.yml ps
```

Esperar el health check (el contenedor ejecuta `pg_isready` cada 5 segundos):

```bash
docker compose -f docker/docker-compose.yml logs -f hymns-postgres
# Buscar: "database system is ready to accept connections"
```

### 2.2 Usuarios y roles de base de datos

Los usuarios se crean automáticamente al iniciar el contenedor vía `docker/postgres/initdb/01-users.sql`.

| Usuario | Contraseña (default Docker) | Rol | Acceso |
|---|---|---|---|
| `postgres` | _(superusuario del contenedor)_ | Superusuario | Administración total |
| `hymns_migrator` | `password_migrations` | Dueño del schema | Crea tablas, SPs, índices; ejecuta migraciones |
| `hymns_app` | `password_app` | Usuario de aplicación | Solo llama stored procedures; sin acceso directo a tablas |

**Relación de privilegios:**

- `hymns_migrator` crea todos los objetos del schema.
- Los stored procedures se ejecutan como `SECURITY DEFINER` (con permisos de `hymns_migrator`), por lo que `hymns_app` puede llamarlos sin necesitar permisos directos sobre las tablas.
- `hymns_app` tiene `SELECT/INSERT/UPDATE/DELETE` sobre `hymns` y `SELECT/UPDATE` sobre `users` (para los casos donde el acceso directo sea necesario).

### 2.3 Migraciones

Las migraciones **se ejecutan automáticamente al arrancar la API**. No es necesario correrlas manualmente.

El runner usa `golang-migrate` y lee los archivos de `resources/migrations/` con el usuario `hymns_migrator`.

| Migración | Descripción |
|---|---|
| `000001_init` | Crea todas las tablas, índices y stored procedures iniciales |
| `000002_security_definer` | Marca todos los SPs como `SECURITY DEFINER` |
| `000003_modifies_hymns_user` | Agrega `sp_update_hymn` y `sp_update_user_password`; ajusta permisos |
| `000004_modifies_user_permissions` | Refina permisos de `hymns_app` y configura `DEFAULT PRIVILEGES` |

Si necesitas conectarte directamente para verificar:

```bash
docker exec -it hymns-postgres psql -U hymns_migrator -d hymns_db
```

Verificar tablas creadas:

```sql
\dt
```

Verificar funciones/procedimientos:

```sql
\df
```

---

## 3. Arrancar la API

### Modo desarrollo (recomendado para desarrollo local)

```bash
make run
# equivale a: go run ./cmd/main.go
```

### Modo compilado

```bash
make build   # genera bin/app con -tags musl
make exec    # ejecuta ./bin/app
```

La API escucha en `http://localhost:8080` (configurable con `SERVER_PORT` si se agrega al `.env`).

---

## 4. Seeder — usuario admin inicial

Al arrancar, la API ejecuta automáticamente un seeder que crea el usuario administrador si no existe.

Los datos del admin se configuran en `resources/properties.yml` y `.env`:

| Campo | Valor (properties.yml) | Fuente |
|---|---|---|
| Email | `admin@admin.com` | hardcodeado en `properties.yml` |
| Nombre | `Admin` | hardcodeado en `properties.yml` |
| Contraseña | _(configurable)_ | variable `ADMIN_INITIAL_PASSWORD` en `.env` |
| Rol | `admin` | asignado por el seeder |

> El seeder es idempotente: si el usuario ya existe, no hace nada.

---

## 5. Usuarios de la aplicación

La API maneja tres roles de usuario. Se crean vía `POST /user/create`.

| Rol | Puede crear himnos | Puede editar himnos | Puede leer himnos |
|---|---|---|---|
| `admin` | Sí | Sí | Sí |
| `editor` | Sí | Sí | Sí |
| `viewer` | No | No | Sí |

### Crear un usuario de prueba

```bash
curl -s -X POST http://localhost:8080/user/create \
  -H "x-client: mbh" \
  -H "country: CL" \
  -H "event-id: setup-001" \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "editor@ejemplo.com",
    "password": "pass1234",
    "name": "Editor de Prueba",
    "role": "editor"
  }'
```

### Autenticar y obtener JWT

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/user/authentication \
  -H "x-client: mbh" \
  -H "country: CL" \
  -H "event-id: setup-002" \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -d '{"email":"editor@ejemplo.com","password":"pass1234"}' \
  | jq -r '.responseOk.data.token')

echo $TOKEN
```

El token expira en **60 minutos**. Pasado ese tiempo, hay que autenticarse nuevamente.

---

## 6. Verificar que el sistema está operativo

### Health check

```bash
curl -s http://localhost:8080/health \
  -H "x-client: mbh" \
  -H "country: CL" \
  -H "event-id: hc-001" \
  -H "Accept: application/json"
```

Respuesta esperada:

```json
{
  "responseOk": {
    "statusCode": 200,
    "message": "ok"
  }
}
```

### Listar himnos (sin auth)

```bash
curl -s http://localhost:8080/hymn/all \
  -H "x-client: mbh" \
  -H "country: CL" \
  -H "event-id: hc-002" \
  -H "Accept: application/json"
```

---

## 7. Headers requeridos en todas las peticiones

Todos los endpoints requieren estos headers. Sin ellos la API retorna `400`.

| Header | Valor válido |
|---|---|
| `x-client` | `mbh` o `front-end-himnary` |
| `country` | `CL` |
| `event-id` | cualquier string no vacío |
| `Accept` | `application/json` o `*/*` |
| `Content-Type` | `application/json` (solo en POST/PUT) |

---

## 8. Detener el sistema

Detener solo la base de datos:

```bash
docker compose -f docker/docker-compose.yml stop
```

Detener y eliminar el contenedor (los datos persisten en el volumen `hymns_pgdata`):

```bash
docker compose -f docker/docker-compose.yml down
```

Eliminar también el volumen (borra todos los datos):

```bash
docker compose -f docker/docker-compose.yml down -v
```

---

## 9. Comandos útiles de desarrollo

```bash
make test          # Tests unitarios + reporte de cobertura HTML
make race-test     # Tests con detección de race conditions
make lint          # Linting completo con golangci-lint
make lint-fast     # Lint rápido (para pre-commit)
make check-all     # test + race-test + lint
make pre-commit    # clean + lint + build + test (antes de cada commit)
make clean-mod     # Limpia caché de módulos Go
```

---

## 10. Problemas comunes

| Síntoma | Causa probable | Solución |
|---|---|---|
| `dial tcp: connect: connection refused` al arrancar | PostgreSQL no está listo | Esperar el health check del contenedor; revisar `docker compose ps` |
| `password authentication failed for user "hymns_migrator"` | Credenciales en `.env` no coinciden con las del contenedor | Verificar `DATABASE_MIGRATION_PASSWORD` en `.env` vs `01-users.sql` |
| `no migration files found` | La API se inicia desde un directorio incorrecto | Ejecutar desde la raíz del repositorio |
| `401 Unauthorized` en endpoints protegidos | Token expirado o no enviado | Reautenticarse y usar el nuevo token |
| `403 Forbidden` en un endpoint | Ruta no está en `allow-paths` de `properties.yml` | Agregar la ruta a la lista `server.allow-paths` |
| `429 Too Many Requests` | Rate limit excedido (100 req/min por IP) | Esperar un minuto o ajustar `limit-by-ip` en `properties.yml` |
