// web/src/components/admin/psicologos/edit/sections/AccountSection.tsx

import { Field, IC, SectionCard } from "../EditPrimitives";
import type { EditFormState } from "../types";

interface Props {
  form: EditFormState;
  setForm: (key: keyof EditFormState, value: any) => void;
}

export function AccountSection(props: Props) {
  return (
    <SectionCard title="Cuenta y Acceso" accent="border-colpsi-yellow">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-5">
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
      </div>
    </SectionCard>
  );
}