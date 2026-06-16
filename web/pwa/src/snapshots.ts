import { fetchSnapshotBlob, fetchSnapshots } from "./api";
import type { Settings } from "./types";

export interface SnapshotThumb {
  id: string;
  capturedAt: string;
  url: string;
}

const cache = new Map<string, SnapshotThumb>();
let lightboxEl: HTMLElement | null = null;

export async function loadSnapshotGallery(
  settings: Settings,
  limit = 5,
): Promise<SnapshotThumb[]> {
  const snaps = await fetchSnapshots(settings, limit);
  const keep = new Set(snaps.map((s) => s.id));

  for (const [id, thumb] of cache) {
    if (!keep.has(id)) {
      URL.revokeObjectURL(thumb.url);
      cache.delete(id);
    }
  }

  const out: SnapshotThumb[] = [];
  for (const snap of snaps) {
    let thumb = cache.get(snap.id);
    if (!thumb) {
      const blob = await fetchSnapshotBlob(settings, snap.id);
      thumb = {
        id: snap.id,
        capturedAt: snap.captured_at,
        url: URL.createObjectURL(blob),
      };
      cache.set(snap.id, thumb);
    }
    out.push(thumb);
  }
  return out;
}

export function openSnapshotLightbox(thumb: SnapshotThumb): void {
  closeSnapshotLightbox();

  const overlay = document.createElement("div");
  overlay.className = "snapshot-lightbox";
  overlay.innerHTML = `
    <div class="snapshot-lightbox-backdrop" data-close="true"></div>
    <div class="snapshot-lightbox-panel" role="dialog" aria-modal="true" aria-label="Snapshot preview">
      <button type="button" class="snapshot-lightbox-close" aria-label="Close">&times;</button>
      <img src="${thumb.url}" alt="Nursery snapshot" />
      <div class="snapshot-lightbox-footer">
        <p class="muted snapshot-lightbox-time"></p>
        <a class="snapshot-download-btn" id="snapshot-download" download aria-label="Download image">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <path d="M12 3v12"/>
            <path d="m7 10 5 5 5-5"/>
            <path d="M5 21h14"/>
          </svg>
        </a>
      </div>
    </div>
  `;

  const timeEl = overlay.querySelector<HTMLElement>(".snapshot-lightbox-time");
  if (timeEl) {
    try {
      timeEl.textContent = new Date(thumb.capturedAt).toLocaleString();
    } catch {
      timeEl.textContent = thumb.capturedAt;
    }
  }

  const download = overlay.querySelector<HTMLAnchorElement>("#snapshot-download");
  if (download) {
    download.href = thumb.url;
    download.download = `sleepguard-${thumb.id.slice(0, 8)}.jpg`;
  }

  const close = () => closeSnapshotLightbox();
  overlay.querySelector("[data-close]")?.addEventListener("click", close);
  overlay.querySelector(".snapshot-lightbox-close")?.addEventListener("click", close);

  const onKey = (e: KeyboardEvent) => {
    if (e.key === "Escape") {
      close();
    }
  };
  document.addEventListener("keydown", onKey);
  overlay.dataset.keyHandler = "1";
  (overlay as HTMLElement & { _onKey?: (e: KeyboardEvent) => void })._onKey = onKey;

  document.body.appendChild(overlay);
  lightboxEl = overlay;
  document.body.style.overflow = "hidden";
}

export function closeSnapshotLightbox(): void {
  if (!lightboxEl) {
    return;
  }
  const el = lightboxEl as HTMLElement & { _onKey?: (e: KeyboardEvent) => void };
  if (el._onKey) {
    document.removeEventListener("keydown", el._onKey);
  }
  lightboxEl.remove();
  lightboxEl = null;
  document.body.style.overflow = "";
}
