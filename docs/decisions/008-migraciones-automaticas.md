# ADR-008 — Migraciones automáticas al arrancar la aplicación

**Estado:** Aceptado
**Fecha:** 2025-01-01

---

## Contexto

El esquema de base de datos evoluciona junto con el código. Necesitamos una estrategia para aplicar cambios de esquema (nuevas tablas, nuevos SPs, ajustes de permisos) de forma reproducible y controlada.

Las alternativas evaluadas fueron:

1. **Migraciones manuales** — el operador aplica los archivos SQL a mano antes de desplegar. Simple, pero propenso a errores humanos y difícil de coordinar en equipos.
2. **Migraciones automáticas al arrancar** — la aplicación detecta y aplica las migraciones pendientes en cada inicio.
3. **Migraciones como job separado** — un contenedor o job de CI aplica migraciones antes de que la aplicación arranque. Más control, más infraestructura.

## Decisión

Se eligió **`golang-migrate`** con ejecución automática al arrancar la API. El runner de migraciones se invoca en `instance/primary.go` como primer paso de la inicialización, antes de conectar los repositorios y levantar el servidor.

- Los archivos de migración viven en `resources/migrations/` con el formato `NNNNNN_nombre.up.sql` / `NNNNNN_nombre.down.sql`.
- Las migraciones se aplican con el usuario `hymns_migrator` (permisos DDL), no con `hymns_app`.
- `golang-migrate` mantiene una tabla interna (`schema_migrations`) para registrar qué versiones ya fueron aplicadas; las migraciones son idempotentes respecto a versiones ya ejecutadas.

## Consecuencias

**Positivas:**
- El esquema siempre está sincronizado con el código que se despliega, sin pasos manuales.
- En desarrollo, levantar el proyecto por primera vez crea toda la estructura automáticamente.
- Las migraciones están versionadas junto al código fuente; el historial de cambios de esquema es visible en git.
- El rollback está disponible vía los archivos `.down.sql` si es necesario revertir manualmente.

**Trade-offs:**
- Si una migración falla al arrancar, la aplicación no levanta; un error de SQL en una migración bloquea el despliegue completo.
- En entornos de producción con múltiples instancias, si dos instancias arrancan simultáneamente pueden intentar migrar al mismo tiempo (aunque `golang-migrate` usa un lock para evitarlo).
- Las migraciones `.down.sql` deben mantenerse actualizadas manualmente; un rollback incompleto puede dejar el esquema en un estado inconsistente.
