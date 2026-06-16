import type { DeviceStatus, EventsResponse, Settings } from "./types";

function headers(readKey: string): HeadersInit {
  return {
    Authorization: `Bearer ${readKey}`,
    Accept: "application/json",
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
