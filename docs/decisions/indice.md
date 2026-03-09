# Decisiones de arquitectura (ADRs)

← [Volver al índice de arquitectura](../architecture/architecture.md)

Registro de las decisiones técnicas significativas del proyecto. Cada ADR documenta el contexto que motivó la decisión, la alternativa elegida y sus consecuencias.

---

## Código

| N° | Decisión | Estado |
|---|---|---|
| [ADR-001](./001-arquitectura-hexagonal.md) | Arquitectura hexagonal (puertos y adaptadores) | Aceptado |
| [ADR-007](./007-response-wrapper.md) | Envoltorio de respuesta uniforme (`ResponseWrapper`) | Aceptado |

## Autenticación y seguridad

| N° | Decisión | Estado |
|---|---|---|
| [ADR-005](./005-jwt-autenticacion.md) | JWT para autenticación (HS256, 60 min) | Aceptado |
| [ADR-006](./006-verificacion-identidad-token-body.md) | Verificación doble de identidad: token y body | Aceptado |

## Base de datos

| N° | Decisión | Estado |
|---|---|---|
| [ADR-002](./002-stored-procedures.md) | Stored procedures para todas las operaciones de base de datos | Aceptado |
| [ADR-003](./003-roles-db-security-definer.md) | Separación de roles de base de datos con SECURITY DEFINER | Aceptado |
| [ADR-004](./004-jsonb-contenido-himnos.md) | JSONB para el contenido de los himnos | Aceptado |
| [ADR-008](./008-migraciones-automaticas.md) | Migraciones automáticas al arrancar la aplicación | Aceptado |

---

## Estados posibles

| Estado | Significado |
|---|---|
| **Propuesto** | En discusión, aún no implementado |
| **Aceptado** | Implementado y vigente |
| **Deprecado** | Reemplazado por una decisión posterior |
| **Rechazado** | Evaluado y descartado |
