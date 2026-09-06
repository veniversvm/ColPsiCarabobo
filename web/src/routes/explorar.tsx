import { A } from "@solidjs/router";
import { Title, Meta, Link } from "@solidjs/meta";

const SITE_URL = import.meta.env.VITE_SITE_URL || "https://colpsicarabobo.org";

export default function PublicPortal() {
  const canonicalUrl = `${SITE_URL}/explorar`; 
  const ogImage = `${SITE_URL}/og-default.jpg`;
  const pageTitle = "Bienvenidos al Colegio de Psicólogos | Colegio de Psicólogos del Estado Carabobo";
  const pageDescription =
    "Conéctate con profesionales certificados, conoce nuestra institución, tramita tu inscripción y entérate de las últimas noticias del Colegio de Psicólogos del Estado Carabobo.";

  return (
    <>
      {/* ── SEO & METADATA (SSR) ───────────────────────────────────────────── */}
      <Title>{pageTitle}</Title>
      <Meta name="description" content={pageDescription} />
      <Meta
        name="keywords"
        content="Psicólogos, Carabobo, Directorio Profesional, Salud Mental, Valencia, Colegio de Psicólogos, Venezuela, Inscripción Psicología"
      />
      <Meta name="robots" content="index, follow" />

      <Meta property="og:type" content="website" />
      <Meta property="og:url" content={canonicalUrl} />
      <Meta property="og:title" content={pageTitle} />
      <Meta property="og:description" content={pageDescription} />
      <Meta property="og:image" content={ogImage} />
      <Meta property="og:site_name" content="Colegio de Psicólogos del Estado Carabobo" />
      <Meta property="og:locale" content="es_VE" />

      <Meta name="twitter:card" content="summary_large_image" />
      <Meta name="twitter:title" content={pageTitle} />
      <Meta name="twitter:description" content={pageDescription} />
      <Meta name="twitter:image" content={ogImage} />

      <Link rel="canonical" href={canonicalUrl} />
      {/* ─────────────────────────────────────────────────────────────────── */}

      <main class="min-h-screen bg-colpsi-bg pb-20 font-sans">
        {/* Banner de Bienvenida a la Comunidad */}
        <section class="bg-heraldic pt-16 pb-32 px-6 text-center text-white relative shadow-inner overflow-hidden">
          <div class="max-w-4xl mx-auto space-y-5 relative z-10">
            <div class="inline-block px-5 py-1.5 bg-colpsi-yellow text-colpsi-blue rounded-full text-[10px] font-black uppercase tracking-[0.2em] mb-2 shadow-sm border border-colpsi-yellow/50">
              Servicio a la comunidad
            </div>
               <h1 class="text-4xl md:text-5xl font-black text-white tracking-tight leading-tight">
              Colegio de Psicólogos <br class="hidden sm:block" />
              <span class="text-colpsi-yellow italic">del Estado Carabobo</span>
            </h1>
            <p class="text-blue-200 text-lg max-w-2xl mx-auto font-medium leading-relaxed">
              Estamos aquí para orientarle y conectarle con profesionales certificados y solventes en su región.
            </p>
          </div>
          
          {/* Isotipo Ψ sutil al fondo */}
          <div class="absolute left-1/2 -translate-x-1/2 bottom-[-40px] opacity-10 text-[12rem] font-black select-none pointer-events-none">
            Ψ
          </div>
        </section>

        {/* HUB DE OPCIONES PÚBLICAS */}
        <div class="max-w-[90rem] mx-auto px-6 -mt-16 relative z-20">
          <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
            
            {/* Opción: Buscar Profesional */}
            <NavCard 
              title="Directorio Profesional" 
              desc="Encuentre psicólogos certificados y solventes en Carabobo." 
              href="/directorio" 
              icon="🔍" 
              color="border-blue-100"
            />

            {/* Opción: Noticias Públicas */}
            <NavCard 
              title="Noticias y Artículos" 
              desc="Información de interés y avisos a la comunidad." 
              href="/noticias" 
              icon="📰" 
              color="border-red-100"
            />

            {/* Opción: Inscripción (NUEVA) */}
            <NavCard 
              title="Inscripción" 
              desc="Guía paso a paso y requisitos legales para nuevos agremiados." 
              href="/inscripcion" 
              icon="⚖️" 
              color="border-emerald-100"
            />

            {/* Opción: Sobre la Institución */}
            <NavCard 
              title="Sobre Nosotros" 
              desc="Conozca nuestra historia, misión y marco ético profesional." 
              href="/nosotros" 
              icon="🏛️" 
              color="border-yellow-100"
            />

            

          </div>
        </div>

        {/* Franja Bandera al final para consistencia */}
        <footer class="mt-24 text-center px-6">
          <div class="flex justify-center gap-2 mb-4">
            <div class="w-8 h-1 bg-colpsi-red rounded-full"></div>
            <div class="w-8 h-1 bg-green-700 rounded-full"></div>
            <div class="w-8 h-1 bg-colpsi-blue rounded-full"></div>
          </div>
          <p class="text-gray-400 text-[10px] font-black uppercase tracking-widest">
            Colegio de Psicólogos de Carabobo
          </p>
        </footer>
      </main>
    </>
  );
}

// Componente para las tarjetas del menú
function NavCard(props: { title: string, desc: string, href: string, icon: string, color: string }) {
  return (
    <A 
      href={props.href} 
      class={`bg-white border-2 ${props.color} p-8 md:p-10 rounded-[2.5rem] shadow-premium hover:shadow-2xl hover:border-colpsi-blue/30 hover:-translate-y-2 transition-all duration-300 group flex flex-col items-center text-center`}
    >
      <div class="w-20 h-20 bg-gray-50 border border-gray-100 rounded-3xl flex items-center justify-center text-4xl mb-6 group-hover:bg-colpsi-yellow group-hover:scale-110 transition-all duration-300">
        {props.icon}
      </div>
      <h3 class="text-colpsi-blue font-black uppercase text-sm tracking-widest mb-3 leading-tight">
        {props.title}
      </h3>
      <p class="text-gray-500 text-sm leading-relaxed font-medium">
        {props.desc}
      </p>
    </A>
  );
}