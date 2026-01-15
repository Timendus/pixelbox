import Databind from "./databind.js";

const data = {
  autoPreview: true,
  scenes: ["Henk", "Sjon", "Ingrid", "Ali"],
  selectedScene: {
    name: "Henk",
    id: "henk",
    changeBrightness: false,
    brightness: 100,
    changeVolume: false,
    volume: undefined,
    sceneType: "clock",
    clock: {
      enabled: true,
      type: "FULL_SCREEN",
      color: "#005500",
    },
    weather: {
      enabled: false,
      type: "RAIN",
    },
    temperature: {
      enabled: false,
      temperature: 20,
    },
    calendar: {
      enabled: false,
    },
    light: {
      type: "PLAIN",
      color: "#FFFF00",
    },
    effect: {
      type: "VJ",
      vjType: 1,
      visualisationType: 5,
      scoreRedPlayer: 10,
      scoreBluePlayer: undefined,
    },
  },
};

const boundData = new Databind("body", data);

// Some more complicated relations between properties go here:
let restoreBrightness = !boundData.selectedScene.changeBrightness;
boundData.addEventListener("change", (path, oldVal, newVal) => {
  switch (path) {
    case "selectedScene.brightness":
      // Check the change brightness checkbox if slider moved
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
        if (restoreBrightness) {
          boundData.selectedScene.changeBrightness = false;
        }
      }
      break;
  }

  if (boundData.autoPreview)
    fetch("/apply/scene", {
      method: "POST",
      body: JSON.stringify(boundData.selectedScene),
    });
});

window.data = boundData;

document
  .getElementById("syncTime")
  .addEventListener("click", () => fetch("/apply/syncTime"));

document.getElementById("preview").addEventListener("click", () =>
  fetch("/apply/scene", {
    method: "POST",
    body: JSON.stringify(boundData.selectedScene),
  })
);

document.getElementById("save").addEventListener("click", () =>
  fetch("/scene/update", {
    method: "POST",
    body: JSON.stringify(boundData.selectedScene),
  })
);
