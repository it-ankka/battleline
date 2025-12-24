const messageLog = document.getElementById("message-log");
const chatLog = document.getElementById("chat-log");
const gameStateLog = document.getElementById("game-state-log");

const copyGameIdForm = document.getElementById("copy-game-id-form");
const createGameForm = document.getElementById("create-game-form");
const joinGameForm = document.getElementById("join-game-form");
const chatForm = document.getElementById("chat-form");
const readyForm = document.getElementById("ready-form");

const copyGameIdInput = document.getElementById("game-id-input");
const joinGameInput = document.getElementById("join-game-id-input");
const chatInput = document.getElementById("chat-input");

let conn;
let isReady = false;

window.addEventListener("unload", () => {
  if (conn.readyState == WebSocket.OPEN) conn.close();
});

// These are for debugging
window.placeCard = (suit, value, lane) => {
  let m = {
    type: "move",
    data: {
      move: {
        action: "placement",
        lane: lane,
        card: { suit: suit, value: value },
      },
    },
  };
  window.conn.send(JSON.stringify(m));
};

window.claim = (lane) => {
  let m = {
    type: "move",
    data: {
      move: {
        action: "claim",
        lane: lane,
      },
    },
  };
  window.conn.send(JSON.stringify(m));
};

window.drawCard = () => {
  let m = {
    type: "move",
    data: {
      move: {
        action: "draw",
        tacticsDeck: false,
      },
    },
  };
  window.conn.send(JSON.stringify(m));
};

function logMessage(msg) {
  messageLog.innerText += `[${new Date().toLocaleTimeString()}] ${msg}\n`;
}

function updateChatLog(chatLogMessages) {
  if (!chatLog) return;
  chatLog.innerText = chatLogMessages
    .map(
      (m) =>
        `[${new Date(m.timestamp).toLocaleTimeString()}] <${m.nickname}>: ${m.content}`,
    )
    .join("\n");
}

function updateUIForConnectedGame(gameId) {
  copyGameIdInput.value = gameId;
  joinGameForm.hidden = true;
  copyGameIdForm.hidden = false;
  chatForm.hidden = false;
  readyForm.hidden = false;
}

function connectToGame(gameId, maxRetries = 5) {
  conn = new WebSocket(`ws://${location.host}/ws/${gameId}`);
  window.conn = conn;

  conn.onclose = (ev) => {
    logMessage(`❌ Disconnected (code: ${ev.code}, reason: ${ev.reason})`);
    if (![1000, 1001, 1008].includes(ev.code) && maxRetries > 0) {
      logMessage(`Reconnecting in 1s (${maxRetries} retries left)`);
      setTimeout(() => connectToGame(gameId, maxRetries - 1), 1000);
      return;
    }
    joinGameForm.hidden = false;
    copyGameIdForm.hidden = true;
    chatForm.hidden = true;
    readyForm.hidden = true;

    const url = new URL(window.location.href);
    url.search = "";
    window.history.pushState(null, "", url.toString());
  };

  conn.onerror = (ev) => console.error("WebSocket error:", ev);

  conn.onopen = () => {
    logMessage("✅ Connected to WebSocket");
    updateUIForConnectedGame(gameId);
  };

  conn.onmessage = (ev) => {
    if (typeof ev.data !== "string") {
      console.error("Unexpected message type", typeof ev.data);
      return;
    }

    try {
      const data = JSON.parse(ev.data);
      handleServerMessage(data);
    } catch (err) {
      console.error("Failed to parse message:", err);
    }
  };
}

function handleServerMessage(data) {
  switch (data.type) {
    case "client_ready":
      logMessage("✅ A player is ready!");
      break;
    case "client_unready":
      logMessage("🔴 A player is not ready!");
      break;

    case "session_start":
      logMessage("🚀 Game has started!");
      readyForm.hidden = true;
      break;

    case "client_chat":
      break;

    case "sync":
      break;

    case "error":
      logMessage(`⚠️ Error: ${JSON.stringify(data.error)}`);
      break;

    default:
      logMessage(`ℹ️ Unknown message: ${JSON.stringify(data)}`);
  }
  window.sessionMessage = data;
  updateChatLog(data.session?.chatLog);
  gameStateLog.innerText = JSON.stringify(data.state, null, 2);
}

async function joinGame(gameId) {
  const response = await fetch(`/game/${gameId}`, {
    method: "POST",
    credentials: "same-origin",
  });

  if (response.ok) {
    const url = new URL(window.location.href);
    url.searchParams.set("game_id", gameId);
    window.history.pushState(null, "", url.toString());
    connectToGame(gameId);
  } else {
    logMessage("❌ Failed to join game");
  }
}

async function createGameSubmitHandler(e) {
  e.preventDefault();
  const response = await fetch("/game", {
    method: "POST",
    credentials: "same-origin",
  });

  if (response.ok) {
    const data = await response.json();
    logMessage(`🎮 Created game: ${data.id}`);
    updateUIForConnectedGame(data.id);

    const url = new URL(window.location.href);
    url.searchParams.set("game_id", data.id);
    window.history.pushState(null, "", url.toString());
    connectToGame(data.id);
  } else {
    logMessage("❌ Failed to create game");
  }
}

async function joinGameSubmitHandler(e) {
  e.preventDefault();
  const gameId = joinGameInput.value.trim();
  if (!gameId) return;
  joinGame(gameId);
}

async function chatSubmitHandler(e) {
  e.preventDefault();
  const chatMessage = chatInput.value.trim();
  if (!chatMessage || conn.readyState !== WebSocket.OPEN) return;

  conn.send(
    JSON.stringify({
      type: "chat",
      data: { chat: chatMessage },
    }),
  );

  chatInput.value = "";
}

async function readySubmitHandler(e) {
  e.preventDefault();
  if (conn.readyState !== WebSocket.OPEN) return;

  isReady = !isReady;
  conn.send(
    JSON.stringify({
      type: "set_ready",
      data: { ready: isReady },
    }),
  );

  logMessage(isReady ? "🟢 You are ready!" : "🔴 You are not ready.");
}

async function reconnectToGame() {
  const queryParams = new URLSearchParams(window.location.search);
  const gameId = queryParams.get("game_id");
  if (gameId) joinGame(gameId);
}

createGameForm.onsubmit = createGameSubmitHandler;
joinGameForm.onsubmit = joinGameSubmitHandler;
chatForm.onsubmit = chatSubmitHandler;
readyForm.onsubmit = readySubmitHandler;

reconnectToGame();
