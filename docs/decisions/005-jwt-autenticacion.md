# ADR-005 — JWT para autenticación

**Estado:** Aceptado
**Fecha:** 2025-01-01

---

## Contexto

La API necesita un mecanismo de autenticación para proteger los endpoints de escritura (crear y editar himnos). Los requisitos principales eran:

- Sin estado en el servidor (no sessions persistidas en base de datos).
- El token debe transportar suficiente información para validar identidad y rol sin consultar la base de datos en cada request.
- Compatibilidad con clientes web y móviles.

Se evaluaron dos alternativas principales:

1. **Sesiones con cookie** — el servidor guarda el estado de sesión (en memoria o base de datos); cada request valida contra ese estado. Requiere almacenamiento de sesiones y complica el escalado horizontal.
2. **JWT (JSON Web Token)** — token firmado que contiene los claims del usuario; el servidor solo necesita la clave secreta para validarlo. Sin estado.

## Decisión

Se implementó autenticación mediante **JWT con algoritmo HS256**. El token se emite en `POST /user/authentication` y se envía en el header `Authorization: Bearer <token>` en cada request protegido.

Configuración:
- **Algoritmo:** HS256 (HMAC-SHA256)
- **Expiración:** 60 minutos
- **Claims incluidos:** `email`, `name`, `role`, `iat`, `exp`, `iss`

El token es validado en el handler del controlador (no en middleware global) para aplicarlo solo a las rutas que lo requieren.

## Consecuencias

**Positivas:**
- Sin estado en el servidor; no hay consultas a base de datos para validar el token.
- El rol y la identidad del usuario viajan en el token, permitiendo validaciones sin round-trips adicionales.
- Fácil integración con cualquier cliente HTTP.
- El escalado horizontal no requiere sesiones compartidas.

**Trade-offs:**
- No es posible invalidar un token antes de su expiración (no hay blacklist); un token comprometido es válido hasta que expire (60 min).
- La clave secreta (`JWT_SECRET_KEY`) debe mantenerse segura y rotarse con cuidado; su exposición compromete todos los tokens vigentes.
- La información del usuario en el token puede quedar desactualizada si cambia el rol durante la vigencia del token.
