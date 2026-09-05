// web/src/lib/notifications.tsx
//
// Notificador de nuevas notificaciones del portal psi.
// - Sondea `/notifications/psi-user` cada 5 minutos en el navegador.
// - Mantiene una línea base de ids ya vistos en `sessionStorage` (clave
//   `notif_seen_<userId>`) para no repetir el sonido de comunicados viejos.
// - El sonido solo suena cuando la pestaña está visible (`document.visibilityState`),
//   respetando las políticas de autoplay del navegador: se desbloquea el objeto
//   Audio con el primer gesto del usuario (pointerdown/keydown).
// - Es puramente aditivo: no altera el conteo ni el listado (cada página hace su fetch).
import { createContext, useContext, JSX, onMount, onCleanup } from "solid-js";
import { isServer } from "solid-js/web";
import { apiGet, ApiError } from "~/lib/api";
import { useAuth } from "~/lib/auth";
import type { Notification } from "~/types/notifications";

const POLL_INTERVAL_MS = 5 * 60 * 1000;
const SEEN_PREFIX = "notif_seen_";
const SOUND_URL = "/sounds/notificacion.mp3";
const LIST_LIMIT = 20;

// ── Utilerías de linea base ───────────────────────────────────────────────────
function loadSeen(userId: string): Set<string> {
  try {
    const raw = sessionStorage.getItem(SEEN_PREFIX + userId);
    if (!raw) return new Set();
    const arr = JSON.parse(raw);
    return new Set(Array.isArray(arr) ? arr : []);
  } catch {
    return new Set();
  }
}

function saveSeen(userId: string, ids: Set<string>) {
  try {
    sessionStorage.setItem(SEEN_PREFIX + userId, JSON.stringify([...ids]));
  } catch {
    // sessionStorage lleno o bloqueado: no bloqueamos el sondeo.
  }
}

// ── Audio (desbloqueado con el primer gesto del usuario) ──────────────────────
let audio: HTMLAudioElement | null = null;

function ensureAudio(): HTMLAudioElement {
  if (!audio) audio = new Audio(SOUND_URL);
  audio.volume = 0.5;
  return audio;
}

function playSound() {
  if (typeof document === "undefined" || document.visibilityState !== "visible") return;
  const a = ensureAudio();
  // Restarte por si ya sonó (permite avisos seguidos).
  a.currentTime = 0;
  a.play().catch(() => {
    // Autoplay bloqueado: se reintenta en el siguiente gesto (unlock()).
  });
}

function unlockAudio() {
  const a = ensureAudio();
  a.load();
  window.removeEventListener("pointerdown", unlockAudio);
  window.removeEventListener("keydown", unlockAudio);
}

// ── Contexto (opcional: expone el conteo para badges futuros) ────────────────
interface NotificationsContextValue {
  unreadCount: () => number;
}

const NotificationsContext = createContext<NotificationsContextValue>({
  unreadCount: () => 0,
});

export function useNotifications() {
  return useContext(NotificationsContext);
}

export function NotificationsProvider(props: { children: JSX.Element }) {
  const { user } = useAuth();

  onMount(() => {
    if (isServer) return;

    const uid = user()?.id;
    if (!uid) return;

    window.addEventListener("pointerdown", unlockAudio, { once: true });
    window.addEventListener("keydown", unlockAudio, { once: true });

    const seen = loadSeen(uid);
    let baseline = seen;
    let firstRun = seen.size === 0;

    const poll = async () => {
      try {
        const res = await apiGet<{ data: Notification[] }>(
          `/notifications/psi-user?page=1&limit=${LIST_LIMIT}`,
        );
        const ids = (res?.data ?? []).map((n) => n.id);

        // En la primera pasada solo se siembra la línea base: no se sondea nada
        // que ya existía antes de abrir el portal.
        if (!firstRun && ids.some((id) => !baseline.has(id))) {
          playSound();
        }

        firstRun = false;
        ids.forEach((id) => baseline.add(id));
        saveSeen(uid, baseline);
      } catch (error) {
        // 401/403/404: sesión revocada o token ausente → no sonar.
        // 503/red: blip de conectividad → no sonar.
        if (error instanceof ApiError && (error.status === 401 || error.status === 403 || error.status === 404)) {
          clearInterval(interval);
        }
      }
    };

    poll();
    const interval = setInterval(poll, POLL_INTERVAL_MS);
    onCleanup(() => clearInterval(interval));
  });

  return (
    <NotificationsContext.Provider value={{ unreadCount: () => 0 }}>
      {props.children}
    </NotificationsContext.Provider>
  );
}