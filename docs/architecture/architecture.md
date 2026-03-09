# Arquitectura — Himnario API

Documentación técnica del sistema dividida por área. Cada sección vive en su propio archivo para facilitar la navegación y el mantenimiento.

---

## Contenido

| Documento | Descripción |
|---|---|
| [Arquitectura de código](./code-architecture.md) | Capas hexagonales, estructura de directorios, ciclo de vida de un request, cadena de middlewares, inyección de dependencias y patrones clave |
| [Arquitectura de base de datos](./database-architecture.md) | Tablas, relaciones, índices, stored procedures, roles de base de datos e historial de migraciones |
| [Decisiones de arquitectura (ADRs)](../decisions/indice.md) | Registro de las decisiones técnicas significativas, con contexto, motivación y consecuencias |

---

## Visión general

El proyecto es una **API REST** construida en Go con el framework Gin, siguiendo **arquitectura hexagonal** (puertos y adaptadores). La base de datos es PostgreSQL 16, accedida exclusivamente a través de stored procedures.

```
┌──────────────────────────────────────────────────────┐
│               Adaptadores Primarios                  │
│         (Controladores HTTP — handlers Gin)          │
└───────────────────────────┬──────────────────────────┘
                            │ llama
┌───────────────────────────▼──────────────────────────┐
│                    Capa de Dominio                    │
│    domain/  ←  port/ (interfaces)  ←  service/       │
└───────────────────────────┬──────────────────────────┘
                            │ implementa
┌───────────────────────────▼──────────────────────────┐
│              Adaptadores Secundarios                  │
│          (Repositorios PostgreSQL — pgx)              │
└──────────────────────────────────────────────────────┘
```

**Stack principal:**

| Componente | Tecnología |
|---|---|
| Lenguaje | Go 1.25.5 |
| Framework HTTP | Gin 1.11.0 |
| Base de datos | PostgreSQL 16 |
| Driver DB | pgx/v5 |
| Migraciones | golang-migrate |
| Autenticación | JWT (HS256, 60 min) |
| Hashing | bcrypt (cost=12) |
| Validación | go-playground/validator |
