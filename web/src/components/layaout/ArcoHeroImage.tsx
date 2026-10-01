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
// Encima va el MISMO degradado heráldico a modo de velo: los títulos son
// `text-white` y la foto sola no garantiza contraste. Al conservar el velo, si
// el WebP llegara a fallar el hero se ve igual que antes (degradado pleno) en
// vez de romperse. `alt=""` + `aria-hidden` porque es decorativa: no aporta
// información que no esté ya en el texto del hero.
export default function ArcoHeroImage(props: { velo?: number }) {
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

      <div
        class="absolute inset-0 -z-10 bg-heraldic"
        style={{ opacity: props.velo ?? 0.72 }}
        aria-hidden="true"
      />
    </>
  );
}
