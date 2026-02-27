import { A } from "@solidjs/router";

export default function PublicPortal() {
  return (
    <main class="min-h-screen bg-[#fcfcfc] pb-20">
      {/* Banner de Bienvenida a la Comunidad */}
      <section class="bg-colpsi-blue pt-16 pb-32 px-6 text-center text-white relative">
        <div class="max-w-4xl mx-auto space-y-4">
          <div class="inline-block px-4 py-1 bg-colpsi-yellow text-colpsi-blue rounded-full text-[10px] font-black uppercase tracking-widest mb-2">
            Servicio a la comunidad
          </div>
          <h2 class="text-3xl md:text-5xl font-black tracking-tight leading-none">
            Bienvenido al Portal Público
          </h2>
          <p class="text-blue-100 text-lg max-w-2xl mx-auto font-medium">
            Estamos aquí para orientarle y conectarle con profesionales certificados del estado Carabobo.
          </p>
        </div>
        
        {/* Isotipo Ψ sutil al fondo */}
        <div class="absolute left-1/2 -translate-x-1/2 bottom-[-40px] opacity-10 text-[12rem] font-black select-none pointer-events-none">
          Ψ
        </div>
      </section>

      {/* HUB DE OPCIONES PÚBLICAS */}
      <div class="max-w-5xl mx-auto px-6 -mt-16 relative z-10">
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
          
          {/* Opción: Buscar Profesional */}
          <NavCard 
            title="Directorio Profesional" 
            desc="Encuentre psicólogos certificados y solventes en Carabobo." 
            href="/directorio" 
            icon="🔍" 
            color="border-blue-100"
          />

          {/* Opción: Sobre la Institución */}
          <NavCard 
            title="Sobre Nosotros" 
            desc="Conozca nuestra historia, misión y marco ético profesional." 
            href="/nosotros" 
            icon="🏛️" 
            color="border-yellow-100"
          />

          {/* Opción: Noticias Públicas */}
          <NavCard 
            title="Noticias y Artículos" 
            desc="Información de interés y avisos a la comunidad." 
            href="/noticias" 
            icon="📰" 
            color="border-red-100"
          />

        </div>
      </div>

      {/* Franja Bandera al final para consistencia */}
      <footer class="mt-20 text-center px-6">
        <div class="flex justify-center gap-2 mb-4">
          <div class="w-8 h-1 bg-colpsi-red rounded-full"></div>
          <div class="w-8 h-1 bg-green-700 rounded-full"></div>
          <div class="w-8 h-1 bg-colpsi-blue rounded-full"></div>
        </div>
        <p class="text-gray-400 text-xs font-bold uppercase tracking-widest">
          Colegio de Psicólogos de Carabobo
        </p>
      </footer>
    </main>
  );
}

// Componente para las tarjetas del menú (puedes moverlo a components/ui/)
function NavCard(props: { title: string, desc: string, href: string, icon: string, color: string }) {
  return (
    <A 
      href={props.href} 
      class={`bg-white border-2 ${props.color} p-8 rounded-[2.5rem] shadow-sm hover:shadow-2xl hover:shadow-blue-900/5 hover:-translate-y-2 transition-all group flex flex-col items-center text-center`}
    >
      <div class="w-20 h-20 bg-gray-50 rounded-3xl flex items-center justify-center text-4xl mb-6 group-hover:bg-colpsi-yellow transition-colors duration-300">
        {props.icon}
      </div>
      <h3 class="text-colpsi-blue font-black uppercase text-sm tracking-wide mb-3">{props.title}</h3>
      <p class="text-colpsi-muted text-xs leading-relaxed font-medium">{props.desc}</p>
    </A>
  );
}