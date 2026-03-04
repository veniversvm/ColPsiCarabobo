// web/src/routes/directorio/index.tsx (refactorizado)
/**
 * Página principal del directorio profesional. Aquí los usuarios pueden buscar psicólogos por nombre, cédula, FPV o especialidad.
 */
import { createEffect, createResource, createSignal, Suspense } from "solid-js";
import { apiGet } from "~/lib/api";
import { DirectoryPsychologist } from "~/types/psi";
import { SearchHeader } from "~/components/directory/SearchHeader";
import { ResultsGrid } from "~/components/directory/ResultsGrid";
import { FlagFooter } from "~/components/ui/FlagFooter";

export default function DirectoryPage() {
  // Estados para los valores del formulario
  const [query, setQuery] = createSignal("");
  const [specialty, setSpecialty] = createSignal("");
  const [location, setLocation] = createSignal("");
  
  // Estado "Trigger" (El que realmente dispara la petición)
  const [searchParams, setSearchParams] = createSignal({ q: "", spec: "", location: "" });

  // Cargamos el catálogo de especialidades
  const [specialties] = createResource(() => apiGet<any[]>("/specialties"));

  // El recurso escucha ÚNICAMENTE a searchParams()
  const [data] = createResource(
    () => searchParams(),
    async (params) => {
      return await apiGet<{ data: DirectoryPsychologist[] }>(
        `/psi/directory?q=${params.q}&specialty=${params.spec}&limit=12&location=${params.location}`
      );
    }
  );

  // Manejador del botón Buscar
  const handleSearch = (e: Event) => {
    e.preventDefault();
    setSearchParams({ q: query(), spec: specialty(), location: location() });
  };


  // En DirectoryPage, agrega este log para ver qué campos vienen:
createEffect(() => {
  const results = data();
  if (results?.data) {
    console.log("Primer psicólogo del directorio:", results.data[0]);
    // Esto mostrará todos los campos disponibles
  }
});

  return (
    <main class="min-h-screen bg-[#fcfcfc] pb-24">
      <SearchHeader
        query={query()}
        specialty={specialty()}
        location={location()}
        specialties={specialties()}
        onQueryChange={setQuery}
        onSpecialtyChange={setSpecialty}
        onLocationChange={setLocation}
        onSearch={handleSearch}
      />

      <div class="max-w-7xl mx-auto px-6 -mt-10 relative z-20">
        <Suspense fallback={<div class="h-96 bg-white animate-pulse rounded-[2.5rem]" />}>
          <ResultsGrid 
            psychologists={data()?.data}
            loading={data.loading}
          />
        </Suspense>
      </div>

      <FlagFooter />
    </main>
  );
}