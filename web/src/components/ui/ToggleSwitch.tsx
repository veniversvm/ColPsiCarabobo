// web/src/components/ui/ToggleSwitch.tsx
// Componente ToggleSwitch Reactivo (Safe for SolidJS Context)
// FIX SENIOR: Componente Reactivo usando classList en lugar de template strings para evitar pérdida de estado.
export function ToggleSwitch(props: { label: string, checked: boolean, onChange: (val: boolean) => void }) {
  return (
    <label class="flex items-center cursor-pointer mt-3 w-max group">
      <div class="relative flex items-center">
        <input
          type="checkbox"
          class="sr-only"
          checked={!!props.checked}
          onChange={(e) => props.onChange(e.currentTarget.checked)}
        />
        <div
          classList={{
            'block w-10 h-6 rounded-full transition-colors duration-300 border': true,
            'bg-colpsi-blue border-colpsi-blue': !!props.checked,
            'bg-slate-200 border-slate-300': !props.checked
          }}
        ></div>
        <div
          classList={{
            'dot absolute left-0.5 top-0.5 bg-white w-5 h-5 rounded-full transition-transform duration-300 shadow-sm': true,
            'translate-x-4': !!props.checked
          }}
        ></div>
      </div>
      <div class="ml-3 text-sm font-medium text-colpsi-text group-hover:text-colpsi-blue transition-colors select-none">
        {props.label}
      </div>
    </label>
  );
}