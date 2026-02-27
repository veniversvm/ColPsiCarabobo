import { createResource, For, Show } from "solid-js";
import { A } from "@solidjs/router";
import { useAuth } from "~/lib/auth";

export default function Portal() {
  const { user, role } = useAuth();

  return (
    <main class="min-h-screen bg-[#fcfcfc] pb-20">
      {/* Banner de Bienvenida Institucional */}
      <section class="bg-colpsi-blue pt-16 pb-28 px-6 text-center text-white relative">
        <div class="max-w-4xl mx-auto space-y-2">
          <h2 class="text-3xl md:text-4xl font-black tracking-tight">
            ¡Bienvenido(a), {user()?.firstName || user()?.username}!
          </h2>
          <p class="text-blue-200 text-lg">Portal del Colegio de Psicólogos del Estado Carabobo</p>
        </div>
        
        {/* Decoración sutil del Sol de Carabobo al fondo */}
        <div class="absolute right-[-50px] top-[-50px] w-64 h-64 bg-colpsi-yellow/5 rounded-full blur-3xl" />
      </section>

      {/* HUB DE OPCIONES (Grid de Navegación) */}
      <div class="max-w-5xl mx-auto px-6 -mt-12">
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
          
          {/* Opción: Directorio (Visible para todos los logueados) */}
          <NavCard 
            title="Directorio Público" 
            desc="Consulta y busca colegas federados." 
            href="/directorio" 
            icon="🔍" 
            color="border-blue-100"
          />

          {/* Opción: Perfil (Psicólogo) */}
          <Show when={role() === "psi"}>
            <NavCard 
              title="Mi Perfil" 
              desc="Actualiza tus datos y visibilidad pública." 
              href="/psi/perfil" 
              icon="👤" 
              color="border-yellow-100"
            />
            <NavCard 
              title="Postgrados" 
              desc="Gestiona tus títulos y certificados." 
              href="/psi/academico" 
              icon="🎓" 
              color="border-green-100"
            />
          </Show>

          {/* Opciones: Admin */}
          <Show when={role() === "admin"}>
            <NavCard 
              title="Gestión de Miembros" 
              desc="Altas, bajas y edición de expedientes." 
              href="/admin/psicologos" 
              icon="📁" 
              color="border-red-100"
            />
            <NavCard 
              title="Noticias y CMS" 
              desc="Publica avisos y noticias gremiales." 
              href="/admin/noticias" 
              icon="📰" 
              color="border-indigo-100"
            />
            <NavCard 
              title="Métricas" 
              desc="Estado del servidor y estadísticas." 
              href="/admin/metrics" 
              icon="📊" 
              color="border-gray-200"
            />
          </Show>

        </div>
      </div>
    </main>
  );
}

// Componente Interno para las tarjetas del menú
function NavCard(props: { title: string, desc: string, href: string, icon: string, color: string }) {
  return (
    <A 
      href={props.href} 
      class={`bg-white border-2 ${props.color} p-6 rounded-[2rem] shadow-sm hover:shadow-xl hover:-translate-y-1 transition-all group flex flex-col items-center text-center`}
    >
      <div class="w-16 h-16 bg-gray-50 rounded-2xl flex items-center justify-center text-3xl mb-4 group-hover:bg-colpsi-yellow transition-colors">
        {props.icon}
      </div>
      <h3 class="text-colpsi-blue font-black uppercase text-sm tracking-wide mb-2">{props.title}</h3>
      <p class="text-colpsi-muted text-xs leading-relaxed">{props.desc}</p>
    </A>
  );
}