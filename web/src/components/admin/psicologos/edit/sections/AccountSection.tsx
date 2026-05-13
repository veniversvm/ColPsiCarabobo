// web/src/components/admin/psicologos/edit/sections/AccountSection.tsx

import QRCodeGenerator from "~/components/psi/profile/QrCode";
import { Field, IC, SectionCard } from "../EditPrimitives";
import type { EditFormState } from "../types";

interface Props {
  url: string;
  form: EditFormState;
  setForm: (key: keyof EditFormState, value: any) => void;
}

export function AccountSection(props: Props) {
  return (
    <SectionCard title="Cuenta y Acceso" accent="border-colpsi-yellow">
      <div class="flex flex-col lg:flex-row gap-8">
        
        {/* Inputs de Cuenta */}
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
          
          <div class="md:col-span-2 p-4 bg-blue-50 rounded-2xl border border-blue-100 mt-2">
            <p class="text-xs text-colpsi-blue font-medium">
              <span class="font-bold">Nota:</span> Estos datos son críticos para el inicio de sesión del psicólogo en la plataforma.
            </p>
          </div>
        </div>

        {/* Generador de QR lateral */}
        <div class="flex justify-center items-start lg:border-l lg:border-gray-100 lg:pl-8">
          <QRCodeGenerator url={props.url} />
        </div>

      </div>
    </SectionCard>
  );
}