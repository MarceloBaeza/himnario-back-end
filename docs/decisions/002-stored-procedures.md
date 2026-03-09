# ADR-002 — Stored procedures para todas las operaciones de base de datos

**Estado:** Aceptado
**Fecha:** 2025-01-01

---

## Contexto

Al diseñar la capa de persistencia se evaluaron tres alternativas:

1. **ORM** (GORM, ent) — mapeo automático de structs a SQL, con query builder.
2. **SQL en el repositorio** — queries escritas directamente en Go usando `pgx`.
3. **Stored procedures** — toda la lógica SQL vive en la base de datos; el repositorio solo llama funciones.

El proyecto maneja lógica de negocio en la base de datos (validación de unicidad case-insensitive de títulos, resolución de usuario por email antes de insertar un himno, verificación de existencia antes de actualizar). Colocar esa lógica en el repositorio Go requería múltiples round-trips o queries condicionales complejas.

## Decisión

**Todas las operaciones de escritura y las lecturas complejas se realizan exclusivamente a través de stored procedures** de PostgreSQL. Los repositorios Go únicamente invocan funciones o procedimientos; no construyen SQL dinámico.

Cada operación tiene su SP dedicado:
- `sp_create_user` — creación con normalización de email y control de duplicados
- `sp_create_hymn_by_author_email` — creación con resolución de autor y unicidad de título
- `sp_update_hymn` — actualización con verificación de existencia
- `sp_list_hymns`, `sp_get_hymn_by_id` — lecturas

## Consecuencias

**Positivas:**
- La lógica compleja (validaciones, resolución de FK, unicidad) vive en un único lugar y se ejecuta atómicamente.
- Los repositorios Go son delgados y predecibles: una función → una llamada SP.
- El SP puede cambiarse o optimizarse sin recompilar la aplicación.
- Los errores de negocio se comunican con `RAISE EXCEPTION` y códigos SQLSTATE conocidos, simplificando el manejo de errores en Go.

**Trade-offs:**
- La lógica de negocio queda dividida entre Go (reglas de dominio) y SQL (validaciones de persistencia), lo que puede dificultar seguir el flujo completo.
- Las migraciones deben gestionar tanto el esquema como los SPs; un cambio de SP requiere una nueva migración.
- Las pruebas de integración de los SPs requieren una base de datos real (no se pueden mockear fácilmente).
