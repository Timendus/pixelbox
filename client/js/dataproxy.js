export class DataProxy {
  constructor(data) {
    this._listeners = [];
    this.data = new Wrapper(data, [], (path, oldVal, newVal) =>
      this._listeners.forEach((f) => f(path, oldVal, newVal))
    );
  }

  /* Public API */

  addEventListener(evnt, listener) {
    // Only accepts change listeners for now
    if (evnt != "change")
      throw new Error("Can currently only support 'change' events");
    if (typeof listener != "function")
      throw new Error("Event listener should be a function");
    this._listeners.push(listener);
  }

  /* API for databind.js */

  store(path, value) {
    if (!validPathString(path))
      throw new Error("attempted to store to invalid path");

    // Find the right place in the object hierarchy based on the path
    const parts = path.split(".");
    let target = this.data;
    for (let i = 0; i < parts.length - 1; i++) {
      const p = parts[i];
      if (target == null || typeof target !== "object" || !(p in target)) {
        throw new Error(`Data object has no property ${p} for path ${path}`);
      }
      target = target[p];
    }

    // We have stopped a bit too early, so we can actually store the value
    // (`target = value` just overwrites the pointer, `target[leaf] = value`
    // actually stores in it).
    const leaf = parts[parts.length - 1];
    if (target == null || typeof target !== "object" || !(leaf in target)) {
      throw new Error(`Data object has no property ${leaf} for path ${path}`);
    }

    target[leaf] = cast(value, target[leaf]);
  }

  retrieve(path) {
    if (!validPathString(path))
      throw new Error("attempted to retrieve from invalid path: " + path);
    const parts = path.split(".");
    let target = this.data;
    for (const p of parts) {
      if (target == null || typeof target !== "object" || !(p in target)) {
        throw new Error(`Data object has no property ${p} for path ${path}`);
      }
      target = target[p];
    }
    return target;
  }

  pathExists(path) {
    try {
      this.retrieve(path);
      return true;
    } catch (e) {
      return false;
    }
  }
}

function validPathString(path) {
  if (typeof path != "string") return false;
  // A valid property of an object may not start with a number, but we do
  // accept number-only indices for arrays
  const validStartChar = "[a-zA-Z_]";
  const validFollowingChar = "[a-zA-Z_0-9]";
  const validIndex = "[0-9]+";
  const validProperty = `(${validStartChar}+${validFollowingChar}*|${validIndex})`;
  // We're looking for at least one valid property, or a chain of valid
  // properties separated by periods.
  return path.match(`^${validProperty}(\\.${validProperty})*$`) != null;
}

function cast(value, oldValue) {
  if (oldValue === null) {
    // Because `typeof null == "object"` :(
    oldValue = undefined;
  }
  switch (typeof oldValue) {
    case "number":
      return Number.parseFloat(value);
    case "string":
      return `${value}`;
    case "boolean":
      return !!value;
    case "undefined":
      // The target has no type, so try to infer from the value itself
      return toValue(value);
    case "object":
      if (typeof value == "object") {
        return value;
      } else {
        // fall through to default below
      }
    default:
      throw new Error(`Can't cast a value to type ${type}`);
  }
}

export function toValue(expression) {
  switch (expression) {
    case "undefined":
      return undefined;
    case "true":
      return true;
    case "false":
      return false;
  }
  const number = Number.parseFloat(expression);
  if (!isNaN(number)) {
    return number;
  }
  return expression;
}

class Wrapper {
  constructor(data, path, notify) {
    this.data = data;
    this.path = path;
    this.notify = notify;
    return new Proxy(data, this);
  }

  set(obj, prop, value, receiver) {
    // This would be a no-op, so ignore and don't bother listeners
    if (obj[prop] === value) {
      return true;
    }

    // Warn user if they are trying to use internal properties
    if (["__isProxied", "__path", "__originalData"].includes(prop)) {
      throw new Error(
        `Can't set property '${prop}' on proxy object; internal use only`
      );
    }

    // When assigning other proxy objects to properties, make a deep copy
    if (typeof value == "object" && value.__isProxied) {
      value = structuredClone(value.__originalData);
    }

    // Apply change
    const oldValue = obj[prop];
    const result = Reflect.set(obj, prop, value, receiver);

    // Notify listeners
    const p = [...this.path, prop].join(".");
    if (this.notify) this.notify(p, oldValue, value);

    return result;
  }

  get(obj, prop) {
    switch (prop) {
      case "__isProxied":
        return true;
      case "__path":
        return this.path;
      case "__originalData":
        return this.data;
    }

    // When getting an object, make sure it is wrapped in a RealProxy too
    if (obj[prop] != null && typeof obj[prop] == "object") {
      return new Wrapper(obj[prop], [...this.path, prop], this.notify);
    }

    // Otherwise, just get me the property of the object
    return Reflect.get(...arguments);
  }
}
