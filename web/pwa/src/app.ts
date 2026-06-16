import { fetchEvents, fetchStatus } from "./api";
import {
  clearSettings,
  defaultApiBase,
  loadSettings,
  saveSettings,
} from "./settings";
import type { DeviceStatus, MotionEvent, Settings } from "./types";

const POLL_MS = 3000;

const BRAND_ICON = `<svg class="brand-icon" viewBox="0 0 128 128" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
  <rect width="128" height="128" rx="28" fill="#f3efe9" stroke="#e0d8ce" stroke-width="2"/>
  <circle cx="88" cy="36" r="10" fill="#d4a574"/>
  <path d="M64 28c-18 0-32 14-32 32 0 12 6.5 22.5 16 28.5" stroke="#d4a574" stroke-width="5" stroke-linecap="round"/>
  <circle cx="52" cy="72" r="14" stroke="#3d8b62" stroke-width="4"/>
</svg>`;

function formatTime(iso: string): string {
  try {
    return new Date(iso).toLocaleString();
  } catch {
    return iso;
  }
}

function patternPill(pattern: string): string {
  if (!pattern) {
    return "—";
  }
  const cls = `pattern pattern-${pattern}`;
  return `<span class="${cls}">${pattern}</span>`;
}

function renderLogin(root: HTMLElement, onSave: (s: Settings) => void): void {
  root.innerHTML = `
    <div class="page-login">
      <div class="brand">
        ${BRAND_ICON}
        <h1>SleepGuard</h1>
      </div>
      <p class="subtitle">Quiet monitoring for peaceful nights</p>
      <div class="card">
        <form id="login-form">
          <label>API base URL
            <input name="apiBase" type="url" value="${defaultApiBase()}" required />
          </label>
          <label>Device ID
            <input name="deviceId" type="text" value="nursery" required />
          </label>
          <label>Read API key
            <input name="readKey" type="password" autocomplete="off" required />
          </label>
          <button type="submit">Connect</button>
        </form>
        <p class="muted" style="margin-top:1rem;margin-bottom:0">Family access key from server deploy/.env</p>
      </div>
    </div>
  `;

  const form = root.querySelector<HTMLFormElement>("#login-form");
  form?.addEventListener("submit", (e) => {
    e.preventDefault();
    const data = new FormData(form);
    onSave({
      apiBase: String(data.get("apiBase") || "").trim(),
      deviceId: String(data.get("deviceId") || "").trim(),
      readKey: String(data.get("readKey") || "").trim(),
    });
  });
}

function renderDashboard(
  root: HTMLElement,
  settings: Settings,
  status: DeviceStatus,
  events: MotionEvent[],
  error: string | null,
): void {
  const rows =
    events.length === 0
      ? `<tr><td colspan="4" class="muted">No events yet — motion will appear here.</td></tr>`
      : events
          .map(
            (e) => `
        <tr>
          <td>${formatTime(e.timestamp)}</td>
          <td class="state-${e.state}">${e.state}</td>
          <td>${patternPill(e.pattern || "")}</td>
          <td>${e.source}</td>
        </tr>`,
          )
          .join("");

  root.innerHTML = `
    <div class="toolbar">
      <div>
        <div class="brand">
          ${BRAND_ICON}
          <h1>SleepGuard</h1>
        </div>
        <p class="subtitle" style="margin-bottom:0">${status.name}
          <span class="badge ${status.online ? "online" : "offline"}">${status.online ? "online" : "offline"}</span>
        </p>
      </div>
      <button type="button" class="secondary" id="logout-btn">Disconnect</button>
    </div>

    ${error ? `<p class="error">${error}</p>` : ""}

    <div class="stats">
      <div class="card stat"><span>Total motion</span><strong>${status.event_count}</strong></div>
      <div class="card stat"><span>In view</span><strong>${events.length}</strong></div>
      <div class="card stat"><span>Last seen</span><strong style="font-size:0.88rem;font-weight:500;color:var(--text)">${status.last_seen_at ? formatTime(status.last_seen_at) : "—"}</strong></div>
    </div>

    <div class="card">
      <div class="toolbar" style="margin-bottom:0.5rem">
        <span class="toolbar-title"><span class="live-dot"></span>Live log</span>
        <span class="muted">every ${POLL_MS / 1000}s</span>
      </div>
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Time</th>
              <th>State</th>
              <th>Pattern</th>
              <th>Source</th>
            </tr>
          </thead>
          <tbody>${rows}</tbody>
        </table>
      </div>
    </div>
  `;

  root.querySelector("#logout-btn")?.addEventListener("click", () => {
    clearSettings();
    window.location.reload();
  });
}

export function mountApp(root: HTMLElement): void {
  let settings = loadSettings();
  let timer: number | undefined;

  const tick = async () => {
    if (!settings) {
      return;
    }
    try {
      const [status, { events }] = await Promise.all([
        fetchStatus(settings),
        fetchEvents(settings),
      ]);
      renderDashboard(root, settings, status, events, null);
    } catch (err) {
      const message = err instanceof Error ? err.message : "request failed";
      renderDashboard(
        root,
        settings,
        {
          id: settings.deviceId,
          name: settings.deviceId,
          event_count: 0,
          online: false,
        },
        [],
        message,
      );
    }
  };

  const start = (s: Settings) => {
    settings = s;
    saveSettings(s);
    void tick();
    timer = window.setInterval(() => void tick(), POLL_MS);
  };

  if (settings) {
    start(settings);
    return;
  }

  renderLogin(root, start);
}
