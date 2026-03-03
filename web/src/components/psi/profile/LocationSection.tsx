import { InputField } from "./InputField";

// web/src/components/psi/profile/LocationSection.tsx
interface LocationSectionProps {
  // Carabobo
  municipalityCarabobo: string;
  phoneCarabobo: string;
  celPhoneCarabobo: string;
  
  // Exterior
  stateOutside: string;
  municipalityOutside: string;
  phoneOutside: string;
  celPhoneOutside: string;
  
  onMunicipalityCaraboboChange: (value: string) => void;
  onPhoneCaraboboChange: (value: string) => void;
  onCelPhoneCaraboboChange: (value: string) => void;
  
  onStateOutsideChange: (value: string) => void;
  onMunicipalityOutsideChange: (value: string) => void;
  onPhoneOutsideChange: (value: string) => void;
  onCelPhoneOutsideChange: (value: string) => void;
}

export function LocationSection(props: LocationSectionProps) {
  return (
    <section class="bg-white rounded-[2.5rem] p-6 md:p-8 shadow-premium border border-gray-100">
      <h2 class="text-xl font-black text-colpsi-blue mb-6 border-l-4 border-colpsi-yellow pl-3">
        Ubicación en Carabobo
      </h2>
      
      <p class="text-sm text-gray-500 mb-8 bg-gray-50 p-4 rounded-2xl border border-gray-100 leading-relaxed italic">
        Los teléfonos de esta sección son de carácter privado, cuya la finalidad de que el Colegio 
        de Psicólogos del Estado Carabobo pueda contactar al psicólogo en caso de necesidad.
        <br/>
        El municipio es información pública.
      </p>

      <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <InputField
          label="Municipio"
          value={props.municipalityCarabobo}
          onInput={props.onMunicipalityCaraboboChange}
        />
        <InputField
          label="Teléfono Fijo"
          type="tel"
          value={props.phoneCarabobo}
          onInput={props.onPhoneCaraboboChange}
        />
        <InputField
          label="Celular Secundario"
          type="tel"
          value={props.celPhoneCarabobo}
          onInput={props.onCelPhoneCaraboboChange}
        />
      </div>

      <h2 class="text-xl font-black text-colpsi-blue mb-6 border-l-4 border-gray-300 pl-3 pt-4 border-t border-gray-50">
        Exterior u Otros Estados
      </h2>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <InputField
          label="Estado / Región / País"
          value={props.stateOutside}
          onInput={props.onStateOutsideChange}
        />
        <InputField
          label="Ciudad / Municipio"
          value={props.municipalityOutside}
          onInput={props.onMunicipalityOutsideChange}
        />
        <InputField
          label="Teléfono Fijo (Internacional)"
          type="tel"
          value={props.phoneOutside}
          onInput={props.onPhoneOutsideChange}
        />
        <InputField
          label="Celular (Internacional)"
          type="tel"
          value={props.celPhoneOutside}
          onInput={props.onCelPhoneOutsideChange}
        />
      </div>
    </section>
  );
}