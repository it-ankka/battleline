const messageLog = document.getElementById("message-log");
const chatLog = document.getElementById("chat-log");

const copyGameIdForm = document.getElementById("copy-game-id-form");
const createGameForm = document.getElementById("create-game-form");
const joinGameForm = document.getElementById("join-game-form");
const chatForm = document.getElementById("chat-form");

const copyGameIdInput = document.getElementById("game-id-input");
const joinGameInput = document.getElementById("join-game-id-input");
const chatInput = document.getElementById("chat-input");

let conn;

function connectToGame(gameId, maxRetries = 5) {
  conn = new WebSocket(`ws://${location.host}/ws/${gameId}`);

  conn.onclose = (ev) => {
    console.log(
      `WebSocket Disconnected code: ${ev.code}, reason: ${ev.reason}`,
    );
    if (![1000, 1001, 1008].includes(ev.code) && maxRetries > 0) {
      console.info(
        "Disconnected. Reconnecting in 1s",
        maxRetries,
        "retries remaining",
      );
      setTimeout(() => connectToGame(gameId, maxRetries - 1), 1000);
      return;
    }
    joinGameForm.hidden = false;
    copyGameIdForm.hidden = true;

    const url = new URL(window.location.href);
    url.search = "";
    window.history.pushState(null, "", url.toString());
  };
  conn.onerror = (ev) => {
    console.error("Websocket error", ev);
  };

  conn.onopen = (ev) => {
    // Update UI
    copyGameIdInput.value = gameId;
    joinGameForm.hidden = true;
    copyGameIdForm.hidden = false;
    chatForm.hidden = false;
    console.info("websocket connected");
  };

  // This is where we handle messages received.
  conn.onmessage = (ev) => {
    if (typeof ev.data !== "string") {
      console.error("unexpected message type", typeof ev.data);
      return;
    }
    try {
      const data = JSON.parse(ev.data);
      if (["tick", "chat"].includes(data.type)) {
        chatLog.innerText = data.session.chatLog
          .map((m) => m.content)
          .join("\n\n");
      }
      console.log(data);
    } catch (err) {
      console.error(err);
    }
  };
}

const joinGame = async (gameId) => {
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
    joinGameForm.hidden = false;
    copyGameIdForm.hidden = true;

    const url = new URL(window.location.href);
    url.search = "";
    window.history.pushState(null, "", url.toString());
  }
};

const createGameSubmitHandler = async (e) => {
  e.preventDefault();
  const response = await fetch("/game", {
    method: "POST",
    credentials: "same-origin",
  });

  if (response.ok) {
    const data = await response.json();
    console.log("CREATE GAME RESPONSE", data);
    // Update UI
    copyGameIdInput.value = data.id;
    joinGameForm.hidden = true;
    copyGameIdForm.hidden = false;

    const url = new URL(window.location.href);
    url.searchParams.set("game_id", data.id);
    window.history.pushState(null, "", url.toString());
    connectToGame(data.id);
  } else {
    joinGameForm.hidden = false;
    copyGameIdForm.hidden = true;

    const url = new URL(window.location.href);
    url.search = "";
    window.history.pushState(null, "", url.toString());
  }
};

const joinGameSubmitHandler = async (e) => {
  e.preventDefault();
  const gameId = joinGameInput.value;
  joinGame(gameId);
};

const chatSubmitHandler = async (e) => {
  e.preventDefault();
  const chatMessage = chatInput.value ?? "";
  if (chatMessage.length < 1) return;

  if (conn.readyState === WebSocket.OPEN) {
    conn.send(
      JSON.stringify({
        type: "chat",
        data: {
          chat: chatMessage,
        },
      }),
    );
  }
};

const reconnectToGame = async () => {
  const queryParams = new URLSearchParams(window.location.search);
  const gameId = queryParams.get("game_id");
  if ([undefined, null, ""].includes(gameId)) {
    return;
  }

  joinGame(gameId);
};

createGameForm.onsubmit = createGameSubmitHandler;
joinGameForm.onsubmit = joinGameSubmitHandler;
chatForm.onsubmit = chatSubmitHandler;

reconnectToGame();
