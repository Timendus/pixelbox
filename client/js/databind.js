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
