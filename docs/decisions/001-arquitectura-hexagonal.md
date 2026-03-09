# ADR-001 — Arquitectura hexagonal (puertos y adaptadores)

**Estado:** Aceptado
**Fecha:** 2025-01-01

---

## Contexto

El proyecto necesita una estructura que permita evolucionar independientemente la lógica de negocio, el transporte HTTP y la persistencia. En una aplicación de gestión de himnos, es probable que cambien el framework HTTP, el motor de base de datos o el protocolo de comunicación sin que eso deba afectar las reglas de negocio.

La alternativa más directa sería una arquitectura en capas planas (controlador → servicio → repositorio) acoplada a tecnologías concretas. Esto funciona para proyectos pequeños, pero dificulta las pruebas unitarias y el reemplazo de componentes.

## Decisión

Se adoptó **arquitectura hexagonal** organizando el código en tres zonas:

- **Núcleo de dominio** (`internal/core/`) — entidades, puertos (interfaces) y servicios. Sin dependencias de infraestructura.
- **Adaptadores primarios** (`internal/infra/primary/`) — controladores HTTP que traducen requests externos a llamadas al dominio.
- **Adaptadores secundarios** (`internal/infra/secondary/`) — repositorios que implementan los puertos usando PostgreSQL/pgx.

Las dependencias siempre apuntan hacia adentro: la infraestructura depende del núcleo, nunca al revés.

```
Controladores HTTP  →  Servicios (dominio)  ←  Repositorios PostgreSQL
                              ↑
                         Interfaces (ports)
```

## Consecuencias

**Positivas:**
- El dominio es testeable sin necesidad de levantar base de datos ni servidor HTTP (usando mocks del puerto).
- Cambiar el framework HTTP o el motor de base de datos no afecta las reglas de negocio.
- La estructura de directorios refleja las capas, facilitando la navegación y la incorporación de nuevos desarrolladores.
- Los contratos entre capas están explícitamente definidos como interfaces Go.

**Trade-offs:**
- Mayor cantidad de archivos y capas de abstracción para funcionalidades simples.
- Requiere un mapeo explícito DTO → dominio en cada controlador.
- El cableado manual de dependencias (sin framework de DI) exige disciplina para mantenerlo ordenado.
