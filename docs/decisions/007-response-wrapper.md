# ADR-007 — Envoltorio de respuesta uniforme (ResponseWrapper)

**Estado:** Aceptado
**Fecha:** 2025-01-01

---

## Contexto

Sin una convención de respuesta, cada endpoint puede retornar JSON con estructura diferente: a veces el dato directamente, a veces con un campo `data`, a veces con `error` como string, otras como objeto. Esto complica la integración del frontend, que necesita manejar múltiples formatos.

Se evaluaron dos enfoques:

1. **Respuestas directas** — cada endpoint retorna la estructura más natural para su caso (el recurso directamente en éxito, un objeto de error en fallo).
2. **Envoltorio uniforme** — todas las respuestas, exitosas o no, siguen la misma estructura raíz.

## Decisión

Todas las respuestas HTTP del sistema se envuelven en un **`ResponseWrapper`** que contiene exactamente uno de dos campos, nunca ambos:

**Éxito:**
```json
{
  "responseOk": {
    "statusCode": 200,
    "message": "descripción legible",
    "data": { }
  }
}
```

**Error:**
```json
{
  "responseError": {
    "statusCode": 400,
    "error": "descripción del error",
    "data": {
      "campo": "descripción del problema"
    }
  }
}
```

El campo `data` en errores permite retornar errores por campo (validaciones), manteniendo la misma estructura raíz.

## Consecuencias

**Positivas:**
- El frontend puede parsear todas las respuestas con el mismo código: verificar si existe `responseOk` o `responseError`.
- Los errores de validación campo a campo tienen un lugar natural en `responseError.data`.
- El `statusCode` dentro del cuerpo permite al cliente leerlo sin inspeccionar el header HTTP.
- La estructura es predecible y documentable de forma uniforme en el `swagger.yaml`.

**Trade-offs:**
- Las respuestas son más verbosas que retornar el recurso directamente.
- Los clientes que esperan el recurso en la raíz del JSON deben adaptarse a acceder a `responseOk.data`.
- Agregar el `statusCode` en el body duplica información que ya está en el header HTTP.
