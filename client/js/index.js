import Databind from "./databind.js";

// We need this to be able to get started
document.getElementById("addScene").addEventListener("click", async () => {
  if (
    await call("/scene/", {
      method: "PUT",
    })
  )
    document.location.reload();
});

// Put the rest of the stuff in motion
const data = await loadDataObject();
if (data != null) {
  const boundData = new Databind("body", data);
  applyRelationships(boundData);
  registerEventHandlers(boundData);
  connectToSSE();
}
document.querySelector(".overlay").classList.add("hidden");

// Aaand done!

async function loadDataObject() {
  const data = {
    autoPreview: true,
    scenes: [],
    selectedScene: {},
  };

  // Load scenes from server
  try {
    const scenesRequest = await fetch("/scene/");
    if (!scenesRequest.ok) {
      showMessage("Got an error from the server when loading scenes");
      return null;
    }
    data.scenes = await scenesRequest.json();
  } catch (e) {
    showMessage("Could not reach server: " + e);
    return null;
  }

  // Select first scene
  if (!data.scenes || data.scenes.length == 0) {
    showMessage("No scenes found");
    return null;
  }
  data.selectedScene = data.scenes[0];

  return data;
}

function applyRelationships(boundData) {
  // Some more complicated relationships between properties than what we can
  // comfortably model with DOM attributes go here

  let restoreBrightness = !boundData.selectedScene.changeBrightness;
  boundData.addEventListener("change", (path, oldVal, newVal) => {
    switch (path) {
      case "selectedScene.brightness":
        // Check the change brightness checkbox if slider moved
        if (!boundData.selectedScene.changeBrightness)
          boundData.selectedScene.changeBrightness = true;
        break;

      case "selectedScene.changeBrightness":
        // If we've manually changed the brightness settings, don't restore them
        // when changing tabs
        restoreBrightness = false;
        break;

      case "selectedScene.sceneType":
        // When user changes to light tab, auto-select brightness checkbox
        if (oldVal != "light" && newVal == "light") {
          const oldChangeBrightness = boundData.selectedScene.changeBrightness;
          boundData.selectedScene.changeBrightness = true;
          if (!oldChangeBrightness) {
            restoreBrightness = true;
          }
        }
        // When user changes away from light tab, restore brightness checkbox if
        // it was unchecked previously and user hasn't manually touched it
        if (oldVal == "light" && newVal != "light") {
          if (restoreBrightness && boundData.selectedScene.changeBrightness) {
            boundData.selectedScene.changeBrightness = false;
          }
        }
        break;

      case "selectedScene.name":
        boundData.selectedScene.id = toId(newVal);
        break;
    }

    if (boundData.autoPreview) {
      call("/apply/preview", {
        method: "POST",
        body: JSON.stringify(boundData.selectedScene),
      });
    }
  });
}

function toId(name) {
  return encodeURI(name.toLowerCase().replaceAll(" ", "-"));
}

function registerEventHandlers(boundData) {
  document
    .getElementById("syncTime")
    .addEventListener("click", () => call("/apply/syncTime"));

  document.getElementById("preview").addEventListener("click", () =>
    call("/apply/preview", {
      method: "POST",
      body: JSON.stringify(boundData.selectedScene),
    })
  );

  document.getElementById("save").addEventListener("click", () =>
    call("/scene/" + boundData.selectedScene.uuid, {
      method: "POST",
      body: JSON.stringify(boundData.selectedScene),
    })
  );

  document.getElementById("deleteScene").addEventListener("click", async () => {
    if (
      await call("/scene/" + boundData.selectedScene.uuid, {
        method: "DELETE",
      })
    )
      document.location.reload();
  });
}

function connectToSSE() {
  const evtSource = new EventSource("/events");
  evtSource.addEventListener("message", (e) => {
    showMessage('Device says: "' + e.data + '"');
  });
}

async function call(url, payload) {
  try {
    const response = await fetch(url, payload);
    if (!response.ok) {
      showMessage(await response.text(), true);
      return false;
    }
  } catch (e) {
    showMessage(e, true);
    return false;
  }
  return true;
}

function showMessage(message, error = false) {
  const item = document.createElement("li");
  item.innerText = message;
  const now = new Date();
  item.setAttribute("data-time", now.getHours() + ":" + now.getMinutes());
  item.classList.toggle("error", error);
  const firstItem = document.querySelector("#messages > li");
  document.getElementById("messages").insertBefore(item, firstItem);
  document.getElementById("messagesContainer").classList.remove("hidden");
}
