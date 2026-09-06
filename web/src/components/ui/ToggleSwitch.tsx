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
            'block w-11 h-6 rounded-full transition-colors duration-300 shadow-inner border border-gray-200': true,
            'bg-colpsi-blue border-colpsi-blue': !!props.checked,
            'bg-gray-200': !props.checked
          }}
        ></div>
        <div 
          classList={{
            'dot absolute left-1 top-1 bg-white w-4 h-4 rounded-full transition-transform duration-300 shadow-md': true,
            'translate-x-5': !!props.checked
          }}
        ></div>
      </div>
      <div class="ml-4 text-xs font-bold text-colpsi-muted group-hover:text-colpsi-blue transition-colors select-none">
        {props.label}
      </div>
    </label>
  );
}