export interface Settings {
  apiBase: string;
  deviceId: string;
  readKey: string;
}

export interface MotionEvent {
  id: number;
  device_id: string;
  timestamp: string;
  type: string;
  source: string;
  state: string;
  pattern?: string;
  received_at?: string;
}

export interface EventsResponse {
  device_id: string;
  count: number;
  events: MotionEvent[];
}

export interface DeviceStatus {
  id: string;
  name: string;
  last_seen_at?: string;
  event_count: number;
  online: boolean;
}

export interface PairedClient {
  id: string;
  device_id: string;
  name: string;
  notify_on_rise: boolean;
  created_at: string;
}

export interface PairingsResponse {
  device_id: string;
  count: number;
  clients: PairedClient[];
}

export interface CleanupResult {
  events_deleted: number;
  snapshots_deleted: number;
  cutoff: string;
}

export interface Snapshot {
  id: string;
  device_id: string;
  captured_at: string;
  size_bytes: number;
}

export interface SnapshotsResponse {
  device_id: string;
  count: number;
  snapshots: Snapshot[];
}
