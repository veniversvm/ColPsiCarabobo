// web/src/components/admin/psicologos/edit/sections/AccountSection.tsx

import { Show, createMemo } from "solid-js";
import QRCodeGenerator from "~/components/psi/profile/QrCode";
import { bucketUrl } from "~/lib/bucket";
import { Field, IC, SectionCard } from "../EditPrimitives";
import type { EditFormState } from "../types";

interface Props {
  url: string;
  form: EditFormState;
  setForm: (key: keyof EditFormState, value: any) => void;
  // --- Nuevas Props para archivos ---
  avatarFile: File | null;
  onAvatarChange: (file: File | null) => void;
  // --- Foto actual en la DB (URL pública) y acción de eliminación ---
  pictureUrl: string;
  onDeletePicture: () => void;
  // --- Reinicio de clave por administración ---
  onResetPassword: () => void;
  resettingPassword: boolean;
}

export function AccountSection(props: Props) {
  // Generar una URL de vista previa si hay un archivo seleccionado
  const previewUrl = createMemo(() => {
    if (props.avatarFile) {
      return URL.createObjectURL(props.avatarFile);
    }
    return bucketUrl(props.pictureUrl) || "";
  });

  const handleFileChange = (e: Event) => {
    const file = (e.currentTarget as HTMLInputElement).files?.[0] || null;
    props.onAvatarChange(file);
  };

  return (
    <SectionCard title="Cuenta y Perfil Visual" accent="border-colpsi-yellow">
      <div class="flex flex-col xl:flex-row gap-10">
        
        {/* 1. Gestión de Foto de Perfil */}
        <div class="flex flex-col items-center space-y-4">
          <label class="text-[10px] font-black text-gray-400 uppercase tracking-widest">
            Foto de Perfil
          </label>
          <div class="relative group">
            <div class="w-32 h-32 rounded-3xl overflow-hidden border-4 border-gray-50 shadow-md bg-gray-100 flex items-center justify-center relative">
              <Show 
                when={previewUrl()} 
                fallback={<span class="text-4xl text-gray-300">👤</span>}
              >
                <img src={previewUrl()} alt="Avatar" class="w-full h-full object-cover" />
              </Show>
              
              {/* Overlay de carga */}
              <label class="absolute inset-0 bg-blue-900/60 opacity-0 group-hover:opacity-100 transition-opacity flex flex-col items-center justify-center cursor-pointer text-white p-2 text-center">
                <span class="text-lg">📷</span>
                <span class="text-[9px] font-bold uppercase leading-tight">Cambiar Foto</span>
                <input type="file" accept="image/*" class="hidden" onChange={handleFileChange} />
              </label>
            </div>
            
            <Show when={props.avatarFile}>
              <button 
                onClick={() => props.onAvatarChange(null)}
                class="absolute -top-2 -right-2 w-6 h-6 bg-red-500 text-white rounded-full text-xs shadow-lg flex items-center justify-center hover:bg-red-600 transition-colors"
                title="Quitar imagen seleccionada"
              >
                ✕
              </button>
            </Show>
          </div>
          <p class="text-[9px] text-gray-400 font-medium max-w-[120px] text-center">
            Formatos: JPG, PNG. Máx 2MB.
          </p>
          <Show when={previewUrl() && !props.avatarFile}>
            <button
              onClick={() => props.onDeletePicture()}
              class="text-[10px] font-black uppercase tracking-widest bg-red-50 text-red-600 border border-red-200 rounded-lg px-3 py-1.5 hover:bg-red-100 transition-colors"
              title="Eliminar la foto de perfil del psicólogo"
            >
              🗑 Eliminar foto
            </button>
          </Show>
        </div>

        {/* 2. Inputs de Cuenta */}
        <div class="flex-1 grid grid-cols-1 md:grid-cols-2 gap-5">
          <Field label="Nombre de Usuario (Login)">
            <input
              type="text"
              value={props.form.username}
              onInput={(e) => props.setForm("username", e.currentTarget.value)}
              class={IC}
            />
          </Field>
          <Field label="Email Institucional (Login)">
            <input
              type="email"
              value={props.form.email}
              onInput={(e) => props.setForm("email", e.currentTarget.value)}
              class={IC}
            />
          </Field>
          
          <div class="md:col-span-2 p-6 bg-blue-50/50 rounded-2xl border border-blue-100 mt-2">
            <p class="text-sm text-blue-700 leading-relaxed">
              <span class="font-black uppercase mr-1">Aviso:</span>
              Cualquier cambio en el Email o Username afectará las credenciales de acceso del psicólogo.
              Si el usuario está logueado, su sesión se invalidará.
            </p>

            <div class="flex flex-wrap items-center justify-between gap-4 mt-5 pt-5 border-t border-blue-100">
              <p class="text-sm text-blue-700 leading-relaxed">
                <span class="font-black uppercase mr-1">Clave de acceso:</span>
                Reinicia la contraseña para cerrar todas sus sesiones y enviarle
                una temporal por correo.
              </p>
              <button
                type="button"
                onClick={() => props.onResetPassword()}
                disabled={props.resettingPassword}
                class="inline-flex items-center gap-2 text-[11px] font-black uppercase tracking-widest bg-red-600 text-white rounded-lg px-4 py-2.5 hover:bg-red-700 active:scale-95 transition-all disabled:opacity-50"
              >
                <Show
                  when={props.resettingPassword}
                  fallback={<span>🔑 Reiniciar clave</span>}
                >
                  <span class="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                  <span>Procesando...</span>
                </Show>
              </button>
            </div>
          </div>
        </div>

        {/* 3. Generador de QR lateral */}
        <div class="flex flex-col items-center justify-start lg:border-l lg:border-gray-100 lg:pl-8">
          <label class="text-[10px] font-black text-gray-400 uppercase tracking-widest mb-4">
            Acceso Directo (QR)
          </label>
          <QRCodeGenerator url={props.url} />
          <p class="text-[9px] text-gray-400 mt-4 text-center max-w-[150px]">
            Escanea para ver la <br/> ficha pública actual.
          </p>
        </div>

      </div>
    </SectionCard>
  );
}