// web/src/components/admin/dashboard/BirthdayBanner.tsx

import { Show, For } from "solid-js";
import { createResource } from "solid-js";
import { apiGet } from "~/lib/api";

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
      <div class="bg-gradient-to-r from-pink-500 to-rose-500 rounded-3xl p-6 shadow-lg">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-pink-100 text-xs font-black uppercase tracking-widest">
              🎂 Cumpleaños del agremiado
            </p>
            <div class="mt-3 space-y-1">
              <Show when={todays().length > 0}>
                <p class="text-white text-sm font-black">
                  Hoy cumplen años:
                </p>
                <For each={todays()}>
                  {(b) => (
                    <p class="text-white text-base font-black">
                      {b.first_name} {b.last_name} · FPV {b.fpv}
                    </p>
                  )}
                </For>
              </Show>
              <Show when={upcoming().length > 0}>
                <p class="text-pink-100 text-xs mt-2">
                  Próximos {upcoming().length} en la semana:
                  {upcoming()
                    .slice(0, 6)
                    .map((b) => `${b.first_name} ${b.last_name}`)
                    .join(", ")}
                </p>
              </Show>
            </div>
          </div>
          <div class="text-7xl opacity-20">🎈</div>
        </div>
      </div>
    </Show>
  );
}