/**
 * This class allows you to bind data to DOM elements, either one way or
 * bi-directional. This enables you to keep your domain model in sync with
 * what's shown in the browser.
 *
 * Bind part of the DOM to a data object:
 *
 * ```js
 * const dataProxy = new Databind("#app", {
 *   user: {
 *     name: "Alice",
 *     visible: true,
 *     role: "admin",
 *   },
 *   clickCounter: 0,
 * });
 * ```
 *
 * Then you can use the following attributes in the DOM to bind to that data:
 *
 *  - `data-bind`: Two-way binding. The value of the attribute is a path in the
 *    data object. The value of the DOM element will be kept in sync with the
 *    value in the data object, and changes to the DOM element will be written
 *    back to the data object.
 *  - `data-read`: One-way binding. The value of the attribute is a path in the
 *    data object. Changes to the data object at the given path will get synced
 *    to the DOM element.
 *  - `data-write`: One-way binding. The value of the attribute is a path in the
 *    data object. Changes to the DOM element will be written to the data object
 *    at that path.
 *  - `data-click`: The value of the attribute is an assignment expression. When
 *    the element is clicked, the expression will be evaluated and the result
 *    will be stored in the data object. For example,
 *    `data-click="counter=counter+1"` will increment the `counter` property in
 *    the data object when the element is clicked.
 *  - `data-active-if`: The value of the attribute is a path or an expression
 *    (see below). If the expression evaluates to a truthy value, the class
 *    specified in the `class` option (default: "active") will be added to the
 *    element. Otherwise, it will be removed.
 *
 * For example:
 *
 * ```html
 * <div id="app" data-active-if="user.visible">
 *   <input type="text" data-bind="user.name">
 *   <p data-read="user.role"></p>
 *   <button data-click="user.role=admin">Admin</button>
 *   <button data-click="user.role=contrib">Contributor</button>
 *   <p data-read="clickCounter"></p>
 *   <button id="counter">Click me</button>
 * </div>
 * ```
 *
 * The constructor returns a proxy object which you can use to read and write to
 * the data object. Changes you make through this proxy will be reflected in the
 * DOM:
 *
 * ```js
 * dataProxy.data.user.visible = false;   // This will hide the entire div with id "app"
 * console.log(dataProxy.data.user.name); // This will log "Alice"
 *
 * // Listen for clicks on the button, and update the counter shown
 * document.getElementById("counter").addEventListener("click", () =>
 *   dataProxy.data.clickCounter++);
 * ```
 *
 * You can also listen for changes to the data object by adding event listeners
 * to the returned object:
 *
 * ```js
 * dataProxy.addEventListener("change", (path, oldValue, newValue) => {
 *   console.log(`Change at ${path} from ${oldValue} to ${newValue}`);
 * });
 * ```
 *
 * ## Expressions
 *
 * The values of the `data-active-if` and `data-click` attributes can be more
 * than just paths in the data object. They allow for simple expressions that
 * include comparisons and assignments. For example, `data-active-if="user.age
 * >= 18"` will add the active class to the element if the user's age is greater
 * than or equal to 18. `data-active-if="user.name==Alice"` will add the active
 * class if the user's name is Alice. The supported comparisons are `==`, `!=`,
 * `>`, `<`, `>=`, and `<=`.
 *
 * The `data-click` attribute supports assignment expressions. For example,
 * `data-click="user.name=John"` will set the user's name to John when the
 * element is clicked.
 *
 * If you want to do something more complicated than what these simple
 * expressions allow, just write proper event listeners that manipulate the data
 * object, like in the examples above.
 *
 * ## Loops
 *
 * You can also use the `data-loop` attribute to repeat a section of the DOM for
 * each element in an array in the data object. The value of the `data-loop`
 * attribute should be an assignment expression where the assignee is a variable
 * name and the assigned value is a path to an array in the data object. For
 * example, `data-loop="item=items"` will repeat the section of the DOM for each
 * element in the `items` array in the data object, and within that section you
 * can use `item` to refer to the current element. The original contents of the
 * element with the `data-loop` attribute will be used as a template for each
 * repeated section.
 *
 * For example:
 *
 * ```html
 * <ul data-loop="item=items">
 *   <li data-read="item"></li>
 * </ul>
 * ```
 *
 * ## Options
 *
 * The DataBind constructor takes an optional third argument with the following
 * options:
 *
 * - `class`: The class to toggle for `data-active-if` expressions (default:
 *   "active")
 * - `stopEvents`: Whether to call `preventDefault` and `stopPropagation` on
 *   events after handling them (default: true)
 * - `immediate`: Whether to listen for `input` events instead of `change`
 *   events for form elements (default: false). If true, changes to form
 *   elements will be written to the data object immediately as the user types,
 *   instead of only when they blur the element.
 * - `customConversions`: An array of custom conversion objects. Each object
 *   should have a `selector` property, and optionally `toDom` and `fromDom`
 *   functions. If an element matches the selector, the corresponding conversion
 *   functions will be used to convert values between the data object and the
 *   DOM, instead of the default conversions. The `toDom` function takes three
 *   arguments: the path that changed in the data object, the new value at that
 *   path, and the DOM element to update. The `fromDom` function takes two
 *   arguments: the path to write to in the data object, and the DOM element to
 *   read from. It should return the value to write to the data object.
 *
 * For example:
 *
 * ```js
 * const dataProxy = new Databind("#app", {
 *   date: new Date(),
 * }, {
 *   class: "enabled",
 *   immediate: true,
 *   customConversions: [
 *     {
 *       selector: "input[type=date]",
 *       toDom: (path, value, element) => element.value = value.toISOString().substring(0, 10),
 *       fromDom: (path, element) => new Date(element.value),
 *     },
 *   ],
 * });
 * ```
 *
 * You can also override the attribute names for the bindings by passing
 * different values in the options object. For example, if you want to use
 * `data-model` instead of `data-bind` and `data-is-active` instead of
 * `data-active-if` for some reason, you can do:
 *
 * ```js
 * const dataProxy = new Databind("#app", {...}, {
 *   bind: "data-model",
 *   activeIf: "data-is-active"
 * });
 * ```
 *
 * @module
 */

import { DataProxy, toValue } from "./dataproxy.js";

export default class Databind {
  constructor(scope, data, options = {}) {
    this._scope = scope;
    this._data = new DataProxy(data);
    this._data.addEventListener("change", (p) => this._applyValue(p));
    this._options = this._normalizeOptions(options);

    // Register for form element change events
    if (this._options.immediate) {
      document.addEventListener("input", (e) => this._handleChange(e));
    } else {
      document.addEventListener("change", (e) => this._handleChange(e));
    }

    // Register for click events
    document.addEventListener("click", (e) => this._handleClick(e));

    // Apply the data to the DOM to bring them into initial sync
    this._applyValue();

    return this._data;
  }

  _normalizeOptions(options) {
    return Object.assign(
      {
        // Settings
        class: "active",
        stopEvents: true,
        immediate: false,
        customConversions: [],

        // Attribute names
        bind: "data-bind",
        write: "data-write",
        read: "data-read",
        click: "data-click",
        activeIf: "data-active-if",
        loop: "data-loop",
      },
      options
    );
  }

  _parseExpression(expression) {
    if (expression.match(`^[^=!]+=[^=].*$`) != null) {
      // This translates to: "an equals sign not part of a comparison". If we
      // match that, we have an assignment in the form of "something=anything".
      const assigned = this._parseExpression(
        expression.substring(expression.indexOf("=") + 1)
      );
      return {
        type: assigned.type == "unparseable" ? "unparseable" : "assignment",
        assignee: expression.split("=")[0],
        assigned: assigned,
        value: assigned.value,
      };
    }

    const comparisons = {
      "==": (a, b) => a == b,
      "!=": (a, b) => a != b,
      ">": (a, b) => a > b,
      "<": (a, b) => a < b,
      "<=": (a, b) => a <= b,
      ">=": (a, b) => a >= b,
    };
    for (const comp in comparisons) {
      if (expression.includes(comp)) {
        const parts = expression.split(comp);
        const left = this._parseExpression(parts[0]);
        const right = this._parseExpression(parts[1]);
        return {
          type:
            left.type == "unparseable" || right.type == "unparseable"
              ? "unparseable"
              : "equality",
          left: left,
          right: right,
          value: comparisons[comp](left.value, right.value),
        };
      }
    }

    if (this._data.pathExists(expression)) {
      return {
        type: "path",
        path: expression,
        value: this._data.retrieve(expression),
      };
    }

    try {
      return {
        type: "value",
        value: toValue(expression),
      };
    } catch (e) {}

    return {
      type: "unparseable",
    };
  }

  /* DOM --> data object */

  _handleChange(evnt) {
    // Which element changed?
    const target = evnt.target.closest(
      `${this._scope} [${this._options.bind}], ${this._scope} [${this._options.write}]`
    );
    if (!target) return;

    // Apply the changes to the data object
    this._apply(target.getAttribute(this._options.bind), target);
    this._apply(target.getAttribute(this._options.write), target);

    if (this._options.stopEvents) {
      // We're done with this event, don't try to evaluate it any further
      evnt.preventDefault();
      evnt.stopPropagation();
    }
  }

  _apply(attribute, target) {
    if (!attribute) return;
    try {
      const expression = this._parseExpression(attribute);
      if (!expression.type == "path") {
        return console.error("Expected path");
      }

      let value;
      for (const conv of this._options.customConversions) {
        if (target.matches(conv.selector) && conv.fromDom) {
          value = conv.fromDom(expression.path, target);
          this._data.store(expression.path, value);
          return;
        }
      }

      if (target.matches("input[type=checkbox")) {
        value = target.checked;
      } else if ("value" in target) {
        value = target.value;
      } else throw new Error("no value to write to data object");

      this._data.store(expression.path, value);
    } catch (e) {
      console.error(e);
    }
  }

  _handleClick(evnt) {
    // Which element was clicked?
    const target = evnt.target.closest(
      `${this._scope} [${this._options.click}]`
    );
    if (!target) return;

    // Apply the changes to the data object
    const attribute = target.getAttribute(this._options.click);
    const expression = this._parseExpression(attribute);

    if (!expression.type == "assignment")
      return console.error(
        `Expected assignment in ${this._options.click}, got ${attribute}`
      );

    try {
      this._data.store(expression.assignee, expression.value);
    } catch (e) {
      console.error(e);
    }

    if (this._options.stopEvents) {
      // We're done with this event, don't try to evaluate it any further
      evnt.preventDefault();
      evnt.stopPropagation();
    }
  }

  /* Data object --> DOM */

  _applyValue(path) {
    // Lists need to be done first, so elements in the lists get updates too.
    // For the rest the order doesn't matter.
    this._applyLists(path);
    this._applyReads(path);
    this._applyBinds(path);
    this._applyClasses(path);
  }

  _applyReads(path) {
    // Update values in `read` elements
    document
      .querySelectorAll(`${this._scope} [${this._options.read}]`)
      .forEach((e) => {
        const readExpression = e.getAttribute(this._options.read);
        if (path != null && !readExpression.includes(path)) return;
        const expression = this._parseExpression(readExpression);

        if (expression.type == "unparseable")
          return console.error("Could not parse expression: " + readExpression);

        for (const conv of this._options.customConversions) {
          if (e.matches(conv.selector) && conv.toDom) {
            conv.toDom(path, expression.value, e);
            return;
          }
        }

        if (e.matches("input[type=checkbox")) {
          e.checked = expression.value;
        } else if ("value" in e) {
          e.value = expression.value;
        } else if ("innerText" in e) {
          e.innerText = expression.value;
        } else throw new Error("don't know how to write value to DOM element");
      });
  }

  _applyBinds(path) {
    // Update values in `bind` elements
    document
      .querySelectorAll(`${this._scope} [${this._options.bind}]`)
      .forEach((e) => {
        const bindExpression = e.getAttribute(this._options.bind);
        if (path != null && !bindExpression.includes(path)) return;
        const expression = this._parseExpression(bindExpression);

        if (expression.type == "unparseable")
          return console.error("Could not parse expression: " + bindExpression);

        for (const conv of this._options.customConversions) {
          if (e.matches(conv.selector) && conv.toDom) {
            conv.toDom(path, expression.value, e);
            return;
          }
        }

        if (e.matches("input[type=checkbox")) {
          e.checked = expression.value;
        } else if ("value" in e) {
          e.value = expression.value;
        } else throw new Error("don't know how to write value to DOM element");
      });
  }

  _applyClasses(path) {
    // Update classes (for active-if expressions)
    document
      .querySelectorAll(`${this._scope} [${this._options.activeIf}]`)
      .forEach((e) => {
        const activeIfExpression = e.getAttribute(this._options.activeIf);
        if (path != null && !activeIfExpression.includes(path)) return;
        const expression = this._parseExpression(activeIfExpression);

        if (expression.type == "unparseable")
          return console.error(
            "Could not parse expression: " + activeIfExpression
          );

        e.classList.toggle("active", !!expression.value);
      });
  }

  _applyLists(path) {
    // Update loops to have all the right elements in them
    document
      .querySelectorAll(`${this._scope} [${this._options.loop}]`)
      .forEach((e) => {
        const loopExpression = e.getAttribute(this._options.loop);
        if (path != null && !loopExpression.includes(path)) return;
        const expression = this._parseExpression(loopExpression);

        if (
          !(
            expression.type == "assignment" &&
            expression.assigned.type == "path"
          )
        )
          return console.error(
            `Expected assignment of path in ${this._options.loop} in ${attribute}`
          );

        if (!Array.isArray(expression.value))
          return console.error(
            `Invalid expression in ${this._options.loop}: Can't iterate over ${expression.assigned.path}`
          );

        // If this is the first time we see this, save original DOM contents as
        // a template for adding child elements
        if (!e.originalContents) {
          e.originalContents = e.innerHTML;
          e._listLength = 0;
        }

        const variable = expression.assignee;
        const list = this._data.retrieve(expression.assigned.path);
        if (list.length == e._listLength) return;

        // Rebuild sub-tree
        e.innerHTML = "";
        for (let i = 0; i < list.length; i++) {
          e.appendChild(
            this._renderListTemplate(e.originalContents, i, variable)
          );
        }
        e._listLength = list.length;
      });
  }

  _renderListTemplate(html, index, variable) {
    // Create safe regexp to find variable name
    variable = variable.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
    const re = new RegExp(`(?<!\\w)${variable}(?!\\w)`, "g");

    // Put the original HTML in a template element
    const template = document.createElement("template");
    template.innerHTML = html;

    // Modify the template's relevant attributes so we point to `index` instead
    // of `variable`.
    const attributes = [
      this._options.read,
      this._options.write,
      this._options.bind,
      this._options.click,
      this._options.activeIf,
    ];
    template.content
      .querySelectorAll(attributes.map((a) => `[${a}]`).join(","))
      .forEach((e) => {
        for (let i = 0; i < e.attributes.length; i++) {
          if (attributes.includes(e.attributes[i].name)) {
            e.attributes[i].value = e.attributes[i].value.replace(re, index);
          }
        }
      });

    return template.content;
  }
}
