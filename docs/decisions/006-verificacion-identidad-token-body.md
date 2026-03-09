# ADR-006 — Verificación doble de identidad: token y body

**Estado:** Aceptado
**Fecha:** 2025-01-01

---

## Contexto

En los endpoints de escritura (`POST /hymn/create`, `PUT /hymn/:id`), el cliente envía un objeto `user` en el body del request con el email y nombre del autor. El sistema también recibe un JWT en el header de autorización que contiene esos mismos datos.

La pregunta era: ¿es suficiente con validar el JWT, o debe exigirse también que el `user` del body coincida con los claims del token?

Validar solo el JWT es el enfoque más común y sencillo. Sin embargo, deja abierta la posibilidad de que un usuario autenticado envíe en el body los datos de otro usuario (por ejemplo, atribuir la creación de un himno a otro autor).

## Decisión

Se exige **verificación doble**: el `user.email` y el `user.name` del body deben coincidir exactamente con los claims `email` y `name` del JWT. Si no coinciden, la respuesta es `401 Unauthorized`.

La validación se realiza en `ValidateUser` dentro del controlador, antes de llamar al servicio:

```
1. Extraer JWT del header Authorization
2. Validar firma y expiración del token
3. Comparar token.email  ==  body.user.email
4. Comparar token.name   ==  body.user.name
5. Verificar que el rol del token sea admin o editor
```

## Consecuencias

**Positivas:**
- Garantiza que el autor registrado en la operación coincide con el usuario autenticado; no es posible suplantar la autoría.
- Añade una capa de defensa adicional ante tokens robados que intentan operar con datos de otro usuario.
- El contrato del endpoint es explícito: el cliente debe saber quién es y enviar esa información coherentemente.

**Trade-offs:**
- El cliente debe enviar datos redundantes (email y nombre ya están en el token), lo que agrega verbosidad al body de los requests.
- Un cambio de nombre del usuario invalida cualquier cliente que no actualice su copia local del perfil antes de operar, hasta que renueve el token.
- La comparación de nombre es sensible a mayúsculas y espacios; un nombre con formato distinto al almacenado causa un 401 difícil de diagnosticar sin logs claros.
