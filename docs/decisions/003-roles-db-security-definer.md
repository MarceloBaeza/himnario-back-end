# ADR-003 — Separación de roles de base de datos con SECURITY DEFINER

**Estado:** Aceptado
**Fecha:** 2025-01-01

---

## Contexto

Con un único usuario de base de datos para todo (migraciones + runtime), cualquier vulnerabilidad de inyección SQL o error de programación podría ejecutarse con permisos DDL (capacidad de alterar o eliminar tablas). Además, el usuario de runtime no necesita los mismos privilegios que el usuario que crea el esquema.

Se evaluaron dos enfoques:

1. **Un solo usuario** con todos los permisos — simple de configurar, máximo riesgo.
2. **Dos usuarios separados** — `hymns_migrator` (DDL) y `hymns_app` (runtime) — principio de mínimo privilegio.

## Decisión

Se establecieron **dos roles de base de datos** con responsabilidades distintas:

- **`hymns_migrator`** — dueño del schema; crea tablas, índices y stored procedures; ejecuta migraciones. Nunca es utilizado por la aplicación en runtime.
- **`hymns_app`** — usuario de la aplicación en runtime; permisos mínimos (`SELECT/INSERT/UPDATE/DELETE` sobre las tablas que necesita, `EXECUTE` sobre los SPs).

Además, todos los stored procedures se marcan como **`SECURITY DEFINER`**: se ejecutan con los privilegios de `hymns_migrator` sin importar qué usuario los llame. Esto permite que `hymns_app` invoque los SPs sin necesitar permisos directos sobre las tablas subyacentes.

```
hymns_app  →  llama SP  →  SP corre como hymns_migrator  →  accede a tabla
```

Los permisos futuros se gestionan con `ALTER DEFAULT PRIVILEGES` para que las nuevas tablas hereden automáticamente los grants correctos.

## Consecuencias

**Positivas:**
- Si `hymns_app` es comprometido, el atacante no puede ejecutar DDL (no puede borrar tablas ni crear objetos).
- El principio de mínimo privilegio está modelado directamente en la base de datos, no solo en la aplicación.
- El modelo SECURITY DEFINER centraliza el control de acceso: cambiar permisos significa cambiar los SPs, no la aplicación.
- La separación es explícita y auditada a través de los scripts de `initdb` y las migraciones.

**Trade-offs:**
- Requiere gestionar dos conjuntos de credenciales (una en el `.env` para runtime, otra para migraciones).
- El `docker-compose` y el entorno de desarrollo deben inicializar ambos usuarios correctamente.
- Un error de configuración en los grants puede impedir que la aplicación funcione aunque el código sea correcto.
