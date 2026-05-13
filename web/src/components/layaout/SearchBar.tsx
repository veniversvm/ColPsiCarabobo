// web/src/components/SearchBar.tsx
import { createSignal, onCleanup } from "solid-js";

export function SearchBar(props: { onSearch: (query: string) => void }) {
  const [searchTerm, setSearchTerm] = createSignal("");
  let timeoutId: ReturnType<typeof setTimeout> | undefined;

  const handleSearch = (value: string) => {
    setSearchTerm(value);
    
    // Limpiar el timeout anterior
    if (timeoutId) clearTimeout(timeoutId);
    
    // Establecer nuevo timeout de 1.5 segundos
    timeoutId = setTimeout(() => {
      props.onSearch(value);
    }, 1500);
  };

  // Limpiar timeout al desmontar el componente
  onCleanup(() => {
    if (timeoutId) clearTimeout(timeoutId);
  });

  return (
    <div class="relative">
      <input
        type="text"
        value={searchTerm()}
        onInput={(e) => handleSearch(e.currentTarget.value)}
        placeholder="Buscar..."
        class="w-full px-4 py-2 pl-10 pr-4 text-sm bg-white border border-gray-200 rounded-2xl focus:outline-none focus:border-[#1e3a8a]"
      />
      <span class="absolute left-3 top-2.5 text-gray-400">🔍</span>
    </div>
  );
}