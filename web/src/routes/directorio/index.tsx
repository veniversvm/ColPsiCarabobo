// web/src/routes/directorio/index.tsx
/**
 * Página principal del directorio profesional.
 * SSR no aplica aquí: es búsqueda interactiva, no hay contenido estático que indexar.
 */
import { createResource, createSignal, Suspense } from "solid-js";
import { apiGet } from "~/lib/api";
import { DirectoryPsychologist } from "~/types/psi";
import { SearchHeader } from "~/components/directory/SearchHeader";
import { ResultsGrid } from "~/components/directory/ResultsGrid";
import { FlagFooter } from "~/components/ui/FlagFooter";
import { Meta, Title } from "@solidjs/meta";

export default function DirectoryPage() {
  const [query, setQuery] = createSignal("");
  const [specialty, setSpecialty] = createSignal("");
  const [location, setLocation] = createSignal("");
  const [searchParams, setSearchParams] = createSignal({ q: "", spec: "", location: "" });

  const [specialties] = createResource(() => apiGet<any[]>("/specialties"));

  const [data] = createResource(
    () => searchParams(),
    async (params) => {
      return await apiGet<{ data: DirectoryPsychologist[] }>(
        `/psi/directory?q=${params.q}&specialty=${params.spec}&limit=12&location=${params.location}`
      );
    }
  );

  const handleSearch = (e: Event) => {
    e.preventDefault();
    setSearchParams({ q: query(), spec: specialty(), location: location() });
  };

  return (
    <>
      <Title>Directorio de Psicólogos | COLPSI Carabobo</Title>
      <Meta
        name="description"
        content="Encuentra psicólogos colegiados en el estado Carabobo, Venezuela. Busca por nombre, especialidad o ubicación."
      />
      <Meta name="robots" content="index, follow" />

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
    </>
  );
}