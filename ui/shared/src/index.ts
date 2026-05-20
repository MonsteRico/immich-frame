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
  status: "ready" | "empty" | "degraded" | string;
  message?: string;
  current?: FrameAsset;
  next?: FrameAsset;
  previous?: FrameAsset;
  updatedAt: string;
  overlays: OverlayConfig;
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
