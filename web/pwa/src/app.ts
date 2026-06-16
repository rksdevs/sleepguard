import {
  fetchEvents,
  fetchPairings,
  fetchStatus,
  pairDevice,
  requestCapture,
  runCleanup,
  unpairDevice,
} from "./api";
import {
  loadSnapshotGallery,
  openSnapshotLightbox,
  type SnapshotThumb,
} from "./snapshots";
import {
  clearPairingId,
  loadPairingId,
  pushSupported,
  savePairingId,
  subscribeForPush,
} from "./push";
import {
  clearSettings,
  defaultApiBase,
  loadSettings,
  saveSettings,
} from "./settings";
import type { DeviceStatus, MotionEvent, PairedClient, Settings } from "./types";

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

interface DashboardView {
  settings: Settings;
  status: DeviceStatus;
  events: MotionEvent[];
  pairings: PairedClient[];
  pairingId: string | null;
  snapshots: SnapshotThumb[];
  error: string | null;
  notice: string | null;
}

function renderPairingRows(pairings: PairedClient[], localId: string | null): string {
  if (pairings.length === 0) {
    return `<p class="muted" style="margin:0">No phones paired yet.</p>`;
  }
  return `<ul class="pairing-list">${pairings
    .map(
      (p) => `
      <li>
        <span>${p.name}${p.id === localId ? ' <em class="muted">(this device)</em>' : ""}</span>
        <span class="muted">${formatTime(p.created_at)}</span>
      </li>`,
    )
    .join("")}</ul>`;
}

function renderSnapshotCarousel(snapshots: SnapshotThumb[]): string {
  if (snapshots.length === 0) {
    return `<p class="muted" style="margin:0">No snapshots yet.</p>`;
  }
  return `
    <p class="muted" style="margin:0 0 0.5rem">Recent captures — tap to enlarge or download.</p>
    <div class="snapshot-carousel">
      ${snapshots
        .map(
          (s) => `
        <button type="button" class="snapshot-thumb" data-snapshot-id="${s.id}" aria-label="View snapshot">
          <img src="${s.url}" alt="" loading="lazy" />
          <span class="snapshot-thumb-time">${formatTime(s.capturedAt)}</span>
        </button>`,
        )
        .join("")}
    </div>`;
}

function renderDashboard(root: HTMLElement, view: DashboardView): void {
  const rows =
    view.events.length === 0
      ? `<tr><td colspan="4" class="muted">No events yet — motion will appear here.</td></tr>`
      : view.events
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

  const pushOk = pushSupported();
  const pairedHere = view.pairingId
    ? view.pairings.some((p) => p.id === view.pairingId)
    : false;

  root.innerHTML = `
    <div class="toolbar">
      <div>
        <div class="brand">
          ${BRAND_ICON}
          <h1>SleepGuard</h1>
        </div>
        <p class="subtitle" style="margin-bottom:0">${view.status.name}
          <span class="badge ${view.status.online ? "online" : "offline"}">${view.status.online ? "online" : "offline"}</span>
        </p>
      </div>
      <button type="button" class="secondary" id="logout-btn">Disconnect</button>
    </div>

    ${view.error ? `<p class="error">${view.error}</p>` : ""}
    ${view.notice ? `<p class="notice">${view.notice}</p>` : ""}

    <div class="stats">
      <div class="card stat"><span>Total motion</span><strong>${view.status.event_count}</strong></div>
      <div class="card stat"><span>In view</span><strong>${view.events.length}</strong></div>
      <div class="card stat"><span>Last seen</span><strong style="font-size:0.88rem;font-weight:500;color:var(--text)">${view.status.last_seen_at ? formatTime(view.status.last_seen_at) : "—"}</strong></div>
    </div>

    <div class="card">
      <div class="toolbar" style="margin-bottom:0.5rem">
        <span class="toolbar-title">Camera</span>
      </div>
      <p class="muted" style="margin-top:0">Capture a still from the nursery whenever you want (Pi must be online).</p>
      <div class="actions">
        <button type="button" id="capture-btn" ${view.status.online ? "" : "disabled"}>Capture image</button>
      </div>
      ${renderSnapshotCarousel(view.snapshots)}
    </div>

    <div class="card">
      <div class="toolbar" style="margin-bottom:0.5rem">
        <span class="toolbar-title">Notifications</span>
      </div>
      ${
        pushOk
          ? `<p class="muted" style="margin-top:0">Push alert every 3 motion cycles (3, 6, 9…). Tap Capture when you want a photo.</p>
      <div class="actions">
        ${
          pairedHere
            ? `<button type="button" class="secondary" id="unpair-btn">Turn off notifications</button>`
            : `<button type="button" id="pair-btn">Enable notifications</button>`
        }
      </div>
      ${renderPairingRows(view.pairings, view.pairingId)}`
          : `<p class="muted" style="margin:0">Push not supported in this browser.</p>`
      }
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

    <div class="card admin-card">
      <div class="toolbar" style="margin-bottom:0.5rem">
        <span class="toolbar-title">Data cleanup</span>
      </div>
      <p class="muted" style="margin-top:0">Remove events and snapshots older than 24 hours.</p>
      <button type="button" class="secondary" id="cleanup-btn">Run cleanup now</button>
    </div>
  `;

  root.querySelector("#logout-btn")?.addEventListener("click", () => {
    clearSettings();
    clearPairingId();
    window.location.reload();
  });

  root.querySelectorAll<HTMLButtonElement>(".snapshot-thumb").forEach((btn) => {
    btn.addEventListener("click", () => {
      const id = btn.dataset.snapshotId;
      const thumb = view.snapshots.find((s) => s.id === id);
      if (thumb) {
        openSnapshotLightbox(thumb);
      }
    });
  });

  root.querySelector("#capture-btn")?.addEventListener("click", () => {
    void (async () => {
      try {
        await requestCapture(view.settings);
        view.notice =
          "Capture queued — image should appear within about 10 seconds.";
        view.error = null;
      } catch (err) {
        view.notice = err instanceof Error ? err.message : "Capture request failed.";
      }
      renderDashboard(root, view);
    })();
  });

  root.querySelector("#pair-btn")?.addEventListener("click", () => {
    void (async () => {
      const perm = await Notification.requestPermission();
      if (perm !== "granted") {
        view.notice = "Notification permission denied.";
        renderDashboard(root, view);
        return;
      }
      try {
        const sub = await subscribeForPush(view.settings.apiBase, view.settings.readKey);
        const client = await pairDevice(view.settings, sub);
        savePairingId(client.id);
        view.pairingId = client.id;
        view.pairings = await fetchPairings(view.settings);
        view.notice = "Notifications enabled for this phone.";
        view.error = null;
      } catch (err) {
        view.notice =
          err instanceof Error ? err.message : "Could not enable notifications.";
      }
      renderDashboard(root, view);
    })();
  });

  root.querySelector("#unpair-btn")?.addEventListener("click", () => {
    void (async () => {
      if (!view.pairingId) {
        return;
      }
      try {
        await unpairDevice(view.settings, view.pairingId);
        clearPairingId();
        view.pairingId = null;
        view.pairings = await fetchPairings(view.settings);
        view.notice = "Notifications turned off.";
      } catch (err) {
        view.notice = err instanceof Error ? err.message : "Unpair failed.";
      }
      renderDashboard(root, view);
    })();
  });

  root.querySelector("#cleanup-btn")?.addEventListener("click", () => {
    void (async () => {
      try {
        const result = await runCleanup(view.settings);
        view.notice = `Cleanup done — ${result.events_deleted} events, ${result.snapshots_deleted} snapshots removed.`;
        view.error = null;
        const [status, { events }] = await Promise.all([
          fetchStatus(view.settings),
          fetchEvents(view.settings),
        ]);
        view.status = status;
        view.events = events;
      } catch (err) {
        view.notice = err instanceof Error ? err.message : "Cleanup failed.";
      }
      renderDashboard(root, view);
    })();
  });
}

export function mountApp(root: HTMLElement): void {
  let settings = loadSettings();
  let timer: number | undefined;
  let notice: string | null = null;
  let snapshots: SnapshotThumb[] = [];

  const tick = async () => {
    if (!settings) {
      return;
    }
    try {
      const [status, { events }, pairings, gallery] = await Promise.all([
        fetchStatus(settings),
        fetchEvents(settings),
        fetchPairings(settings).catch(() => [] as PairedClient[]),
        loadSnapshotGallery(settings, 5).catch(() => snapshots),
      ]);
      snapshots = gallery;
      renderDashboard(root, {
        settings,
        status,
        events,
        pairings,
        pairingId: loadPairingId(),
        snapshots: gallery,
        error: null,
        notice,
      });
      notice = null;
    } catch (err) {
      const message = err instanceof Error ? err.message : "request failed";
      renderDashboard(root, {
        settings,
        status: {
          id: settings.deviceId,
          name: settings.deviceId,
          event_count: 0,
          online: false,
        },
        events: [],
        pairings: [],
        pairingId: loadPairingId(),
        snapshots,
        error: message,
        notice,
      });
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
