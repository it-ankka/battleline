const messageLog = document.getElementById("message-log");
const createGameForm = document.getElementById("create-game-form");
const joinGameForm = document.getElementById("join-game-form");

createGameForm.onsubmit = async (e) => {
  e.preventDefault();
  const response = await fetch("/game", {
    method: "POST",
    credentials: "same-origin",
  });
  const data = await response.json();
  console.log("CREATE GAME RESPONSE", data);

  function dial() {
    const conn = new WebSocket(`ws://${location.host}/ws/${data.id}`);

    conn.addEventListener("close", (ev) => {
      console.log(
        `WebSocket Disconnected code: ${ev.code}, reason: ${ev.reason}`,
      );
      if (ev.code !== 1001) {
        console.log("Reconnecting in 1s");
        setTimeout(dial, 1000);
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
        console.log(JSON.parse(ev.data));
      } catch (err) {
        console.log(ev.data);
      }
    });
  }
  dial();
};
