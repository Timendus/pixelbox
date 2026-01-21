import Databind from "./databind.js";
import {
  imagePixelsToCanvas,
  animationFramesToCanvas,
  getImagePixels,
  getAnimationFrames,
  canvasToImagePixels,
  emptyImageArray,
} from "./images.js";

// We need this to be able to get started
document.getElementById("addScene").addEventListener("click", () =>
  call(
    "/scene/",
    {
      method: "PUT",
    },
    true
  )
);

// Put the rest of the stuff in motion
let globalState;
let drawing;
const data = await loadDataObject();
if (data != null) {
  const dataProxy = new Databind("body", data, {
    immediate: true,
    customConversions: [
      {
        selector: "canvas#imageCanvas",
        toDom: (path, newVal, elm) => {
          if (drawing) return;
          if (!newVal) {
            imagePixelsToCanvas(emptyImageArray, elm);
          } else {
            imagePixelsToCanvas(newVal, elm);
          }
        },
      },
      {
        selector: "canvas#animationCanvas",
        toDom: (path, newVal, elm) => {
          if (drawing) return;
          if (!newVal) {
            animationFramesToCanvas([{ pixels: emptyImageArray }], elm);
          } else {
            animationFramesToCanvas(newVal, elm);
          }
        },
      },
    ],
  });
  globalState = dataProxy.data;
  applyRelationships(dataProxy);
  registerEventHandlers();
  connectToSSE();
}
document.querySelector(".overlay").classList.add("hidden");

// Aaand done!

async function loadDataObject() {
  const data = {
    paint: {
      tool: "pen",
      fgColor: "#ffffff",
      bgColor: "#000000",
    },
    autoPreview: true,
    showMessages: false,
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

function applyRelationships(dataProxy) {
  // Some more complicated relationships between properties than what we can
  // comfortably model with DOM attributes go here

  const data = dataProxy.data;
  let restoreBrightness = !data.selectedScene.changeBrightness;
  dataProxy.addEventListener("change", (path, oldVal, newVal) => {
    console.info(`Change at ${path} from ${oldVal} to ${newVal}`);
    switch (path) {
      case "scenes":
        // If we add a new scene to the list, select it in the UI (bit of a
        // hack, since scenes can be added by other users, but we're kinda
        // assuming single-user operation anyway)
        if (oldVal.length < newVal.length)
          data.selectedScene = data.scenes[data.scenes.length - 1];

        // Refresh the data in our selected scene with the latest info received
        // from the server (which may delete the current scene)
        let selected = data.scenes.find(
          (scene) => scene.uuid == data.selectedScene.uuid
        );
        if (!selected) selected = data.scenes[0];
        data.selectedScene = selected;
        break;

      case "selectedScene.volume":
        // Check the change volume checkbox if slider moved
        data.selectedScene.changeVolume = true;
        break;

      case "selectedScene.brightness":
        // Check the change brightness checkbox if slider moved
        data.selectedScene.changeBrightness = true;
        break;

      case "selectedScene.changeBrightness":
        // If we've manually changed the brightness settings, don't restore them
        // when changing tabs
        restoreBrightness = false;
        break;

      case "selectedScene.sceneType":
        // When user changes to light tab, auto-select brightness checkbox
        if (oldVal != "light" && newVal == "light") {
          const oldChangeBrightness = data.selectedScene.changeBrightness;
          data.selectedScene.changeBrightness = true;
          if (!oldChangeBrightness) {
            restoreBrightness = true;
          }
        }
        // When user changes away from light tab, restore brightness checkbox if
        // it was unchecked previously and user hasn't manually touched it
        if (oldVal == "light" && newVal != "light") {
          if (restoreBrightness) {
            data.selectedScene.changeBrightness = false;
          }
        }
        break;

      case "selectedScene.name":
        data.selectedScene.id = toId(newVal);
        break;
    }

    if (data.autoPreview && path.startsWith("selectedScene")) {
      call("/apply/preview", {
        method: "POST",
        body: JSON.stringify(data.selectedScene),
      });
    }
  });
}

function toId(name) {
  return encodeURI(name.toLowerCase().replaceAll(" ", "-"));
}

function registerEventHandlers() {
  document
    .getElementById("syncTime")
    .addEventListener("click", () => call("/apply/syncTime"));

  document.getElementById("preview").addEventListener("click", () =>
    call("/apply/preview", {
      method: "POST",
      body: JSON.stringify(globalState.selectedScene),
    })
  );

  document.getElementById("save").addEventListener("click", () =>
    call(
      "/scene/" + globalState.selectedScene.uuid,
      {
        method: "POST",
        body: JSON.stringify(globalState.selectedScene),
      },
      true
    )
  );

  document.getElementById("deleteScene").addEventListener("click", () =>
    call(
      "/scene/" + globalState.selectedScene.uuid,
      {
        method: "DELETE",
      },
      true
    )
  );

  document.getElementById("imageFile").addEventListener("change", async (e) => {
    if (e.target.files.length != 1) {
      return showMessage("Expected user to select an image file", true);
    }
    globalState.selectedScene.image.pixels = await getImagePixels(
      e.target.files[0]
    );
  });

  document
    .getElementById("animationFile")
    .addEventListener("change", async (e) => {
      if (e.target.files.length != 1) {
        return showMessage("Expected user to select an image file", true);
      }
      globalState.selectedScene.animation.frames = await getAnimationFrames(
        e.target.files[0]
      );
    });

  const canvas = document.querySelector("canvas#imageCanvas");
  const context = canvas.getContext("2d");
  let pixWidth, pixHeight, selectedTool;

  canvas.addEventListener("mousedown", (e) => {
    const rect = canvas.getClientRects();
    pixWidth = rect[0].width / 16;
    pixHeight = rect[0].height / 16;

    if (globalState.paint.tool == "picker") {
      const imageData = context.getImageData(
        Math.floor(e.offsetX / pixWidth),
        Math.floor(e.offsetY / pixHeight),
        1,
        1
      );
      const r = imageData.data[0].toString(16).padStart(2, 0);
      const g = imageData.data[1].toString(16).padStart(2, 0);
      const b = imageData.data[2].toString(16).padStart(2, 0);
      if (e.button == 2) {
        globalState.paint.bgColor = `#${r}${g}${b}`;
      } else {
        globalState.paint.fgColor = `#${r}${g}${b}`;
      }
      globalState.paint.tool = "pen";
      return;
    }

    selectedTool = globalState.paint.tool;
    if (e.button == 2) {
      globalState.paint.tool = "eraser";
    }

    drawing = true;
    context.fillStyle =
      globalState.paint.tool == "pen"
        ? globalState.paint.fgColor
        : globalState.paint.bgColor;

    context.fillRect(
      Math.floor(e.offsetX / pixWidth),
      Math.floor(e.offsetY / pixHeight),
      1,
      1
    );
    globalState.selectedScene.image.pixels = canvasToImagePixels(canvas);
  });

  canvas.addEventListener("mousemove", (e) => {
    if (!drawing) return;
    context.fillRect(
      Math.floor(e.offsetX / pixWidth),
      Math.floor(e.offsetY / pixHeight),
      1,
      1
    );
    globalState.selectedScene.image.pixels = canvasToImagePixels(canvas);
  });

  canvas.addEventListener("mouseup", (e) => {
    drawing = false;
    globalState.paint.tool = selectedTool;
  });

  canvas.addEventListener("contextmenu", (e) => {
    e.preventDefault();
    e.stopPropagation();
    return false;
  });
}

function connectToSSE() {
  const evtSource = new EventSource("/events");
  evtSource.addEventListener("message", (e) => {
    showMessage('Device says: "' + e.data + '"');
  });
}

async function call(url, payload, updateScenes = false) {
  let response;
  try {
    response = await fetch(url, payload);
    if (!response.ok) {
      return showMessage(await response.text(), true);
    }
  } catch (e) {
    return showMessage(e, true);
  }
  if (updateScenes) {
    globalState.scenes = await response.json();
  }
}

function showMessage(message, error = false) {
  const item = document.createElement("li");
  item.innerText = message;
  const now = new Date();
  item.setAttribute("data-time", now.getHours() + ":" + now.getMinutes());
  item.classList.toggle("error", error);
  const firstItem = document.querySelector("#messages > li");
  document.getElementById("messages").insertBefore(item, firstItem);
  if (error) globalState.showMessages = true;
}
