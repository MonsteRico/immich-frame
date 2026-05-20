import { render } from "preact";
import { useState } from "preact/hooks";
import "./styles.css";

function App() {
  const [immichUrl, setImmichUrl] = useState("");
  const [apiKey, setApiKey] = useState("");

  return (
    <main className="setup-shell">
      <section className="setup-panel">
        <p className="device-name">Immich Frame</p>
        <h1>Connect your frame</h1>
        <p className="intro">
          This setup scaffold keeps the portal separate from the kiosk slideshow. Full setup validation and auth arrive in a later phase.
        </p>
        <label>
          Immich URL
          <input value={immichUrl} onInput={(event) => setImmichUrl(event.currentTarget.value)} placeholder="https://immich.example.com" />
        </label>
        <label>
          Dedicated API key
          <input value={apiKey} onInput={(event) => setApiKey(event.currentTarget.value)} placeholder="Paste API key" type="password" />
        </label>
        <button type="button">Save setup</button>
      </section>
    </main>
  );
}

render(<App />, document.getElementById("app")!);
