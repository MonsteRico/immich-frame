import { render } from "preact";
import { useEffect, useMemo, useState } from "preact/hooks";
import {
  fetchFrameState,
  postPlayback,
  reportDisplaySize,
  subscribeFrameState,
  type FrameAsset,
  type FrameState
} from "@immich-frame/shared";
import "./styles.css";

function App() {
  const [state, setState] = useState<FrameState | null>(null);
  const [error, setError] = useState("");
  const [controlsVisible, setControlsVisible] = useState(false);

  useEffect(() => {
    fetchFrameState().then(setState).catch((err: Error) => setError(err.message));
    const events = subscribeFrameState(setState);
    reportDisplaySize({
      width: window.innerWidth,
      height: window.innerHeight,
      devicePixelRatio: window.devicePixelRatio,
      screenWidth: window.screen.width,
      screenHeight: window.screen.height
    });
    return () => events.close();
  }, []);

  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      if (event.key === "ArrowRight") void handleCommand("next");
      if (event.key === "ArrowLeft") void handleCommand("previous");
      if (event.key === " ") void handleCommand(state?.paused ? "resume" : "pause");
      if (event.key === "i") setControlsVisible((visible) => !visible);
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [state?.paused]);

  const handleCommand = async (command: "next" | "previous" | "pause" | "resume") => {
    try {
      setState(await postPlayback(command));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Playback command failed");
    }
  };

  const current = state?.current;
  const next = state?.next;
  const statusText = error || state?.message || "";

  return (
    <main
      className="frame"
      onMouseMove={() => setControlsVisible(true)}
      onMouseLeave={() => setControlsVisible(false)}
      onClick={() => setControlsVisible((visible) => !visible)}
    >
      <SlideshowImage asset={current} next={next} setup={state?.setup} />
      <OverlayLayer asset={current} status={statusText} />
      <Controls
        visible={controlsVisible}
        paused={state?.paused ?? false}
        onPrevious={() => void handleCommand("previous")}
        onNext={() => void handleCommand("next")}
        onTogglePause={() => void handleCommand(state?.paused ? "resume" : "pause")}
      />
    </main>
  );
}

function SlideshowImage({ asset, next, setup }: { asset?: FrameAsset; next?: FrameAsset; setup?: FrameState["setup"] }) {
  const background = asset?.mediaUrl ?? next?.mediaUrl;
  if (!asset) {
    if (setup?.setupCodeRequired) {
      return <SetupScreen setup={setup} />;
    }
    return (
      <section className="empty-state">
        <h1>Immich Frame</h1>
        <p>Add a few images to dev/photos, then run the local daemon to preview the slideshow.</p>
      </section>
    );
  }
  return (
    <section className="stage" aria-label={asset.title || "Current photo"}>
      <img className="backdrop" src={background} alt="" aria-hidden="true" />
      <img className="photo photo-current" key={asset.id} src={asset.mediaUrl} alt="" />
    </section>
  );
}

function SetupScreen({ setup }: { setup: FrameState["setup"] }) {
  const ipURL = setup.ipAddress ? `http://${setup.ipAddress}:8787/setup` : "";
  return (
    <section className="setup-state">
      <div className="setup-copy">
        <p>First boot</p>
        <h1>Set up Immich Frame</h1>
        <div className="setup-routes">
          <span>{setup.setupUrl}</span>
          {ipURL ? <span>{ipURL}</span> : null}
        </div>
      </div>
      <div className="setup-code" aria-label="Setup code">
        <span>Setup code</span>
        <strong>{setup.setupCode || "------"}</strong>
      </div>
    </section>
  );
}

function OverlayLayer({ asset, status }: { asset?: FrameAsset; status: string }) {
  const clock = useClock();
  const takenAt = useMemo(() => {
    if (!asset?.takenAt) return "";
    return new Intl.DateTimeFormat([], { month: "short", day: "numeric", year: "numeric" }).format(new Date(asset.takenAt));
  }, [asset?.takenAt]);
  return (
    <div className="overlays" aria-live="polite">
      <div className="overlay clock">{clock}</div>
      {asset ? (
        <div className="overlay photo-info">
          <strong>{asset.title}</strong>
          <span>{[takenAt, asset.sourceName].filter(Boolean).join(" - ")}</span>
        </div>
      ) : null}
      {status ? <div className="overlay status">{status}</div> : null}
    </div>
  );
}

function Controls({
  visible,
  paused,
  onPrevious,
  onNext,
  onTogglePause
}: {
  visible: boolean;
  paused: boolean;
  onPrevious: () => void;
  onNext: () => void;
  onTogglePause: () => void;
}) {
  return (
    <div className={`controls ${visible ? "is-visible" : ""}`}>
      <button aria-label="Previous photo" onClick={onPrevious} type="button">
        <ChevronLeft />
      </button>
      <button aria-label={paused ? "Resume slideshow" : "Pause slideshow"} onClick={onTogglePause} type="button">
        {paused ? <PlayIcon /> : <PauseIcon />}
      </button>
      <button aria-label="Next photo" onClick={onNext} type="button">
        <ChevronRight />
      </button>
    </div>
  );
}

function useClock() {
  const [now, setNow] = useState(new Date());
  useEffect(() => {
    const timer = window.setInterval(() => setNow(new Date()), 30000);
    return () => window.clearInterval(timer);
  }, []);
  return new Intl.DateTimeFormat([], { hour: "numeric", minute: "2-digit" }).format(now);
}

function ChevronLeft() {
  return <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M15 5 8 12l7 7" /></svg>;
}

function ChevronRight() {
  return <svg viewBox="0 0 24 24" aria-hidden="true"><path d="m9 5 7 7-7 7" /></svg>;
}

function PauseIcon() {
  return <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M8 5v14M16 5v14" /></svg>;
}

function PlayIcon() {
  return <svg viewBox="0 0 24 24" aria-hidden="true"><path d="m8 5 11 7-11 7V5z" /></svg>;
}

render(<App />, document.getElementById("app")!);
