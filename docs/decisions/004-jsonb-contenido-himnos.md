# ADR-004 — JSONB para el contenido de los himnos

**Estado:** Aceptado
**Fecha:** 2025-01-01

---

## Contexto

El contenido de un himno (versos, estrofas, coro, notas musicales, etc.) es inherentemente variable. Distintos himnos pueden tener estructuras diferentes: algunos tienen coro, otros no; algunos tienen secciones adicionales como puentes o introducciones. El esquema exacto puede evolucionar con el tiempo a medida que se incorporen más tipos de contenido.

Se evaluaron tres alternativas:

1. **Columnas relacionales fijas** — `verse_1`, `verse_2`, `chorus`, etc. Requería conocer de antemano la estructura máxima y generaba columnas vacías para la mayoría de los registros.
2. **Tabla separada de secciones** — relación `hymns` ↔ `hymn_sections` con tipo y contenido. Flexible, pero requería múltiples joins para leer un himno completo.
3. **JSONB** — el contenido completo como un único documento JSON almacenado en PostgreSQL con soporte de consultas e índices.

## Decisión

El campo `content` de la tabla `hymns` se almacena como **`JSONB`**. La estructura del documento es libre y la define la aplicación cliente al momento de crear o editar un himno.

Ejemplo de estructura típica:

```json
{
  "verses": [
    { "number": 1, "lines": ["Línea uno", "Línea dos"] },
    { "number": 2, "lines": ["Línea tres", "Línea cuatro"] }
  ],
  "chorus": "Estribillo del himno"
}
```

Se agrega un índice GIN (`idx_hymns_content_gin`) sobre `content` para habilitar búsquedas eficientes dentro del documento.

## Consecuencias

**Positivas:**
- La estructura del contenido puede evolucionar sin necesidad de migraciones de esquema.
- Un himno completo (metadata + contenido) se lee en una sola fila, sin joins.
- El índice GIN permite búsquedas full-text y consultas JSON path directamente en PostgreSQL.
- Distintos tipos de himnos pueden tener estructuras diferentes sin afectar el resto.

**Trade-offs:**
- No existe un esquema formal validado a nivel de base de datos; la validación estructural es responsabilidad de la aplicación o del cliente.
- Las consultas que filtran por campos internos del JSONB son menos intuitivas que el SQL estándar sobre columnas relacionales.
- La legibilidad directa de los datos en la base de datos es menor que con columnas convencionales.
