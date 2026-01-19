export default class DataProxy {
  constructor(data) {
    this._listeners = [];
    data._path = [];
    data.addEventListener = (e, f) => this.addEventListener(e, f);
    this._data = this._wrap(data);
  }

  addEventListener(evnt, listener) {
    // Only accepts change listeners for now
    if (evnt != "change")
      throw new Error("Can currently only support 'change' events");
    if (typeof listener != "function")
      throw new Error("Event listener should be a function");
    this._listeners.push(listener);
  }

  _wrap(data, prefix = []) {
    for (const key of Object.keys(data)) {
      if (data[key] != null && typeof data[key] == "object" && key != "_path") {
        data[key]._path = [...prefix, key];
        data[key] = this._wrap(data[key], [...prefix, key]);
      }
    }
    return new Proxy(data, this);
  }

  set(obj, prop, value) {
    const path = [...obj._path, prop];
    const oldValue = obj[prop];
    const result = Reflect.set(...arguments);
    const p = path.join(".");
    this._listeners.forEach((f) => f(p, oldValue, value));
    return result;
  }

  getProxy() {
    return this._data;
  }

  store(path, value) {
    if (!this.validPathString(path))
      throw new Error("attempted to store to invalid path");

    // Find the right place in the object hierarchy based on the path
    const parts = path.split(".");
    let target = this._data;
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

    if (target[leaf] == null) {
      target[leaf] = this.toValue(value);
    } else {
      target[leaf] = this._cast(value, typeof target[leaf]);
    }
  }

  _cast(value, type) {
    switch (type) {
      case "number":
        return Number.parseFloat(value);
      case "string":
        return `${value}`;
      case "boolean":
        return !!value;
      case "undefined":
        // The target has no type, so try to infer from the value itself
        return this.toValue(value);
      case "object":
        if (typeof value == "object") return value;
      // else fall through to default
      default:
        throw new Error(`Can't cast a value to type ${type}`);
    }
  }

  toValue(expression) {
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

  validPathString(path) {
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

  pathExists(path) {
    if (!this.validPathString(path)) return false;
    const parts = path.split(".");
    let target = this._data;
    for (const p of parts) {
      if (target == null || typeof target !== "object" || !(p in target)) {
        return false;
      }
      target = target[p];
    }
    return true;
  }

  retrieve(path) {
    if (!this.validPathString(path))
      throw new Error("attempted to retrieve from invalid path: " + path);
    const parts = path.split(".");
    let target = this._data;
    for (const p of parts) {
      if (target == null || typeof target !== "object" || !(p in target)) {
        throw new Error(`Data object has no property ${p} for path ${path}`);
      }
      target = target[p];
    }
    return target;
  }
}
