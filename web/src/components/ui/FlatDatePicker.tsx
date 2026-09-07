// web/src/components/ui/FlatDatePicker.tsx
import { JSX, onCleanup, onMount } from "solid-js";
import flatpickr from "flatpickr";
import { Spanish } from "flatpickr/dist/l10n/es.js";
import "flatpickr/dist/flatpickr.min.css";

interface FlatDatePickerProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  class?: string;
  disabled?: boolean;
  enableTime?: boolean;
  dateFormat?: string;
  minDate?: string;
  maxDate?: string;
}

export default function FlatDatePicker(props: FlatDatePickerProps) {
  let inputRef: HTMLInputElement | undefined;

  onMount(() => {
    if (!import.meta.env.SSR) {
      const fp = flatpickr(inputRef!, {
        locale: Spanish,
        dateFormat: props.dateFormat ?? "Y-m-d",
        enableTime: props.enableTime ?? false,
        time_24hr: true,
        minDate: props.minDate,
        maxDate: props.maxDate,
        defaultDate: props.value || undefined,
        allowInput: false,
        disableMobile: true,
        onChange: (_selectedDates: Date[], dateStr: string) => props.onChange(dateStr),
      });
      onCleanup(() => fp.destroy());
    }
  });

  return (
    <input
      ref={inputRef}
      class={props.class}
      placeholder={props.placeholder}
      disabled={props.disabled}
      readonly
      value={props.value}
      aria-label="Selector de fecha"
    />
  );
}
