const WIDTH = 16;
const HEIGHT = 16;

let animationCanvas, animationFrames, animationIndex;
let animationRunning = false;

export async function getImagePixels(file) {
  const contents = await loadFile(file);
  const image = await decodeImage(contents);
  const pixels = toPixels(image, WIDTH, HEIGHT);
  return pixels;
}

export async function getAnimationFrames(file) {
  const contents = await loadFile(file);
  let frame = 0;
  const frames = [];
  while (true) {
    try {
      const image = await decodeGIFFrame(contents, frame++);
      frames.push({
        pixels: toPixels(image, WIDTH, HEIGHT),
        duration: (image.duration ?? 100000) / 1000,
      });
    } catch (e) {
      // We expect a RangeError here when we've read all the frames. That just
      // means we're done.
      if (e instanceof RangeError) return frames;
      throw new Error("Could not decode the file as a GIF image");
    }
  }
}

export function imagePixelsToCanvas(byteArray, canvas) {
  const imageData = new ImageData(
    new Uint8ClampedArray(byteArray),
    WIDTH,
    HEIGHT
  );
  const context = canvas.getContext("2d");
  canvas.width = imageData.width;
  canvas.height = imageData.height;
  context.putImageData(imageData, 0, 0);
}

export function animationFramesToCanvas(frames, canvas) {
  animationCanvas = canvas;
  animationFrames = frames;
  animationIndex = 0;

  if (!animationRunning) {
    startAnimation();
    animationRunning = true;
  }
}

function loadFile(file) {
  return new Promise((resolve, reject) => {
    var reader = new FileReader();
    reader.addEventListener("load", (e) => resolve(e.target.result));
    reader.readAsArrayBuffer(file);
  });
}

async function decodeImage(data) {
  try {
    return await new ImageDecoder({ data, type: "image/png" }).decode();
  } catch (e) {}
  try {
    return await new ImageDecoder({ data, type: "image/jpg" }).decode();
  } catch (e) {}
  try {
    return await new ImageDecoder({ data, type: "image/gif" }).decode();
  } catch (e) {}
  throw new Error(
    "Could not decode the file as an image of type png, jpg or gif"
  );
}

function decodeGIFFrame(image, frameIndex) {
  return new ImageDecoder({ data: image, type: "image/gif" }).decode({
    frameIndex,
  });
}

function toPixels(frame, width, height) {
  const canvas = document.createElement("canvas");
  const context = canvas.getContext("2d");
  canvas.width = width;
  canvas.height = height;
  context.drawImage(frame.image, 0, 0, width, height);
  return [...context.getImageData(0, 0, canvas.width, canvas.height).data];
}

async function startAnimation() {
  animationCanvas.width = WIDTH;
  animationCanvas.height = HEIGHT;
  while (true) {
    const frame = animationFrames[animationIndex++];
    if (animationIndex == animationFrames.length) animationIndex = 0;
    const context = animationCanvas.getContext("2d");
    const imageData = new ImageData(
      new Uint8ClampedArray(frame.pixels),
      WIDTH,
      HEIGHT
    );
    context.putImageData(imageData, 0, 0);
    await sleep(frame.duration);
  }
}

let timeout;
async function sleep(duration, abortPrevious = false) {
  if (abortPrevious) clearTimeout(timeout);
  return new Promise((resolve, reject) => {
    timeout = setTimeout(resolve, duration);
  });
}
