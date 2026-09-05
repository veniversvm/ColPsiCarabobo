# solid-dnd 0.7.5 (parche local)

Copia de `@thisbeyond/solid-dnd@0.7.5` (última versión publicada) con UN cambio.

## Por qué existe

Con Solid 1.9.x el `createPointerSensor` del dist original lanza
`ReferenceError: Cannot access 'attach' before initialization` al montar el
kanban: el `onMount(() => addSensor({ activators: { pointerdown: attach } }))`
referencia `const attach`, que en el dist original está declarado DESPUÉS del
`onMount`. Solid 1.9 ejecuta los efectos de forma eager durante el render, así
que la lectura ocurre antes de que `attach` se inicialice (temporal dead zone).

El bug no tiene fix upstream (0.7.5 es `latest`).

## El parche

En `createPointerSensor` (`index.js`) la declaración de `attach` se movió
ARRIBA del `onMount()` que la referencia. No cambia comportamiento: `attach`
solo se ejecuta en el evento `pointerdown`, momento en que el resto de consts
ya están inicializadas.

## Re-sincronizar

- Cambios del paquete: `cp node_modules/@thisbeyond/solid-dnd/dist/index.js index.js`
  y `cp …/dist/index.d.ts index.d.ts`, reaplicando el reorder descrito arriba.
- El único consumidor es `src/routes/admin/proyectos/[id].tsx` (import `~/vendor/thisbeyond-solid-dnd`).