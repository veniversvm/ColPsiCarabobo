// web/src/components/psi/profile/AvatarUploader.tsx
import { Show } from "solid-js";

interface AvatarUploaderProps {
  currentAvatarUrl?: string;
  avatarFile: File | null;
  onFileChange: (file: File | null) => void;
}

export function AvatarUploader(props: AvatarUploaderProps) {
  const previewUrl = () => {
    if (props.avatarFile) {
      return URL.createObjectURL(props.avatarFile);
    }
    return props.currentAvatarUrl 
      ? `http://localhost:9000/colpsi-bucket/${props.currentAvatarUrl}`
      : null;
  };

  return (
    <section class="bg-white rounded-[2.5rem] p-8 shadow-premium border border-gray-100 flex flex-col items-center">
      <div class="relative group">
        <div class="w-32 h-32 rounded-full overflow-hidden border-4 border-colpsi-yellow bg-gray-50 shadow-inner">
          <Show 
            when={previewUrl()} 
            fallback={<div class="w-full h-full flex items-center justify-center text-4xl">👤</div>}
          >
            <img src={previewUrl()!} class="w-full h-full object-cover" alt="Avatar" />
          </Show>
        </div>
        <label class="absolute bottom-0 right-0 bg-colpsi-blue text-white p-2 rounded-full cursor-pointer shadow-lg hover:bg-colpsi-yellow hover:text-colpsi-blue transition-all">
          <input 
            type="file" 
            class="sr-only" 
            accept="image/jpeg, image/png" 
            onChange={(e) => props.onFileChange(e.currentTarget.files?.[0] || null)} 
          />
          📷
        </label>
      </div>
      <p class="text-[10px] font-bold text-gray-400 uppercase mt-4 tracking-widest">Foto de Perfil</p>
    </section>
  );
}