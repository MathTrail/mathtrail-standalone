import { App, PostMessageTransport } from "@modelcontextprotocol/ext-apps";

const drawingEl = document.getElementById("drawing")!;
const contextEl = document.getElementById("context")!;
const logEl = document.getElementById("log")!;
const pokeBtn = document.getElementById("btn-poke")!;
const messageBtn = document.getElementById("btn-message")!;
const updateContextBtn = document.getElementById("btn-update-context")!;

// 30 columns wide, per T03 ("текстовый рисунок 30-40 символов моноширинным шрифтом").
// Kept identical to spikeDrawing in main.go (server tool output) — every row is
// exactly 30 runes; a live Claude run caught a first version where content rows
// were one column wider than the border.
const DRAWING = [
  "┌────────────────────────────┐",
  "│ 2 + 2 = ?                  │",
  "│ A) 3   B) 4   C) 5   D) 22 │",
  "└────────────────────────────┘",
].join("\n");
drawingEl.textContent = DRAWING;

function log(line: string): void {
  logEl.textContent = `${new Date().toISOString()}  ${line}\n${logEl.textContent}`;
}

const app = new App(
  { name: "mathtrail-spike-widget", version: "0.1.0" },
  {}, // no app-registered tools in this spike
);

function renderContext(): void {
  const ctx = app.getHostContext();
  contextEl.textContent = JSON.stringify(
    {
      locale: ctx?.locale,
      platform: ctx?.platform,
      theme: ctx?.theme,
      displayMode: ctx?.displayMode,
      containerDimensions: ctx?.containerDimensions,
      safeAreaInsets: ctx?.safeAreaInsets,
    },
    null,
    2,
  );
}

app.onhostcontextchanged = (ctx) => {
  renderContext();
  log(`hostcontextchanged: ${JSON.stringify(ctx)}`);
};

pokeBtn.addEventListener("click", () => {
  void (async () => {
    try {
      const result = await app.callServerTool({
        name: "widget_poke",
        arguments: { note: "clicked from widget" },
      });
      log(`widget_poke -> ${JSON.stringify(result.structuredContent ?? result.content)}`);
    } catch (err) {
      log(`widget_poke failed: ${String(err)}`);
    }
  })();
});

messageBtn.addEventListener("click", () => {
  void (async () => {
    try {
      const result = await app.sendMessage({
        role: "user",
        content: [{ type: "text", text: "Hello from the MathTrail spike widget." }],
      });
      log(`ui/message -> ${JSON.stringify(result)}`);
    } catch (err) {
      log(`ui/message failed: ${String(err)}`);
    }
  })();
});

updateContextBtn.addEventListener("click", () => {
  void (async () => {
    try {
      const result = await app.updateModelContext({
        content: [
          {
            type: "text",
            text: "The MathTrail spike widget's counter was clicked. Treat this as background context, not a new turn.",
          },
        ],
      });
      log(`ui/update-model-context -> ${JSON.stringify(result)}`);
    } catch (err) {
      log(`ui/update-model-context failed: ${String(err)}`);
    }
  })();
});

async function main(): Promise<void> {
  try {
    await app.connect(new PostMessageTransport(window.parent, window.parent));
    renderContext();
    log(`connected: host=${JSON.stringify(app.getHostVersion())}`);
  } catch (err) {
    contextEl.textContent = `connect() failed: ${String(err)}`;
    log(`connect failed: ${String(err)}`);
  }
}

void main();
