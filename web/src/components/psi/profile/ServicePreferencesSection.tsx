// web/src/components/psi/profile/ServicePreferencesSection.tsx
import { ToggleSwitch } from "~/components/ui/ToggleSwitch";

interface ServicePreferencesSectionProps {
  // Modalidad de servicio (cómo atiende; puede ser combinación)
  serviceModalityPresencial: boolean;
  serviceModalityDistance: boolean;
  serviceModalityTelephone: boolean;
  // Opt-in Privacy Shield: mostrar la modalidad en el directorio público
  showServiceModality: boolean;
  // Opt-in de aviso de cumpleaños a la administración
  birthdayNotification: boolean;

  onServiceModalityPresencialChange: (value: boolean) => void;
  onServiceModalityDistanceChange: (value: boolean) => void;
  onServiceModalityTelephoneChange: (value: boolean) => void;
  onShowServiceModalityChange: (value: boolean) => void;
  onBirthdayNotificationChange: (value: boolean) => void;
}

const boxClass =
  "space-y-3 bg-colpsi-bg/50 p-4 rounded-lg border border-colpsi-border";
const boxTitleClass =
  "text-[11px] font-semibold text-colpsi-text uppercase tracking-wide border-b border-colpsi-border pb-2";

export function ServicePreferencesSection(props: ServicePreferencesSectionProps) {
  return (
    <section>
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        {/* ── Modalidad de servicio ────────────────────────────────────── */}
        <div class={boxClass}>
          <h3 class={boxTitleClass}>
            Modalidad de Servicio
          </h3>
          <p class="text-xs text-colpsi-muted">
            Indique cómo presta atención a sus pacientes. Puede marcar varias
            opciones. Si ninguna está activa, se mostrará que actualmente no
            presta servicio.
          </p>
          <ToggleSwitch
            label="Presencial"
            checked={props.serviceModalityPresencial}
            onChange={props.onServiceModalityPresencialChange}
          />
          <ToggleSwitch
            label="A distancia (en línea)"
            checked={props.serviceModalityDistance}
            onChange={props.onServiceModalityDistanceChange}
          />
          <ToggleSwitch
            label="Telefónica"
            checked={props.serviceModalityTelephone}
            onChange={props.onServiceModalityTelephoneChange}
          />
        </div>

        {/* ── Preferencias ─────────────────────────────────────────────── */}
        <div class={boxClass}>
          <h3 class={boxTitleClass}>
            Preferencias
          </h3>
          <p class="text-xs text-colpsi-muted">
            Controle qué información se comparte con el Colegio y con la
            comunidad.
          </p>
          <ToggleSwitch
            label="Mostrar mi modalidad en el directorio público"
            checked={props.showServiceModality}
            onChange={props.onShowServiceModalityChange}
          />
          <ToggleSwitch
            label="Autorizar aviso de cumpleaños a la administración"
            checked={props.birthdayNotification}
            onChange={props.onBirthdayNotificationChange}
          />
        </div>
      </div>
    </section>
  );
}