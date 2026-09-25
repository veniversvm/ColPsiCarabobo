// web/src/components/admin/dashboard/BirthdayBanner.tsx

import { Show, For } from "solid-js";
import { createResource } from "solid-js";
import { apiGet } from "~/lib/api";
import { Badge } from "~/components/admin/ui/Badge";

export interface BirthdayPerson {
  id: string;
  first_name: string;
  last_name: string;
  fpv: number;
  is_active: boolean;
  month: number;
  day: number;
}

interface BirthdayResponse {
  range: string;
  data: BirthdayPerson[];
  total: number;
}

export function BirthdayBanner() {
  const [birthdays] = createResource<BirthdayResponse>(() =>
    apiGet("/admin/psi/birthdays?range=week")
  );

  const todays = () => {
    const data = birthdays()?.data ?? [];
    const now = new Date();
    const todayMonth = now.getMonth() + 1;
    const todayDay = now.getDate();
    return data.filter(
      (b) => b.month === todayMonth && b.day === todayDay,
    );
  };

  const upcoming = () =>
    (birthdays()?.data ?? []).filter((b) => {
      const now = new Date();
      const todayMonth = now.getMonth() + 1;
      const todayDay = now.getDate();
      return !(b.month === todayMonth && b.day === todayDay);
    });

  return (
    <Show when={(birthdays()?.data?.length ?? 0) > 0}>
      <div class="border border-colpsi-border rounded-lg bg-white p-4">
        <div class="flex items-center gap-2 mb-3">
          <Badge tone="info">Cumpleaños de la semana</Badge>
          <span class="text-[11px] text-colpsi-muted">{birthdays()?.range}</span>
        </div>
        <Show when={todays().length > 0}>
          <p class="text-sm font-semibold text-colpsi-text">Hoy cumplen años:</p>
          <div class="flex flex-wrap gap-2 mt-2">
            <For each={todays()}>
              {(b) => (
                <Badge tone="info">
                  {b.first_name} {b.last_name} · FPV {b.fpv}
                </Badge>
              )}
            </For>
          </div>
        </Show>
        <Show when={upcoming().length > 0}>
          <p class="text-sm text-colpsi-muted mt-2">
            Próximos {upcoming().length} en la semana:{" "}
            {upcoming()
              .slice(0, 6)
              .map((b) => `${b.first_name} ${b.last_name}`)
              .join(", ")}
          </p>
        </Show>
      </div>
    </Show>
  );
}