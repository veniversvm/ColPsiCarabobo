// web/src/components/admin/psicologos/edit/sections/LegalIdentitySection.tsx

import { Field, IC } from "../EditPrimitives";
import type { EditFormState } from "../types";

interface Props {
  form: EditFormState;
  setForm: (key: keyof EditFormState, value: any) => void;
}

export function LegalIdentitySection(props: Props) {
  return (
    <div class="grid grid-cols-1 md:grid-cols-4 gap-5">

        <div class="md:col-span-2">
          <Field label="Nombres">
            <div class="flex gap-2">
              <input type="text" required placeholder="Primer Nombre"
                value={props.form.first_name}
                onInput={(e) => props.setForm("first_name", e.currentTarget.value)}
                class={IC}
              />
              <input type="text" placeholder="Segundo Nombre"
                value={props.form.second_name}
                onInput={(e) => props.setForm("second_name", e.currentTarget.value)}
                class={IC}
              />
            </div>
          </Field>
        </div>

        <div class="md:col-span-2">
          <Field label="Apellidos">
            <div class="flex gap-2">
              <input type="text" required placeholder="Primer Apellido"
                value={props.form.last_name}
                onInput={(e) => props.setForm("last_name", e.currentTarget.value)}
                class={IC}
              />
              <input type="text" placeholder="Segundo Apellido"
                value={props.form.second_last_name}
                onInput={(e) => props.setForm("second_last_name", e.currentTarget.value)}
                class={IC}
              />
            </div>
          </Field>
        </div>

        <div>
          <Field label="Cédula de Identidad">
            <div class="flex gap-2">
              <select
                value={props.form.nationality}
                onChange={(e) => props.setForm("nationality", e.currentTarget.value)}
                class={`${IC} w-24`}
              >
                <option value="V">V</option>
                <option value="E">E</option>
              </select>
              <input type="number" required
                value={props.form.ci}
                onInput={(e) => props.setForm("ci", e.currentTarget.value)}
                class={IC}
              />
            </div>
          </Field>
        </div>

        <div>
          <Field label="Nro. FPV ★">
            <input
              type="number"
              required
              value={props.form.fpv}
              onInput={(e) => props.setForm("fpv", e.currentTarget.value)}
              class={`${IC} bg-yellow-50 border-yellow-300 font-bold text-yellow-800`}
            />
          </Field>
        </div>

        <div>
          <Field label="Nº de Control (interno)">
            <input
              type="text"
              readOnly
              value={props.form.control_number || "—"}
              title="Número de control interno asignado desde el Excel maestro. Solo visible para administradores."
              class={`${IC} bg-gray-100 text-gray-600 cursor-not-allowed`}
            />
          </Field>
        </div>

        <div>
          <Field label="Género">
            <select
              value={props.form.genre}
              onChange={(e) => props.setForm("genre", e.currentTarget.value)}
              class={IC}
            >
              <option value="M">Masculino</option>
              <option value="F">Femenino</option>
            </select>
          </Field>
        </div>

        <div>
          <Field label="Fecha de Nacimiento">
            <input type="date" required
              value={props.form.born_date}
              onInput={(e) => props.setForm("born_date", e.currentTarget.value)}
              class={IC}
            />
          </Field>
        </div>

      </div>
  );
}