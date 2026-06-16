import type {
  CleanupResult,
  DeviceStatus,
  EventsResponse,
  PairedClient,
  PairingsResponse,
  Settings,
  Snapshot,
  SnapshotsResponse,
} from "./types";

function headers(readKey: string): HeadersInit {
  return {
    Authorization: `Bearer ${readKey}`,
    Accept: "application/json",
  };
}

function jsonHeaders(readKey: string): HeadersInit {
  return {
    ...headers(readKey),
    "Content-Type": "application/json",
  };
}

function base(settings: Settings): string {
  return settings.apiBase.replace(/\/$/, "");
}

export async function fetchEvents(
  settings: Settings,
  limit = 50,
): Promise<EventsResponse> {
  const url = new URL(`${base(settings)}/api/v1/events`);
  url.searchParams.set("device_id", settings.deviceId);
  url.searchParams.set("limit", String(limit));

  const res = await fetch(url, { headers: headers(settings.readKey) });
  if (!res.ok) {
    const body = await res.text();
    throw new Error(`events ${res.status}: ${body}`);
  }
  return res.json() as Promise<EventsResponse>;
}

export async function fetchStatus(settings: Settings): Promise<DeviceStatus> {
  const url = `${base(settings)}/api/v1/devices/${encodeURIComponent(settings.deviceId)}/status`;
  const res = await fetch(url, { headers: headers(settings.readKey) });
  if (!res.ok) {
    const body = await res.text();
    throw new Error(`status ${res.status}: ${body}`);
  }
  return res.json() as Promise<DeviceStatus>;
}

export async function fetchPairings(settings: Settings): Promise<PairedClient[]> {
  const url = new URL(`${base(settings)}/api/v1/pair`);
  url.searchParams.set("device_id", settings.deviceId);
  const res = await fetch(url, { headers: headers(settings.readKey) });
  if (!res.ok) {
    const body = await res.text();
    throw new Error(`pairings ${res.status}: ${body}`);
  }
  const data = (await res.json()) as PairingsResponse;
  return data.clients ?? [];
}

export async function pairDevice(
  settings: Settings,
  subscription: PushSubscription,
  name = "This phone",
): Promise<PairedClient> {
  const res = await fetch(`${base(settings)}/api/v1/pair`, {
    method: "POST",
    headers: jsonHeaders(settings.readKey),
    body: JSON.stringify({
      device_id: settings.deviceId,
      name,
      subscription: subscription.toJSON(),
    }),
  });
  if (!res.ok) {
    const body = await res.text();
    throw new Error(`pair ${res.status}: ${body}`);
  }
  const data = (await res.json()) as { client: PairedClient };
  return data.client;
}

export async function unpairDevice(settings: Settings, pairingId: string): Promise<void> {
  const res = await fetch(
    `${base(settings)}/api/v1/pair/${encodeURIComponent(pairingId)}`,
    {
      method: "DELETE",
      headers: headers(settings.readKey),
    },
  );
  if (!res.ok) {
    const body = await res.text();
    throw new Error(`unpair ${res.status}: ${body}`);
  }
}

export async function runCleanup(settings: Settings): Promise<CleanupResult> {
  const res = await fetch(`${base(settings)}/api/v1/admin/cleanup`, {
    method: "POST",
    headers: headers(settings.readKey),
  });
  if (!res.ok) {
    const body = await res.text();
    throw new Error(`cleanup ${res.status}: ${body}`);
  }
  return res.json() as Promise<CleanupResult>;
}

export async function requestCapture(settings: Settings): Promise<void> {
  const url = `${base(settings)}/api/v1/devices/${encodeURIComponent(settings.deviceId)}/capture`;
  const res = await fetch(url, {
    method: "POST",
    headers: headers(settings.readKey),
  });
  if (!res.ok) {
    const body = await res.text();
    throw new Error(`capture ${res.status}: ${body}`);
  }
}

export async function fetchSnapshots(
  settings: Settings,
  limit = 5,
): Promise<Snapshot[]> {
  const url = new URL(`${base(settings)}/api/v1/snapshots`);
  url.searchParams.set("device_id", settings.deviceId);
  url.searchParams.set("limit", String(limit));
  const res = await fetch(url, { headers: headers(settings.readKey) });
  if (!res.ok) {
    const body = await res.text();
    throw new Error(`snapshots ${res.status}: ${body}`);
  }
  const data = (await res.json()) as SnapshotsResponse;
  return data.snapshots ?? [];
}

export async function fetchSnapshotBlob(
  settings: Settings,
  snapshotId: string,
): Promise<Blob> {
  const url = `${base(settings)}/api/v1/snapshots/${encodeURIComponent(snapshotId)}/image`;
  const res = await fetch(url, { headers: headers(settings.readKey) });
  if (!res.ok) {
    const body = await res.text();
    throw new Error(`snapshot image ${res.status}: ${body}`);
  }
  return res.blob();
}
