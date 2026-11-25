type Payload = Record<string, unknown>;

export function moengageTrackEvent(event: string, payload?: Payload) {
  if (typeof window === "undefined") return;
  const moengage = (window as any).Moengage;
  if (moengage?.track_event) {
    moengage.track_event(event, payload ?? {});
  }
}
