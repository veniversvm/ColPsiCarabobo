// Imagen de fondo del hero de las vistas públicas.
//
// Antes cada hero usaba `bg-heraldic`, un degradado azul en 135° definido en
// `app.css`. Estos son los WebP del "arco": `arco-desktop.webp` (16:9) en
// pantallas anchas y `arco-movil.webp` (9:16) en las verticales.
//
// El <picture> va ABSOLUTO detrás del contenido: los heroes son `relative` con
// `overflow-hidden`, así que el recorte lo hace el propio contenedor y esta capa
// no empuja ni desplaza nada.
//
// El contenedor de cada hero DEBE llevar `isolate`: `position: relative` sin
// `z-index` no crea stacking context, y entonces estas capas `-z-10` se pintan
// antes que el fondo del propio bloque, es decir, quedan tapadas y la imagen no
// se ve. `isolate` las confine dentro del hero.
//
// Encima va `bg-velo-hero`: un degradado VERTICAL que oscurece por la banda
// central, donde están el h1 y el párrafo, y se abre en el borde superior e
// inferior, que es donde la foto se ve.
//
// No se puede quitar del todo, y conviene tener claro por qué antes de volver a
// intentarlo: el cielo de estas fotos es brillante de punta a punta — el p90
// está en 180-210 y el p95 en 200-222, medido —, así que el texto blanco se
// apoya en un fondo que da 2.3:1. Una vez se quitó el velo y las 6 franjas
// quedaron ilegibles. Y tampoco vale un velo plano: para tapar el píxel más
// claro (254) haría falta un fondo casi sólido, que es la pantalla azul que se
// rechazó. La medida completa y el razonamiento están en `app.css`, junto a la
// utilidad.
//
// Los textos de estos heroes llevan `text-shadow-hero` (ver `app.css`) como
// segunda garantía, y `bg-colpsi-blue` en el contenedor del hero: si el WebP no
// cargara, el título blanco se quedaría sobre blanco. Ese azul sólido es la red
// de seguridad; el velo solo afina la foto.
//
// OJO: las 9 páginas de /psi NO pasan por aquí. Ellas usan `bg-heraldic` como
// clase de su propio contenedor (`<div class="bg-heraldic pt-12 pb-20">`), no
// como velo, y ahí el degradado es el fondo único de la página. No tocar.
//
// `alt=""` + `aria-hidden` porque es decorativa: no aporta información que no
// esté ya en el texto del hero.
export default function ArcoHeroImage() {
  return (
    <>
      <picture class="absolute inset-0 -z-10">
        <source media="(min-width: 768px)" srcset="/arco-desktop.webp" />
        <img
          src="/arco-movil.webp"
          alt=""
          aria-hidden="true"
          loading="eager"
          decoding="async"
          class="h-full w-full object-cover object-center"
        />
      </picture>

      <div class="absolute inset-0 -z-10 bg-velo-hero" aria-hidden="true" />
    </>
  );
}
