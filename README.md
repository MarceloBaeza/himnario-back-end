# Himnario API

REST API para gestión de himnos. Construida en Go con Gin.

---

## Base URL

```
http://<host>:<port>
```

Puerto por defecto: `8080` (configurable via variable de entorno `SERVER_PORT`).

---

## Headers requeridos

Todos los endpoints exigen los siguientes headers:

| Header | Valor requerido | Notas |
|---|---|---|
| `x-client` | `mbh` o `front-end-himnary` | Identifica al cliente |
| `country` | `CL` | Único valor aceptado |
| `event-id` | cualquier string no vacío | Identificador de la solicitud |
| `Accept` | `application/json` o `*/*` | |
| `Content-Type` | `application/json` | Solo en requests `POST` |

---

## Formato de respuesta

Todas las respuestas siguen el mismo envoltorio:

**Éxito**
```json
{
  "responseOk": {
    "statusCode": 200,
    "message": "mensaje descriptivo",
    "data": { }
  }
}
```

**Error**
```json
{
  "responseError": {
    "statusCode": 400,
    "error": "mensaje descriptivo",
    "data": {
      "campo": "descripción del error"
    }
  }
}
```

---

## Autenticación

Los endpoints protegidos requieren un JWT en el header:

```
Authorization: Bearer <token>
```

El token se obtiene desde `POST /user/authentication`. Expira en **60 minutos**.

---

## Endpoints

### Health Check

```
GET /health
```

Verifica que el servidor está activo.

**Respuesta `200`**
```json
{
  "responseOk": {
    "statusCode": 200,
    "message": "ok"
  }
}
```

---

### Registrar usuario

```
POST /user/create
```

Crea una nueva cuenta de usuario.

**Body**
```json
{
  "email": "usuario@ejemplo.com",
  "password": "minimo8caracteres",
  "name": "Nombre Apellido",
  "role": "viewer"
}
```

| Campo | Tipo | Requerido | Validaciones |
|---|---|---|---|
| `email` | string | sí | formato email válido |
| `password` | string | sí | mínimo 8 caracteres |
| `name` | string | no | — |
| `role` | string | sí | `admin`, `editor` o `viewer` |

**Respuesta `200`**
```json
{
  "responseOk": {
    "statusCode": 200,
    "message": "User created successfully"
  }
}
```

**Respuesta `409` — email ya registrado**
```json
{
  "responseError": {
    "statusCode": 409,
    "error": "User creation failed",
    "data": {
      "email": "email already exists"
    }
  }
}
```

---

### Autenticar usuario

```
POST /user/authentication
```

Retorna un JWT y los datos del usuario autenticado.

**Body**
```json
{
  "email": "usuario@ejemplo.com",
  "password": "minimo8caracteres"
}
```

**Respuesta `200`**
```json
{
  "responseOk": {
    "statusCode": 200,
    "message": "Authentication successful",
    "data": {
      "email": "usuario@ejemplo.com",
      "name": "Nombre Apellido",
      "role": "viewer",
      "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    }
  }
}
```

**Respuesta `401` — credenciales inválidas**
```json
{
  "responseError": {
    "statusCode": 401,
    "error": "Authentication failed",
    "data": {
      "authentication": "invalid credentials"
    }
  }
}
```

---

### Crear himno

```
POST /hymn/create
```

> **Requiere JWT.** Solo usuarios con rol `admin` o `editor`.
> El email y nombre del body deben coincidir exactamente con los del token.

**Headers adicionales**
```
Authorization: Bearer <token>
```

**Body**
```json
{
  "title": "Título del himno",
  "content": {
    "verses": [
      { "number": 1, "lines": ["Línea uno", "Línea dos"] }
    ],
    "chorus": "Estribillo del himno"
  },
  "user": {
    "email": "autor@ejemplo.com",
    "name": "Nombre Autor"
  }
}
```

| Campo | Tipo | Requerido | Notas |
|---|---|---|---|
| `title` | string | sí | debe ser único (case-insensitive) |
| `content` | objeto JSON | sí | estructura libre (se almacena como JSONB) |
| `user.email` | string | sí | debe coincidir con el token |
| `user.name` | string | sí | debe coincidir con el token |

**Respuesta `200`**
```json
{
  "responseOk": {
    "statusCode": 200,
    "message": "Hymn created successfully"
  }
}
```

**Respuesta `401` — token inválido, expirado o datos de usuario no coinciden**
```json
{
  "responseError": {
    "statusCode": 401,
    "error": "Unauthorized",
    "data": {
      "error": "invalid or expired token"
    }
  }
}
```

---

### Listar todos los himnos

```
GET /hymn/all
```

Retorna todos los himnos ordenados por título. No incluye el contenido.

**Respuesta `200`**
```json
{
  "responseOk": {
    "statusCode": 200,
    "message": "Get all hymns successfully",
    "data": [
      {
        "id": 1,
        "title": "Himno de ejemplo",
        "created_at": "2025-01-15T10:30:00Z"
      }
    ]
  }
}
```

---

### Obtener himno por ID

```
GET /hymn/{id}
```

Retorna un himno completo incluyendo su contenido.

**Path params**

| Parámetro | Tipo | Descripción |
|---|---|---|
| `id` | integer | ID del himno |

**Respuesta `200`**
```json
{
  "responseOk": {
    "statusCode": 200,
    "message": "Get hymn successfully",
    "data": {
      "id": 1,
      "title": "Himno de ejemplo",
      "content": {
        "verses": [
          { "number": 1, "lines": ["Línea uno", "Línea dos"] }
        ],
        "chorus": "Estribillo del himno"
      },
      "created_at": "2025-01-15T10:30:00Z"
    }
  }
}
```

**Respuesta `404` — himno no existe**
```json
{
  "responseError": {
    "statusCode": 404,
    "error": "Hymn not found",
    "data": {
      "error": "hymn with given ID does not exist"
    }
  }
}
```

**Respuesta `400` — ID no es un número entero**
```json
{
  "responseError": {
    "statusCode": 400,
    "error": "Invalid hymn ID",
    "data": {
      "error": "strconv.Atoi: parsing \"abc\": invalid syntax"
    }
  }
}
```

---

## Códigos de error comunes

| Código | Descripción |
|---|---|
| `400` | Body o headers inválidos |
| `401` | Token ausente, expirado o datos de usuario no coinciden |
| `403` | Ruta no permitida |
| `404` | Recurso no encontrado |
| `409` | Conflicto (ej. email ya registrado) |
| `429` | Límite de solicitudes excedido (100 req/min por IP) |
| `500` | Error interno del servidor |

---

## Límites

| Restricción | Valor |
|---|---|
| Tamaño máximo del body | 1 MB |
| Tiempo máximo de respuesta | 30 segundos |
| Rate limit | 100 solicitudes/minuto por IP |

---

## Ejemplo de flujo completo (curl)

```bash
BASE="http://localhost:8080"

# Headers comunes
HEADERS=(
  -H "x-client: mbh"
  -H "country: CL"
  -H "event-id: test-001"
  -H "Accept: application/json"
)

# 1. Registrar usuario
curl -s -X POST "$BASE/user/create" \
  "${HEADERS[@]}" \
  -H "Content-Type: application/json" \
  -d '{"email":"editor@ejemplo.com","password":"pass1234","name":"Editor","role":"editor"}'

# 2. Autenticar y guardar token
TOKEN=$(curl -s -X POST "$BASE/user/authentication" \
  "${HEADERS[@]}" \
  -H "Content-Type: application/json" \
  -d '{"email":"editor@ejemplo.com","password":"pass1234"}' \
  | jq -r '.responseOk.data.token')

# 3. Crear himno
curl -s -X POST "$BASE/hymn/create" \
  "${HEADERS[@]}" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Mi Primer Himno",
    "content": {"verses": [{"number": 1, "lines": ["Línea uno"]}]},
    "user": {"email": "editor@ejemplo.com", "name": "Editor"}
  }'

# 4. Listar himnos
curl -s "$BASE/hymn/all" "${HEADERS[@]}"

# 5. Obtener himno por ID
curl -s "$BASE/hymn/1" "${HEADERS[@]}"
```
