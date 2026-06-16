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
