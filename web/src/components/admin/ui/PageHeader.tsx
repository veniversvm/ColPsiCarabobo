// web/src/components/admin/ui/PageHeader.tsx
// Cabecera de página del admin: breadcrumbs + título sobrio + acciones.
import { JSX, Show } from "solid-js";
import { Icon } from "./icons";

export interface Crumb {
  label: string;
  href?: string;
}

interface PageHeaderProps {
  title: string;
  crumbs?: Crumb[];
  description?: string;
  actions?: JSX.Element;
  class?: string;
}

export function PageHeader(props: PageHeaderProps) {
  return (
    <header class={`flex items-end justify-between gap-4 pb-4 border-b border-colpsi-border ${props.class ?? ""}`}>
      <div class="min-w-0">
        <Show when={props.crumbs && props.crumbs.length > 0}>
          <nav class="text-xs text-colpsi-muted mb-1.5 flex items-center gap-1.5">
            {props.crumbs!.map((c, i) => (
              <span class="flex items-center gap-1.5">
                {i > 0 && <Icon name="chevronRight" class="w-3.5 h-3.5 text-slate-300" />}
                {c.href
                  ? <a href={c.href} class="hover:text-colpsi-blue transition-colors">{c.label}</a>
                  : <span class="text-colpsi-text font-medium">{c.label}</span>}
              </span>
            ))}
          </nav>
        </Show>
        <h1 class="text-lg font-semibold text-colpsi-text leading-tight truncate">{props.title}</h1>
        <Show when={props.description}>
          <p class="text-sm text-colpsi-muted mt-0.5">{props.description}</p>
        </Show>
      </div>
      <Show when={props.actions}>
        <div class="flex items-center gap-2 shrink-0">{props.actions}</div>
      </Show>
    </header>
  );
}