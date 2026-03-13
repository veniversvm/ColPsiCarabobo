// web/src/components/admin/noticias/NoticiasHeader.tsx
import { A } from "@solidjs/router";

export function NoticiasHeader() {
  return (
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-8 bg-white p-6 rounded-3xl shadow-sm border border-gray-100">
      <div>
        <h1 class="text-2xl font-black text-blue-900 uppercase tracking-tight">Publicaciones</h1>
        <p class="text-gray-400 text-sm mt-0.5 font-medium">Gestión de noticias y comunicados del Colegio</p>
      </div>
      <A
        href="/admin/noticias/crear"
        class="inline-flex items-center gap-2 bg-blue-800 hover:bg-blue-900 text-white font-black px-6 py-3 rounded-2xl shadow-lg hover:scale-105 active:scale-95 transition-all text-sm"
      >
        <span class="text-lg leading-none">＋</span>
        Nueva Publicación
      </A>
    </div>
  );
}