# Arquitectura de código

← [Volver al índice](./architecture.md)

---

## Contenido

1. [Estructura de directorios](#1-estructura-de-directorios)
2. [Capas y responsabilidades](#2-capas-y-responsabilidades)
3. [Ciclo de vida de un request](#3-ciclo-de-vida-de-un-request)
4. [Cadena de middlewares](#4-cadena-de-middlewares)
5. [Inyección de dependencias](#5-inyección-de-dependencias)
6. [Patrones clave](#6-patrones-clave)

---

## 1. Estructura de directorios

```
.
├── cmd/
│   └── main.go                        # Punto de entrada
├── internal/
│   ├── core/
│   │   ├── domain/                    # Entidades y reglas de dominio
│   │   │   ├── hymns.go               # Entidad Hymn, helper de validación de rol
│   │   │   ├── user.go                # Entidad User, constantes UserRole
│   │   │   ├── errors.go              # Errores centinela del dominio
│   │   │   └── responses.go           # ResponseWrapper, ResponseOk, ResponseError
│   │   ├── port/
│   │   │   ├── hymns.go               # Interfaz PersistenceHymns
│   │   │   └── users.go               # Interfaz PersistenceUsers
│   │   ├── service/
│   │   │   ├── hymns.go               # HymnsService (implementa HymnUseCaseHandler)
│   │   │   └── users.go               # UserService (implementa UserUseCaseHandler)
│   │   └── usecase.go                 # Interfaces HymnUseCaseHandler, UserUseCaseHandler
│   └── infra/
│       ├── primary/
│       │   └── controllers/
│       │       ├── hymns/             # Handlers HTTP de himnos
│       │       │   ├── controller.go
│       │       │   ├── mapper.go
│       │       │   ├── request/dto.go
│       │       │   └── validations/
│       │       └── user/              # Handlers HTTP de usuarios
│       │           ├── controller.go
│       │           ├── mapper.go
│       │           ├── request/dto.go
│       │           └── validations/
│       ├── secondary/
│       │   ├── hymnary/client.go      # Repositorio PostgreSQL de himnos
│       │   └── users/
│       │       ├── client.go          # Repositorio PostgreSQL de usuarios
│       │       └── responses/user.go  # Structs para escaneo de filas DB
│       └── config/
│           ├── controller/process.go  # Motor Gin + configuración de middlewares
│           ├── database/              # Pool pgx + runner de migraciones
│           ├── instance/primary.go    # Cableado de inyección de dependencias
│           ├── property/              # Structs de configuración tipados
│           ├── security/              # Utilidades JWT + bcrypt
│           └── validation/            # Singleton de go-playground/validator
├── resources/
│   ├── migrations/                    # Archivos SQL (golang-migrate)
│   └── properties.yml                 # Configuración de servidor, DB, JWT, validaciones
├── docs/
│   ├── architecture/                  # Esta documentación
│   ├── decisions/                     # Architecture Decision Records (ADRs)
│   └── runbooks/                      # Runbooks operacionales
├── docker/
│   ├── docker-compose.yml
│   └── postgres/initdb/
├── swagger.yaml                       # Especificación OpenAPI 3.0.3
└── Dockerfile
```

---

## 2. Capas y responsabilidades

### Dominio (`internal/core/domain/`)

Structs Go puros y reglas de negocio sin dependencias externas. Es la capa más interna; nada dentro de ella importa infraestructura.

| Tipo | Descripción |
|---|---|
| `Hymn` | Entidad central del himno: `Id`, `Title`, `Content` (any/JSONB), `EmailUser`, `CreatedAt` |
| `User` | Usuario para respuestas de autenticación: `Email`, `Name`, `Role`, `Token` |
| `UserRegistry` | Payload de registro: email, contraseña, nombre, rol |
| `UserLogin` | Payload de inicio de sesión: email, contraseña |
| `UserRole` | String tipado: `admin`, `editor`, `viewer` |
| `ResponseWrapper` | Envoltorio HTTP que contiene `ResponseOk` o `ResponseError` (nunca ambos) |
| `ValidateRolCreateEditHymns` | Regla de dominio: solo `admin` y `editor` pueden crear o editar himnos |

### Puertos (`internal/core/port/`)

Interfaces Go que definen los contratos de persistencia. Los servicios dependen de ellas; los repositorios las implementan.

```go
// PersistenceHymns
AddHymn(hymn *domain.Hymn) error
GetHymnByID(id int) (*domain.Hymn, error)
GetAllHymns() ([]*domain.Hymn, error)
EditHymn(hymn *domain.Hymn) error

// PersistenceUsers
CreateUser(user *domain.UserRegistry) error
GetUserByEmail(email string) (*UserDB, error)
```

### Servicios (`internal/core/service/`)

Singletons de lógica de negocio, creados una sola vez con `sync.Once`. Delegan la persistencia al puerto inyectado y aplican las reglas de dominio antes o después.

| Servicio | Handler de caso de uso |
|---|---|
| `HymnsService` | `HymnUseCaseHandler` |
| `UserService` | `UserUseCaseHandler` |

### Adaptadores primarios — Controladores (`internal/infra/primary/controllers/`)

Cada controlador sigue un patrón consistente:

1. **Bind y validar** el DTO de request (binding JSON + validación de struct)
2. **Validar headers** (middleware de grupo, se ejecuta antes del handler)
3. **Extraer JWT** y validar el token (rutas protegidas)
4. **Verificar identidad** — los claims del token deben coincidir con `user.email` y `user.name` del body
5. **Verificar rol** — solo `admin`/`editor` para operaciones de escritura
6. **Mapear** DTO → objeto de dominio vía `mapper.go`
7. **Llamar** al handler de caso de uso
8. **Retornar** `ResponseWrapper` (éxito o error)

### Adaptadores secundarios — Repositorios (`internal/infra/secondary/`)

Repositorios PostgreSQL que implementan las interfaces de puerto usando `pgx/v5`. Todas las operaciones de base de datos van a través de stored procedures con un timeout de contexto de 5 segundos.

---

## 3. Ciclo de vida de un request

```
Request HTTP entrante
        │
        ▼
[Middleware: Logging]
        │
        ▼
[Middleware: Timeout — 30s]
        │
        ▼
[Middleware: Allowed Paths — 403 si la ruta no está listada]
        │
        ▼
[Middleware: Request Size — rechaza bodies > 1 MB]
        │
        ▼
[Middleware: Rate Limit — 100 req/min por IP]
        │
        ▼
[Middleware: Recovery — captura panics]
        │
        ▼
[Middleware: Security Headers]
        │
        ▼
[Middleware de grupo: Header Validations — x-client, country, event-id, Accept]
        │
        ▼
[Route Handler]
        ├── Bind JSON
        ├── Validación de struct
        ├── Validación JWT (rutas protegidas)
        ├── Verificación de identidad y rol (rutas de escritura)
        ├── Mapeo DTO → dominio
        ├── Llamada al servicio
        └── Escritura de ResponseWrapper como JSON
```

---

## 4. Cadena de middlewares

| Orden | Middleware | Propósito |
|---|---|---|
| 1 | Logging | Formateador personalizado; registra el body del request como JSON |
| 2 | Timeout | Límite de 30 segundos por request |
| 3 | Allowed Paths | Rechaza con 403 cualquier ruta no listada en `properties.yml` |
| 4 | Request Size | Rechaza bodies superiores a 1 MB |
| 5 | Rate Limiting | 100 req/min por IP de cliente (almacén en memoria) |
| 6 | Recovery | Convierte panics en respuesta 500 |
| 7 | Security Headers | CORS, CSP, HSTS, X-Frame-Options, XSS, X-Content-Type-Options, Referrer-Policy, Permissions-Policy |
| 8 | Header Validations | Middleware de grupo por ruta; valida los headers obligatorios |
| 9 | JWT Validation | Por ruta, aplicado dentro del handler para endpoints protegidos |

---

## 5. Inyección de dependencias

Cableado en `internal/infra/config/instance/primary.go`, siguiendo este orden:

```
Pool DB (pgx)
    │
    ├── hymnary.NewHymns()              → PersistenceHymns
    │       └── service.NewHymnsService()       → HymnUseCaseHandler
    │               └── hymns.NewHymnController()
    │
    └── users.NewClient()               → PersistenceUsers
            └── service.NewUserService()        → UserUseCaseHandler
                    └── user.NewUserController()
```

Todos los componentes usan `sync.Once` para garantizar una única instancia durante toda la vida de la aplicación.

Al finalizar el cableado, se ejecuta el seeder del usuario administrador antes de levantar el servidor.

---

## 6. Patrones clave

| Patrón | Dónde se aplica |
|---|---|
| Singleton con `sync.Once` | Servicios, pool DB, propiedades JWT, propiedades del servidor, validador |
| Stored procedures para todas las escrituras DB | `hymnary/client.go`, `users/client.go` |
| Timeout de contexto pgx (5s) | Cada query a la base de datos |
| Control de acceso basado en roles (RBAC) | Los claims del JWT llevan el rol; el controlador lo verifica antes de toda escritura |
| Verificación de identidad | `user.email` + `user.name` del body deben coincidir exactamente con los claims del JWT |
| Normalización de email | Minúsculas y trim en el mapper del DTO y en los stored procedures |
| Configuración externalizada | Todos los ajustes en `resources/properties.yml` + variables de entorno |
| Envoltorio de respuesta | Todas las respuestas HTTP envueltas en `ResponseWrapper` |
