import { render, type ComponentChildren } from "preact";
import { useEffect, useMemo, useState } from "preact/hooks";
import {
  claimSetupCode,
  completeSetup,
  createAdminPassword,
  fetchAuthSession,
  fetchImmichAlbums,
  fetchSettings,
  fetchSetupState,
  fetchStatus,
  login,
  saveSettings,
  testImmich,
  type AppConfig,
  type ImmichAlbum,
  type PortalStatus,
  type SetupPublicState
} from "@immich-frame/shared";
import "./styles.css";

type Step = "loading" | "claim" | "login" | "password" | "immich" | "source" | "settings" | "done";

function App() {
  const [step, setStep] = useState<Step>("loading");
  const [setup, setSetup] = useState<SetupPublicState | null>(null);
  const [settings, setSettings] = useState<AppConfig | null>(null);
  const [hasKey, setHasKey] = useState(false);
  const [status, setStatus] = useState<PortalStatus | null>(null);
  const [message, setMessage] = useState("");

  useEffect(() => {
    void bootstrap();
  }, []);

  const bootstrap = async () => {
    try {
      const setupState = await fetchSetupState();
      setSetup(setupState);
      const session = await fetchAuthSession();
      if (!setupState.configured && !session.setup && !session.admin) {
        setStep("claim");
        return;
      }
      if (setupState.configured && !session.admin) {
        setStep("login");
        return;
      }
      const loaded = await fetchSettings();
      setSettings(loaded.config);
      setHasKey(loaded.hasImmichApiKey);
      setStatus(loaded.status);
      if (!setupState.adminPasswordExists) setStep("password");
      else if (!setupState.configured) setStep("immich");
      else setStep("settings");
    } catch (err) {
      setMessage(errorText(err));
      setStep("claim");
    }
  };

  const refreshSettings = async () => {
    const loaded = await fetchSettings();
    setSettings(loaded.config);
    setHasKey(loaded.hasImmichApiKey);
    setStatus(loaded.status);
    return loaded.config;
  };

  const refreshStatus = async () => {
    const loaded = await fetchStatus();
    setStatus(loaded);
    return loaded;
  };

  return (
    <main className="setup-shell">
      <section className="topbar">
        <div>
          <p className="eyebrow">Immich Frame</p>
          <h1>{titleFor(step, setup)}</h1>
        </div>
        <span className="status-pill">{setup?.configured ? "Configured" : "First boot"}</span>
      </section>
      {message ? <div className="notice">{message}</div> : null}
      {step === "loading" ? <Panel><p>Loading setup...</p></Panel> : null}
      {step === "claim" ? <ClaimPanel onClaim={async (code) => {
        setMessage("");
        const claimed = await claimSetupCode(code);
        setSetup(claimed);
        const loaded = await fetchSettings();
        setSettings(loaded.config);
        setHasKey(loaded.hasImmichApiKey);
        setStatus(loaded.status);
        setStep(claimed.adminPasswordExists ? "immich" : "password");
      }} onError={setMessage} /> : null}
      {step === "login" ? <LoginPanel onLogin={async (password) => {
        setMessage("");
        await login(password);
        await refreshSettings();
        setStep("settings");
      }} onError={setMessage} /> : null}
      {step === "password" ? <PasswordPanel onSave={async (password) => {
        setMessage("");
        await createAdminPassword(password);
        setSetup((current) => current ? { ...current, adminPasswordExists: true } : current);
        await refreshSettings();
        setStep("immich");
      }} onError={setMessage} /> : null}
      {step === "immich" && settings ? <ImmichPanel config={settings} hasKey={hasKey} status={status} onValidated={setStatus} onSave={async (next, key) => {
        setMessage("");
        const saved = await saveSettings(next, key);
        setSettings(saved.config);
        setHasKey(saved.hasImmichApiKey);
        setStatus(saved.status);
        setStep("source");
      }} onError={setMessage} /> : null}
      {step === "source" && settings ? <SourcePanel config={settings} status={status} onBack={() => setStep("immich")} onSave={async (next) => {
        setMessage("");
        const saved = await saveSettings(next);
        setSettings(saved.config);
        setStatus(saved.status);
        const finished = await completeSetup();
        setSetup(finished);
        await refreshStatus();
        setStep("done");
      }} onError={setMessage} /> : null}
      {step === "settings" && settings ? <SettingsPanel config={settings} hasKey={hasKey} status={status} onValidated={setStatus} onSave={async (next, key) => {
        setMessage("");
        const saved = await saveSettings(next, key);
        setSettings(saved.config);
        setHasKey(saved.hasImmichApiKey);
        setStatus(saved.status);
        setMessage("Settings saved.");
      }} onError={setMessage} /> : null}
      {step === "done" ? <Panel>
        <p className="success-mark">Ready</p>
        <h2>Setup complete</h2>
        <p className="muted">The setup code is no longer valid. The frame can now start the Immich slideshow as photos are cached.</p>
        <button type="button" onClick={() => setStep("settings")}>Open settings</button>
      </Panel> : null}
    </main>
  );
}

function ClaimPanel({ onClaim, onError }: { onClaim: (code: string) => Promise<void>; onError: (message: string) => void }) {
  const [code, setCode] = useState("");
  return <Panel>
    <p className="step">Step 1</p>
    <h2>Enter the code shown on the frame</h2>
    <p className="muted">This pairs your phone with the local frame for first boot setup.</p>
    <label>Setup code<input inputMode="numeric" autoComplete="one-time-code" value={code} onInput={(event) => setCode(event.currentTarget.value)} placeholder="123456" /></label>
    <button type="button" onClick={() => onClaim(code).catch((err) => onError(errorText(err)))}>Continue</button>
  </Panel>;
}

function LoginPanel({ onLogin, onError }: { onLogin: (password: string) => Promise<void>; onError: (message: string) => void }) {
  const [password, setPassword] = useState("");
  return <Panel>
    <h2>Admin sign in</h2>
    <p className="muted">Settings are protected after first setup.</p>
    <label>Password<input type="password" value={password} onInput={(event) => setPassword(event.currentTarget.value)} /></label>
    <button type="button" onClick={() => onLogin(password).catch((err) => onError(errorText(err)))}>Sign in</button>
  </Panel>;
}

function PasswordPanel({ onSave, onError }: { onSave: (password: string) => Promise<void>; onError: (message: string) => void }) {
  const [password, setPassword] = useState("");
  return <Panel>
    <p className="step">Step 2</p>
    <h2>Create the local admin password</h2>
    <p className="muted">Use at least 8 characters. This password protects setup, settings, and LAN media access.</p>
    <label>Admin password<input type="password" value={password} onInput={(event) => setPassword(event.currentTarget.value)} /></label>
    <button type="button" onClick={() => onSave(password).catch((err) => onError(errorText(err)))}>Save password</button>
  </Panel>;
}

function ImmichPanel({ config, hasKey, status, onValidated, onSave, onError }: {
  config: AppConfig;
  hasKey: boolean;
  status: PortalStatus | null;
  onValidated: (status: PortalStatus) => void;
  onSave: (config: AppConfig, key?: string) => Promise<void>;
  onError: (message: string) => void;
}) {
  const [url, setUrl] = useState(config.immich.url);
  const [apiKey, setApiKey] = useState("");
  const [result, setResult] = useState("");
  const [validatedFor, setValidatedFor] = useState(status?.immich.validated ? validationKey(config.immich.url, "", hasKey) : "");
  const next = useMemo(() => ({ ...config, immich: { ...config.immich, url } }), [config, url]);
  const isHTTP = url.trim().startsWith("http://");
  const currentValidationKey = validationKey(url, apiKey, hasKey);
  const missingURL = url.trim() === "";
  const missingKey = !hasKey && apiKey.trim() === "";
  const validated = validatedFor === currentValidationKey;
  const canTest = !missingURL && !missingKey;
  const canSave = canTest && validated;
  const resetValidation = () => {
    setResult("");
    setValidatedFor("");
  };
  return <Panel>
    <p className="step">Step 3</p>
    <h2>Connect Immich</h2>
    <label>Immich URL<input value={url} onInput={(event) => { setUrl(event.currentTarget.value); resetValidation(); }} placeholder="https://immich.example.com" /></label>
    {isHTTP ? <p className="warning">HTTP sends the Immich API key over the local network without encryption. Use it only for trusted homelab URLs.</p> : null}
    <label>{hasKey ? "Replace API key" : "Dedicated API key"}<input type="password" value={apiKey} onInput={(event) => { setApiKey(event.currentTarget.value); resetValidation(); }} placeholder={hasKey ? "Leave blank to keep saved key" : "Paste API key"} /></label>
    {missingURL ? <p className="warning">Enter the Immich URL before testing the connection.</p> : null}
    {missingKey ? <p className="warning">Paste a dedicated Immich API key before testing the connection.</p> : null}
    {!validated && canTest ? <p className="warning">Test this URL and API key successfully before saving and finishing setup.</p> : null}
    {result ? <p className="test-result">{result}</p> : null}
    <div className="button-row">
      <button type="button" className="secondary" disabled={!canTest} onClick={() => testImmich(url, apiKey).then((info) => {
        setValidatedFor(currentValidationKey);
        setResult(`Connected to Immich ${info.version}${info.keyName ? ` as ${info.keyName}` : ""}.`);
        onValidated(info.status);
      }).catch((err) => onError(errorText(err)))}>Test</button>
      <button type="button" disabled={!canSave} onClick={() => canSave ? onSave(next, apiKey || undefined).catch((err) => onError(errorText(err))) : onError("Test the Immich connection successfully before saving.")}>Save</button>
    </div>
  </Panel>;
}

function SourcePanel({ config, status, onBack, onSave, onError }: {
  config: AppConfig;
  status: PortalStatus | null;
  onBack: () => void;
  onSave: (config: AppConfig) => Promise<void>;
  onError: (message: string) => void;
}) {
  const [mode, setMode] = useState<"album" | "random">(config.source.mode === "album" ? "album" : "random");
  const [albumID, setAlbumID] = useState(config.source.album.id);
  const [query, setQuery] = useState("");
  const [albums, setAlbums] = useState<ImmichAlbum[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    setLoading(true);
    fetchImmichAlbums().then(setAlbums).catch((err) => onError(errorText(err))).finally(() => setLoading(false));
  }, []);

  const filtered = albums.filter((album) => albumName(album).toLowerCase().includes(query.toLowerCase()));
  const next = { ...config, source: { ...config.source, mode, album: { ...config.source.album, id: albumID }, random: { shuffle: true } } };
  const validationReady = status?.immich.validated ?? false;
  const sourceReady = mode === "random" || albumID.trim() !== "";
  const canFinish = validationReady && sourceReady;
  return <Panel>
    <p className="step">Step 4</p>
    <h2>Choose what to show</h2>
    <div className="segmented">
      <button type="button" className={mode === "album" ? "active" : ""} onClick={() => setMode("album")}>Album</button>
      <button type="button" className={mode === "random" ? "active" : ""} onClick={() => setMode("random")}>Random</button>
    </div>
    {mode === "album" ? <>
      <label>Search albums<input value={query} onInput={(event) => setQuery(event.currentTarget.value)} placeholder="Family, trips, favorites" /></label>
      <div className="album-list" aria-busy={loading}>
        {filtered.map((album) => {
          const id = albumId(album);
          return <button type="button" key={id} className={albumID === id ? "album active" : "album"} onClick={() => setAlbumID(id)}>
            <span>{albumName(album)}</span>
            <small>{albumCount(album)} items</small>
          </button>;
        })}
      </div>
    </> : <p className="muted">Random mode pulls a changing set of photos from the Immich library using conservative photo-only filters.</p>}
    {!validationReady ? <p className="warning">Finish is locked until the saved Immich URL and API key pass a connection test.</p> : null}
    {mode === "album" && !albumID.trim() ? <p className="warning">Choose an album or switch to random library mode before finishing setup.</p> : null}
    <div className="button-row">
      <button type="button" className="secondary" onClick={onBack}>Back</button>
      <button type="button" disabled={!canFinish} onClick={() => canFinish ? onSave(next).catch((err) => onError(errorText(err))) : onError("Validate Immich and choose a source before finishing setup.")}>Finish setup</button>
    </div>
  </Panel>;
}

function SettingsPanel({ config, hasKey, status, onValidated, onSave, onError }: {
  config: AppConfig;
  hasKey: boolean;
  status: PortalStatus | null;
  onValidated: (status: PortalStatus) => void;
  onSave: (config: AppConfig, key?: string) => Promise<void>;
  onError: (message: string) => void;
}) {
  const [draft, setDraft] = useState(config);
  const [apiKey, setApiKey] = useState("");
  const [testResult, setTestResult] = useState("");
  useEffect(() => setDraft(config), [config]);
  const canTest = draft.immich.url.trim() !== "" && (hasKey || apiKey.trim() !== "");
  return <Panel>
    <h2>Settings</h2>
    {status ? <StatusSummary status={status} /> : null}
    <label>Frame name<input value={draft.device.name} onInput={(event) => setDraft({ ...draft, device: { ...draft.device, name: event.currentTarget.value } })} /></label>
    <label>Immich URL<input value={draft.immich.url} onInput={(event) => setDraft({ ...draft, immich: { url: event.currentTarget.value } })} /></label>
    {draft.immich.url.startsWith("http://") ? <p className="warning">HTTP Immich URLs should only be used on trusted local networks.</p> : null}
    <label>{hasKey ? "Replace Immich API key" : "Immich API key"}<input type="password" value={apiKey} onInput={(event) => setApiKey(event.currentTarget.value)} placeholder={hasKey ? "Saved key is hidden" : "Paste API key"} /></label>
    {status?.immich.validationRequired ? <p className="warning">Immich needs a successful connection test before first setup can finish. Saving a new URL or key clears the previous validation.</p> : null}
    {testResult ? <p className="test-result">{testResult}</p> : null}
    <div className="settings-grid">
      <label>Slide seconds<input type="number" min="5" value={draft.slideshow.intervalSeconds} onInput={(event) => setDraft({ ...draft, slideshow: { ...draft.slideshow, intervalSeconds: Number(event.currentTarget.value) } })} /></label>
      <label>Display fit<select value={draft.display.fit} onInput={(event) => setDraft({ ...draft, display: { ...draft.display, fit: event.currentTarget.value as "contain" | "cover" } })}><option value="contain">Contain</option><option value="cover">Cover</option></select></label>
      <label>Cache preset<select value={draft.cache.preset} onInput={(event) => setDraft({ ...draft, cache: presetCache(draft.cache, event.currentTarget.value) })}><option value="extra-small">Extra small</option><option value="light">Light</option><option value="balanced">Balanced</option><option value="large">Large</option></select></label>
    </div>
    <div className="toggle-row">
      <label><input type="checkbox" checked={draft.overlays.clock.enabled} onInput={(event) => setDraft({ ...draft, overlays: { ...draft.overlays, clock: { ...draft.overlays.clock, enabled: event.currentTarget.checked } } })} /> Clock</label>
      <label><input type="checkbox" checked={draft.overlays.photoInfo.enabled} onInput={(event) => setDraft({ ...draft, overlays: { ...draft.overlays, photoInfo: { ...draft.overlays.photoInfo, enabled: event.currentTarget.checked } } })} /> Photo info</label>
      <label><input type="checkbox" checked={draft.overlays.status.enabled} onInput={(event) => setDraft({ ...draft, overlays: { ...draft.overlays, status: { ...draft.overlays.status, enabled: event.currentTarget.checked } } })} /> Status</label>
    </div>
    <div className="button-row">
      <button type="button" className="secondary" disabled={!canTest} onClick={() => testImmich(draft.immich.url, apiKey).then((info) => {
        setTestResult(`Connected to Immich ${info.version}${info.keyName ? ` as ${info.keyName}` : ""}.`);
        onValidated(info.status);
      }).catch((err) => onError(errorText(err)))}>Test Immich</button>
      <button type="button" onClick={() => onSave(draft, apiKey || undefined).catch((err) => onError(errorText(err)))}>Save settings</button>
    </div>
  </Panel>;
}

function StatusSummary({ status }: { status: PortalStatus }) {
  const validatedAt = status.immich.validatedAt ? new Date(status.immich.validatedAt).toLocaleString() : "";
  return <div className="status-grid" aria-label="Frame status">
    <div><span>Setup</span><strong>{status.configured ? "Complete" : "In progress"}</strong></div>
    <div><span>Immich</span><strong>{status.immich.validated ? "Validated" : "Needs test"}</strong></div>
    <div><span>Source</span><strong>{status.sourceMode || "Not chosen"}</strong></div>
    <div><span>Cache</span><strong>{status.cacheCount} photos</strong></div>
    {status.immich.version ? <div><span>Version</span><strong>{status.immich.version}</strong></div> : null}
    {validatedAt ? <div><span>Validated</span><strong>{validatedAt}</strong></div> : null}
    {status.lastError ? <div className="status-wide"><span>Last error</span><strong>{status.lastError}</strong></div> : null}
  </div>;
}

function Panel({ children }: { children: ComponentChildren }) {
  return <section className="setup-panel">{children}</section>;
}

function titleFor(step: Step, setup: SetupPublicState | null) {
  if (step === "settings") return "Settings";
  if (step === "done") return "Ready to play";
  if (setup?.configured) return "Welcome back";
  return "Set up your frame";
}

function albumId(album: ImmichAlbum) {
  return album.id ?? album.ID ?? "";
}

function albumName(album: ImmichAlbum) {
  return album.name ?? album.Name ?? "Untitled album";
}

function albumCount(album: ImmichAlbum) {
  return album.assetCount ?? album.AssetCount ?? 0;
}

function presetCache(cache: AppConfig["cache"], preset: string): AppConfig["cache"] {
  if (preset === "extra-small") return { ...cache, preset: "extra-small", maxSizeMb: 128, targetItems: 10, prefetchItems: 3, refreshBatchItems: 5, refreshAfterShownItems: 5 };
  if (preset === "light") return { ...cache, preset: "light", maxSizeMb: 512, targetItems: 150, prefetchItems: 10, refreshBatchItems: 75, refreshAfterShownItems: 75 };
  if (preset === "large") return { ...cache, preset: "large", maxSizeMb: 4096, targetItems: 1000, prefetchItems: 40, refreshBatchItems: 500, refreshAfterShownItems: 500 };
  return { ...cache, preset: "balanced", maxSizeMb: 2048, targetItems: 500, prefetchItems: 20, refreshBatchItems: 250, refreshAfterShownItems: 250 };
}

function validationKey(url: string, apiKey: string, hasSavedKey: boolean) {
  return `${url.trim().replace(/\/+$/, "")}|${apiKey.trim() || (hasSavedKey ? "saved" : "")}`;
}

function errorText(err: unknown) {
  return err instanceof Error ? err.message.trim() : "Something went wrong.";
}

render(<App />, document.getElementById("app")!);
