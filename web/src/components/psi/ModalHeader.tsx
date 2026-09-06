// web/src/components/ui/ModalHeader.tsx
// Componente UI reutilizable
interface ModalHeaderProps {
  title: string;
  onClose: () => void;
}

export function ModalHeader(props: ModalHeaderProps) {
  return (
    <div class="sticky top-0 bg-white border-b border-colpsi-border px-4 py-3 flex items-center justify-between">
      <h3 class="font-black text-colpsi-blue text-sm md:text-base uppercase tracking-wider">
        {props.title}
      </h3>
      <button 
        onClick={props.onClose}
        class="w-8 h-8 rounded-full bg-gray-100 hover:bg-gray-200 flex items-center justify-center transition-colors"
      >
        ✕
      </button>
    </div>
  );
}