# Arquitectura de base de datos

← [Volver al índice](./architecture.md)

---

## Contenido

1. [Tablas](#1-tablas)
2. [Relaciones](#2-relaciones)
3. [Índices](#3-índices)
4. [Stored Procedures](#4-stored-procedures)
5. [Roles y permisos](#5-roles-y-permisos)
6. [Historial de migraciones](#6-historial-de-migraciones)

---

## 1. Tablas

### `users`

Almacena las cuentas de usuario registradas.

| Columna | Tipo | Restricciones |
|---|---|---|
| `id` | `BIGSERIAL` | Clave primaria |
| `email` | `VARCHAR(255)` | Único, no nulo |
| `password_hash` | `VARCHAR(255)` | No nulo (bcrypt, cost=12) |
| `name` | `VARCHAR(100)` | Nullable |
| `role` | `VARCHAR(20)` | No nulo, default `'viewer'` |
| `created_at` | `TIMESTAMP` | No nulo, default `NOW()` |
| `updated_at` | `TIMESTAMP` | Nullable, se actualiza en cada modificación |

Roles permitidos: `admin`, `editor`, `viewer`.

---

### `hymns`

Tabla principal de contenido. El campo `content` se almacena como JSONB para flexibilidad de esquema.

| Columna | Tipo | Restricciones |
|---|---|---|
| `id` | `BIGSERIAL` | Clave primaria |
| `title` | `VARCHAR(255)` | No nulo |
| `content` | `JSONB` | No nulo |
| `author` | `VARCHAR(255)` | Nullable (nombre libre del autor histórico) |
| `created_by` | `BIGINT` | FK → `users(id)` ON DELETE SET NULL |
| `created_at` | `TIMESTAMP` | No nulo, default `NOW()` |
| `updated_at` | `TIMESTAMP` | Nullable, se actualiza en cada modificación |

La unicidad del título se valida de forma case-insensitive dentro de `sp_create_hymn_by_author_email` (no existe un constraint UNIQUE a nivel de tabla; el control está en el stored procedure).

---

### `categories`

Taxonomía de clasificación de himnos.

| Columna | Tipo | Restricciones |
|---|---|---|
| `id` | `BIGSERIAL` | Clave primaria |
| `name` | `VARCHAR(100)` | No nulo |
| `slug` | `VARCHAR(100)` | Único, no nulo |

---

### `hymn_categories`

Relación muchos-a-muchos entre himnos y categorías.

| Columna | Tipo | Restricciones |
|---|---|---|
| `hymn_id` | `BIGINT` | FK → `hymns(id)` ON DELETE CASCADE |
| `category_id` | `BIGINT` | FK → `categories(id)` ON DELETE CASCADE |

Clave primaria compuesta: `(hymn_id, category_id)`.

---

### `favorites`

Registra qué himnos ha marcado cada usuario como favorito.

| Columna | Tipo | Restricciones |
|---|---|---|
| `user_id` | `BIGINT` | FK → `users(id)` ON DELETE CASCADE |
| `hymn_id` | `BIGINT` | FK → `hymns(id)` ON DELETE CASCADE |
| `created_at` | `TIMESTAMP` | No nulo, default `NOW()` |

Clave primaria compuesta: `(user_id, hymn_id)`.

---

### `setlists`

Colecciones ordenadas de himnos para eventos.

| Columna | Tipo | Restricciones |
|---|---|---|
| `id` | `BIGSERIAL` | Clave primaria |
| `name` | `VARCHAR(255)` | No nulo |
| `date` | `DATE` | No nulo |
| `time` | `TIME` | No nulo |
| `created_by` | `BIGINT` | FK → `users(id)` ON DELETE SET NULL |
| `created_at` | `TIMESTAMP` | No nulo, default `NOW()` |

Restricción única: `(date, time)` — un solo setlist por franja horaria de evento.

---

### `setlist_hymns`

Membresía ordenada de himnos dentro de un setlist.

| Columna | Tipo | Restricciones |
|---|---|---|
| `setlist_id` | `BIGINT` | FK → `setlists(id)` ON DELETE CASCADE |
| `hymn_id` | `BIGINT` | FK → `hymns(id)` ON DELETE CASCADE |
| `position` | `INTEGER` | No nulo (orden dentro del setlist) |

Clave primaria compuesta: `(setlist_id, hymn_id)`.

---

## 2. Relaciones

```
users ──< hymns            (created_by — SET NULL al eliminar usuario)
users ──< setlists         (created_by — SET NULL al eliminar usuario)
users ──< favorites >── hymns
hymns ──< hymn_categories >── categories
hymns ──< setlist_hymns >── setlists
```

Al eliminar un usuario, los himnos y setlists que creó se conservan con `created_by = NULL`.
Al eliminar un himno o categoría, los registros relacionados en tablas intermedias se eliminan en cascada.

---

## 3. Índices

| Índice | Tabla | Columna(s) | Tipo | Propósito |
|---|---|---|---|---|
| `idx_hymns_title` | `hymns` | `title` | B-tree | Búsqueda rápida y ordenamiento por título |
| `idx_hymns_content_gin` | `hymns` | `content` | GIN | Consultas JSON path y búsqueda en el contenido JSONB |

---

## 4. Stored Procedures

Todas las operaciones de la aplicación sobre la base de datos se realizan a través de stored procedures. El usuario de aplicación (`hymns_app`) no tiene acceso directo a las tablas más allá de lo que se le otorgó explícitamente; las escrituras se ejecutan siempre vía rutinas `SECURITY DEFINER` propias del usuario de migraciones (`hymns_migrator`).

---

### `sp_create_user(p_email, p_password_hash, p_name, p_role)`

- **Tipo:** `PROCEDURE`
- Normaliza el email (minúsculas + trim).
- Inserta en `users`.
- Lanza `23505` (violación de unicidad) si el email ya existe.

---

### `sp_get_user_auth_by_email(p_email)`

- **Tipo:** `FUNCTION` → `TABLE(id, email, password_hash, role, name)`
- Obtiene el registro de usuario para el flujo de login.
- Normaliza el email en la búsqueda.

---

### `sp_create_hymn_by_author_email(p_title, p_content, p_author_email)`

- **Tipo:** `FUNCTION` → `BIGINT` (ID del nuevo himno)
- Valida que título, contenido y email del autor no estén vacíos.
- Normaliza el email del autor; resuelve `users.id` a partir del email.
- Lanza `P0001` si el autor no existe.
- Verifica unicidad del título (case-insensitive); lanza `23505` si está duplicado.
- Inserta en `hymns` con `created_by = v_user_id`.

---

### `sp_list_hymns()`

- **Tipo:** `FUNCTION` → `TABLE(id, title, created_by, created_at, updated_at)`
- Retorna todos los himnos ordenados por `title ASC`.
- No incluye `content` (listado liviano para la vista de índice).

---

### `sp_get_hymn_by_id(p_hymn_id)`

- **Tipo:** `FUNCTION` → `TABLE(id, title, content, created_at)`
- Retorna el himno completo incluyendo el JSONB de `content`.
- Retorna conjunto de filas vacío si el ID no existe — la capa de aplicación lo traduce a HTTP 404.

---

### `sp_update_hymn(p_hymn_id, p_title, p_content)`

- **Tipo:** `FUNCTION` → `void`
- Verifica que el himno exista; lanza `P0001` si no (mapeado a HTTP 404).
- Actualiza `title`, `content` y `updated_at = NOW()`.

---

### `sp_update_user_password(p_email, p_password_hash)`

- **Tipo:** `FUNCTION` → `void`
- Verifica que el usuario exista por email; lanza `P0001` si no.
- Actualiza `password_hash` y `updated_at = NOW()`.

---

## 5. Roles y permisos

| Rol | Propósito |
|---|---|
| `hymns_migrator` | Dueño de todos los objetos del schema; ejecuta migraciones |
| `hymns_app` | Usuario de aplicación en runtime; permisos mínimos |

**Permisos de `hymns_app`:**

- `USAGE` en el schema `public`
- `SELECT, INSERT, UPDATE, DELETE` sobre la tabla `hymns`
- `SELECT, UPDATE` sobre la tabla `users`
- `USAGE, SELECT` sobre todas las secuencias (para lecturas de BIGSERIAL)
- Las tablas futuras heredan los mismos permisos vía `ALTER DEFAULT PRIVILEGES`

**Modelo SECURITY DEFINER:**

Todos los stored procedures se ejecutan con los privilegios de `hymns_migrator` (su dueño), independientemente de qué usuario los llame. Esto permite que `hymns_app` invoque los procedures sin necesitar acceso directo elevado a las tablas.

```
hymns_app  →  llama SP  →  SP corre como hymns_migrator  →  accede a tabla
```

---

## 6. Historial de migraciones

Las migraciones se aplican automáticamente al arrancar la API mediante `golang-migrate`, leyendo los archivos de `resources/migrations/` con el usuario `hymns_migrator`.

| Versión | Archivo | Resumen |
|---|---|---|
| `000001` | `init` | Crea todas las tablas (`users`, `hymns`, `categories`, `hymn_categories`, `favorites`, `setlists`, `setlist_hymns`), los índices y los stored procedures iniciales (`sp_create_user`, `sp_get_user_auth_by_email`, `sp_create_hymn_by_author_email`, `sp_list_hymns`, `sp_get_hymn_by_id`) |
| `000002` | `security_definer` | Marca todos los SPs existentes como `SECURITY DEFINER` para que se ejecuten con privilegios del dueño |
| `000003` | `modifies_hymns_user` | Agrega `sp_update_hymn` y `sp_update_user_password`; otorga permisos explícitos a `hymns_app` sobre tablas y secuencias |
| `000004` | `modifies_user_permissions` | Refina los permisos de `hymns_app`; configura `DEFAULT PRIVILEGES` para tablas y secuencias futuras |
