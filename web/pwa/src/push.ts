const PAIRING_KEY = "sleepguard.pairingId";

export function loadPairingId(): string | null {
  return localStorage.getItem(PAIRING_KEY);
}

export function savePairingId(id: string): void {
  localStorage.setItem(PAIRING_KEY, id);
}

export function clearPairingId(): void {
  localStorage.removeItem(PAIRING_KEY);
}

function urlBase64ToUint8Array(base64: string): Uint8Array {
  const padding = "=".repeat((4 - (base64.length % 4)) % 4);
  const raw = atob((base64 + padding).replace(/-/g, "+").replace(/_/g, "/"));
  const out = new Uint8Array(raw.length);
  for (let i = 0; i < raw.length; i++) {
    out[i] = raw.charCodeAt(i);
  }
  return out;
}

export async function registerServiceWorker(): Promise<ServiceWorkerRegistration | null> {
  if (!("serviceWorker" in navigator)) {
    return null;
  }
  try {
    return await navigator.serviceWorker.register("/sw.js");
  } catch {
    return null;
  }
}

export async function subscribeForPush(
  apiBase: string,
  readKey: string,
): Promise<PushSubscription> {
  const base = apiBase.replace(/\/$/, "");
  const res = await fetch(`${base}/api/v1/push/vapid-key`);
  if (!res.ok) {
    throw new Error("push not configured on server");
  }
  const { public_key: publicKey } = (await res.json()) as { public_key: string };
  if (!publicKey) {
    throw new Error("missing VAPID public key");
  }

  const reg = await registerServiceWorker();
  if (!reg) {
    throw new Error("service worker not supported");
  }

  let sub = await reg.pushManager.getSubscription();
  if (!sub) {
    sub = await reg.pushManager.subscribe({
      userVisibleOnly: true,
      applicationServerKey: urlBase64ToUint8Array(publicKey),
    });
  }
  if (!sub) {
    throw new Error("could not create push subscription");
  }
  return sub;
}

export function pushSupported(): boolean {
  return "serviceWorker" in navigator && "PushManager" in window && "Notification" in window;
}
