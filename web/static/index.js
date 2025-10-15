const messageLog = document.getElementById("message-log");
const createGameForm = document.getElementById("create-game-form");
const joinGameForm = document.getElementById("join-game-form");
const copyGameIdInput = document.getElementById("game-id-input");
const joinGameInput = document.getElementById("join-game-id-input");
const joinGameButton = document.getElementById("join-game-button");
const copyGameIdForm = document.getElementById("copy-game-id-form");

function connectToGame(gameId) {
  const conn = new WebSocket(`ws://${location.host}/ws/${gameId}`);

  conn.addEventListener("close", (ev) => {
    console.log(
      `WebSocket Disconnected code: ${ev.code}, reason: ${ev.reason}`,
    );
    if (ev.code !== 1001) {
      console.log("Reconnecting in 1s");
      setTimeout(connectToGame, 1000);
    }
  });
  conn.addEventListener("open", (ev) => {
    console.info("websocket connected");
  });

  // This is where we handle messages received.
  conn.addEventListener("message", (ev) => {
    if (typeof ev.data !== "string") {
      console.error("unexpected message type", typeof ev.data);
      return;
    }
    try {
      jsonString = JSON.stringify(JSON.parse(ev.data), null, 2);
      messageLog.innerText = jsonString;

      console.log(JSON.parse(ev.data));
    } catch (err) {
      console.log(ev.data);
    }
  });
}

const joinGame = async (gameId) => {
  const response = await fetch(`/game/${gameId}`, {
    method: "POST",
    credentials: "same-origin",
  });
  const data = await response.json();
  console.log("JOIN GAME RESPONSE", data);
  copyGameIdInput.value = data.id;
  joinGameForm.hidden = true;
  copyGameIdForm.hidden = false;

  if (response.ok) {
    const url = new URL(window.location.href);
    url.searchParams.set("gameId", data.id);
    window.history.pushState(null, "", url.toString());
    connectToGame(data.id);
  }
};

const createGameSubmitHandler = async (e) => {
  e.preventDefault();
  const response = await fetch("/game", {
    method: "POST",
    credentials: "same-origin",
  });
  const data = await response.json();
  console.log("CREATE GAME RESPONSE", data);
  copyGameIdInput.value = data.id;
  joinGameForm.hidden = true;
  copyGameIdForm.hidden = false;

  if (response.ok) {
    const url = new URL(window.location.href);
    url.searchParams.set("gameId", data.id);
    window.history.pushState(null, "", url.toString());
    connectToGame(data.id);
  }
};

const joinGameSubmitHandler = async (e) => {
  e.preventDefault();
  gameId = joinGameInput.value;
  joinGame(gameId);
};

const reconnectToGame = async () => {
  const queryParams = new URLSearchParams(window.location.search);
  const gameId = queryParams.get("gameId");
  if (gameId?.length > 0) {
    try {
      await joinGame(gameId);
    } catch (err) {
      try {
        connectToGame(gameId);
        copyGameIdInput.value = gameId;
        joinGameForm.hidden = true;
        copyGameIdForm.hidden = false;
      } catch (err) {
        const url = new URL(window.location.href);
        url.searchParams = new URLSearchParams();
        window.history.pushState(null, "", url.toString());
      }
    }
  }
};

createGameForm.onsubmit = createGameSubmitHandler;
joinGameForm.onsubmit = joinGameSubmitHandler;

reconnectToGame();
