export type OverlaySlot =
  | "top-left"
  | "top-center"
  | "top-right"
  | "middle-left"
  | "center"
  | "middle-right"
  | "bottom-left"
  | "bottom-center"
  | "bottom-right";

export type OverlayVisibility = "always" | "on-photo-change" | "when-degraded";

export interface OverlayEnvelope {
  enabled: boolean;
  slot: OverlaySlot;
  visibility: OverlayVisibility;
}

export interface OverlayConfig {
  clock: OverlayEnvelope;
  photoInfo: OverlayEnvelope;
  status: OverlayEnvelope;
}

export interface SetupPublicState {
  configured: boolean;
  status: string;
  setupCodeRequired: boolean;
  setupCode?: string;
  setupUrl: string;
  hostname: string;
  ipAddress?: string;
  adminPasswordExists: boolean;
}

export interface FrameAsset {
  id: string;
  mediaUrl: string;
  title: string;
  sourceName: string;
  takenAt?: string;
}

export interface FrameState {
  configured: boolean;
  paused: boolean;
  status: "ready" | "empty" | "degraded" | "error" | string;
  message?: string;
  current?: FrameAsset;
  next?: FrameAsset;
  previous?: FrameAsset;
  updatedAt: string;
  overlays: OverlayConfig;
  setup: SetupPublicState;
}

export interface AppConfig {
  device: { name: string; timezone: string };
  server: { host: string; port: number; hostname: string };
  immich: { url: string };
  source: {
    mode: "local_folder" | "album" | "random";
    album: { id: string; shuffle: boolean };
    random: { shuffle: boolean };
    localFolder: { path: string; shuffle: boolean };
  };
  filters: {
    photosOnly: boolean;
    excludeArchived: boolean;
    excludeHidden: boolean;
    excludeTrashed: boolean;
    excludeVideos: boolean;
  };
  display: {
    orientation: string;
    width: number;
    height: number;
    fit: "contain" | "cover";
    background: string;
    transition: string;
    transitionMs: number;
  };
  slideshow: { intervalSeconds: number; recentHistoryLimit: number };
  cache: {
    preset: "extra-small" | "light" | "balanced" | "large" | "custom";
    maxSizeMb: number;
    minFreeMb: number;
    targetItems: number;
    prefetchItems: number;
    refreshBatchItems: number;
    refreshAfterShownItems: number;
    rendition: string;
  };
  sync: { refreshIntervalMinutes: number };
  overlays: OverlayConfig;
  weather: {
    enabled: boolean;
    provider: string;
    location: string;
    units: string;
    refreshMinutes: number;
  };
}

export interface SettingsResponse {
  config: AppConfig;
  hasImmichApiKey: boolean;
  status: PortalStatus;
}

export interface PortalStatus {
  setup: SetupPublicState;
  configured: boolean;
  hasImmichApiKey: boolean;
  immich: ImmichStatus;
  sourceMode: string;
  cacheCount: number;
  lastError?: string;
}

export interface ImmichStatus {
  url?: string;
  configured: boolean;
  validated: boolean;
  validationRequired: boolean;
  validatedAt?: string;
  version?: string;
  keyName?: string;
}

export interface ImmichAlbum {
  ID?: string;
  Name?: string;
  AssetCount?: number;
  id?: string;
  name?: string;
  assetCount?: number;
}

export interface ImmichTestResponse {
  ok: boolean;
  version: string;
  keyName: string;
  status: PortalStatus;
}

export interface DisplayReport {
  width: number;
  height: number;
  devicePixelRatio: number;
  screenWidth: number;
  screenHeight: number;
}

export async function fetchFrameState(): Promise<FrameState> {
  const response = await fetch("/api/state");
  if (!response.ok) {
    throw new Error(`State request failed: ${response.status}`);
  }
  return response.json();
}

export async function fetchSetupState(): Promise<SetupPublicState> {
  const response = await fetch("/api/setup/state");
  if (!response.ok) {
    throw new Error(`Setup state request failed: ${response.status}`);
  }
  return response.json();
}

export async function claimSetupCode(code: string): Promise<SetupPublicState> {
  return postJSON("/api/setup/claim", { code });
}

export async function createAdminPassword(password: string): Promise<void> {
  await postJSON("/api/setup/admin-password", { password });
}

export async function login(password: string): Promise<void> {
  await postJSON("/api/auth/login", { password });
}

export async function logout(): Promise<void> {
  await postJSON("/api/auth/logout", {});
}

export async function fetchAuthSession(): Promise<{ admin: boolean; setup: boolean }> {
  const response = await fetch("/api/auth/session");
  if (!response.ok) {
    throw new Error(`Session request failed: ${response.status}`);
  }
  return response.json();
}

export async function fetchSettings(): Promise<SettingsResponse> {
  const response = await fetch("/api/settings");
  if (!response.ok) {
    throw new Error(`Settings request failed: ${response.status}`);
  }
  return response.json();
}

export async function fetchStatus(): Promise<PortalStatus> {
  const response = await fetch("/api/status");
  if (!response.ok) {
    throw new Error(`Status request failed: ${response.status}`);
  }
  return response.json();
}

export async function saveSettings(config: AppConfig, immichApiKey?: string): Promise<SettingsResponse> {
  const body: { config: AppConfig; immichApiKey?: string } = { config };
  if (immichApiKey !== undefined) body.immichApiKey = immichApiKey;
  return putJSON("/api/settings", body);
}

export async function testImmich(url: string, apiKey: string): Promise<ImmichTestResponse> {
  return postJSON("/api/immich/test", { url, apiKey });
}

export async function fetchImmichAlbums(): Promise<ImmichAlbum[]> {
  const response = await fetch("/api/immich/albums");
  if (!response.ok) {
    throw new Error(await response.text());
  }
  const data = await response.json();
  return data.albums ?? [];
}

export async function completeSetup(): Promise<SetupPublicState> {
  return postJSON("/api/setup/complete", {});
}

export function subscribeFrameState(onState: (state: FrameState) => void): EventSource {
  const source = new EventSource("/api/events");
  source.addEventListener("state", (event) => {
    onState(JSON.parse((event as MessageEvent).data));
  });
  return source;
}

export async function postPlayback(command: "next" | "previous" | "pause" | "resume"): Promise<FrameState> {
  const response = await fetch(`/api/playback/${command}`, { method: "POST" });
  if (!response.ok) {
    throw new Error(`Playback command failed: ${response.status}`);
  }
  return response.json();
}

export function reportDisplaySize(report: DisplayReport): void {
  void fetch("/api/display", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(report)
  });
}

async function postJSON<T>(url: string, body: unknown): Promise<T> {
  return sendJSON("POST", url, body);
}

async function putJSON<T>(url: string, body: unknown): Promise<T> {
  return sendJSON("PUT", url, body);
}

async function sendJSON<T>(method: "POST" | "PUT", url: string, body: unknown): Promise<T> {
  const response = await fetch(url, {
    method,
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body)
  });
  if (!response.ok) {
    throw new Error(await response.text());
  }
  return response.json();
}
