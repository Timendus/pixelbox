export default class Databind {
  constructor(scope, data, options = {}) {
    this._scope = scope;
    this._data = new DataProxy(data);
    this._data.addEventListener("change", (p) => this._applyValue(p));
    this._options = this._normalizeOptions(options);

    // Register for DOM change events within `scope`
    Change.instance().register(
      `${scope} [${this._options.bind}], ${scope} [${this._options.write}]`,
      (e) => this._handleChange(e)
    );

    // Register for click events within `scope`
    Click.instance().register(`${scope} [${this._options.click}]`, (e) =>
      this._handleClick(e)
    );

    // Apply the data to the DOM to bring them into initial sync
    this._applyValue();

    return this._data.getProxy();
  }

  _normalizeOptions(options) {
    return Object.assign(
      {
        class: "active",
        stopEvents: true,
        bind: "data-bind",
        write: "data-write",
        read: "data-read",
        click: "data-click",
        activeIf: "data-active-if",
      },
      options
    );
  }

  _handleChange(evnt) {
    // Which element changed?
    const target = evnt.target.closest(
      `[${this._options.bind}], [${this._options.write}]`
    );

    // Apply the changes to the data object
    this._apply(target.getAttribute(this._options.bind), target);
    this._apply(target.getAttribute(this._options.write), target);

    if (this._options.stopEvents) {
      // We're done with this event, don't try to evaluate it any further
      evnt.preventDefault();
      evnt.stopPropagation();
    }
  }

  _handleClick(evnt) {
    // Which element changed?
    const target = evnt.target.closest(`[${this._options.click}]`);

    // Apply the changes to the data object
    this._apply(target.getAttribute(this._options.click), target);

    if (this._options.stopEvents) {
      // We're done with this event, don't try to evaluate it any further
      evnt.preventDefault();
      evnt.stopPropagation();
    }
  }

  // Write changes from DOM to data object, based on expression
  _apply(expression, target) {
    if (!expression) return;
    try {
      let [path, value] = this._parseExpression(expression);
      if (value == null) {
        if (target.matches("input[type=checkbox")) {
          value = target.checked;
        } else if ("value" in target) {
          value = target.value;
        } else throw new Error("no value to write to data object");
      }
      this._data.store(path, value);
    } catch (e) {
      console.error(e);
    }
  }

  _parseExpression(expression) {
    if (this._data.validPath(expression)) {
      return [expression, null];
    }
    let path = null;
    if (expression.match(`^[^=!]+=[^=].*$`) != null) {
      // This translates to: "an equals sign not preceded or directly followed
      // by another equals sign". If we match that, we have an assignment in the
      // form of "something=anything".
      path = expression.split("=")[0];
      expression = expression.substring(expression.indexOf("=") + 1);
    }
    if (expression.includes("==")) {
      const parts = expression.split("==");
      return [path, this._toValue(parts[0]) == this._toValue(parts[1])];
    }
    if (expression.includes("!=")) {
      const parts = expression.split("!=");
      return [path, this._toValue(parts[0]) != this._toValue(parts[1])];
    }
    if (path != null) {
      return [path, this._toValue(expression)];
    }
    throw "Could not parse expression: " + expression;
  }

  _toValue(expression) {
    switch (expression) {
      case "undefined":
        return undefined;
      case "null":
        return null;
      case "true":
        return true;
      case "false":
        return false;
    }
    const number = Number.parseFloat(expression);
    if (!isNaN(number)) {
      return number;
    }
    try {
      return this._data.retrieve(expression);
    } catch (e) {
      return expression;
    }
  }

  // Write changes from data object to DOM
  _applyValue(path) {
    // Update values (for bound form elements)
    document
      .querySelectorAll(
        `${this._scope} [${this._options.read}], ${this._scope} [${this._options.bind}]`
      )
      .forEach((e) => {
        const readExpression = e.getAttribute(this._options.read);
        if (readExpression && (path == null || readExpression.includes(path))) {
          const [path, val] = this._parseExpression(readExpression);
          if (e.matches("input[type=checkbox")) {
            e.checked = val ?? this._data.retrieve(path);
          } else if ("value" in e) {
            e.value = val ?? this._data.retrieve(path);
          } else
            throw new Error("don't know how to write value to DOM element");
        }

        const bindExpression = e.getAttribute(this._options.bind);
        if (bindExpression && (path == null || bindExpression.includes(path))) {
          const [path, val] = this._parseExpression(bindExpression);
          if (e.matches("input[type=checkbox")) {
            e.checked = val ?? this._data.retrieve(path);
          } else if ("value" in e) {
            e.value = val ?? this._data.retrieve(path);
          } else
            throw new Error("don't know how to write value to DOM element");
        }
      });

    // Update classes (for active-if expressions)
    document
      .querySelectorAll(`${this._scope} [${this._options.activeIf}]`)
      .forEach((e) => {
        const activeIfExpression = e.getAttribute(this._options.activeIf);
        if (path != null && !activeIfExpression.includes(path)) return;
        const [p, val] = this._parseExpression(activeIfExpression);
        e.classList.toggle("active", !!(val ?? this._data.retrieve(p)));
      });
  }
}

class DataProxy {
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
      if (typeof data[key] == "object" && key != "_path") {
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
    if (!this.validPath(path))
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
    target[leaf] = this._cast(value, typeof target[leaf]);
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
        return this._toValue(value);
      default:
        throw new Error(`Can't cast a value to type ${type}`);
    }
  }

  _toValue(expression) {
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

  retrieve(path) {
    if (!this.validPath(path))
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

  validPath(path) {
    // A valid property of an object may not start with a number
    const validStartChar = "[a-zA-Z_]";
    const validFollowingChar = "[a-zA-Z_0-9]";
    const validProperty = `${validStartChar}+${validFollowingChar}*`;
    // We're looking for at least one valid property, or a chain of valid
    // properties separated by periods.
    return path.match(`^${validProperty}(\\.${validProperty})*$`) != null;
  }
}

class Change {
  constructor() {
    this._handlers = {};
    document.addEventListener("change", (e) => this._callHandler(e));
  }

  register(selector, handler) {
    this._handlers[selector] = this._handlers[selector] || [];
    this._handlers[selector].push(handler);
  }

  _callHandler(e) {
    Object.keys(this._handlers).forEach((selector) => {
      if (e.target.closest(selector) !== null) {
        this._handlers[selector].forEach((handler) => {
          if (typeof handler == "function" && !e.defaultPrevented)
            handler(e, selector);
        });
      }
    });
  }
}

Change.instance = function () {
  if (!!Change._instance) return Change._instance;
  return (Change._instance = new Change());
};

/*
 * This class installs one single click handler on the whole document, and
 * evaluates which callback to call at click time, based on the element that has
 * been clicked. This allows us to swap out and rerender whole sections of the
 * DOM without having to reinstall a bunch of click handlers each time. This
 * nicely decouples the render logic from the click event management logic.
 *
 * To make sure we really only install a single click handler, you can use the
 * singleton pattern and ask for `Click.instance()` instead of creating a new
 * object.
 */

class Click {
  constructor() {
    this._handlers = {};

    document.addEventListener("click", (e) => this._callHandler("click", e));
    document.addEventListener("mousedown", (e) =>
      this._callHandler("mousedown", e)
    );
    document.addEventListener("mouseup", (e) =>
      this._callHandler("mouseup", e)
    );
  }

  register(
    selector,
    handlers = { click: null, mousedown: null, mouseup: null }
  ) {
    if (typeof handlers == "function") handlers = { click: handlers };
    this._handlers[selector] = this._handlers[selector] || [];
    this._handlers[selector].push(handlers);
  }

  _callHandler(type, e) {
    Object.keys(this._handlers).forEach((selector) => {
      if (e.target.closest(selector) !== null) {
        const handlers = this._handlers[selector].map((h) => h[type]);
        handlers.forEach((handler) => {
          if (typeof handler == "function" && !e.defaultPrevented)
            handler(e, selector);
        });
      }
    });
  }
}

Click.instance = function () {
  if (!!Click._instance) return Click._instance;
  return (Click._instance = new Click());
};
