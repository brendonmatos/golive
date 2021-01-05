package golive
var BasePageString = `<!DOCTYPE html>
<html lang="{{ .Lang }}">
  <head>
    <meta charset="UTF-8" />
    <title>{{ .Title }}</title>
    {{ .Head }}
  </head>
  <script type="application/javascript">
    const GO_LIVE_CONNECTED = "go-live-connected";
    const GO_LIVE_COMPONENT_ID = "go-live-component-id";

    const findLiveInputsFromElement = (el) => {
      return el.querySelectorAll(
        ['*[go-live-input]:not([', GO_LIVE_CONNECTED, '])'].join('')
      );
    };

    const findLiveClicksFromElement = (el) => {
      return el.querySelectorAll(
        ['*[go-live-click]:not([', GO_LIVE_CONNECTED, '])'].join('')
      );
    };

    function getElementChild(element, index) {
      let el = element.firstChild;

      if (el === Node.TEXT_NODE) {
        throw new Error("Element is a text node, without children");
      }

      while (index > 0) {
        if (!el) {
          console.log("Element not found in path", element);
          return;
        }

        el = el.nextSibling;
        index--;
      }

      return el;
    }

    function isElement(o) {
      return typeof HTMLElement === "object"
        ? o instanceof HTMLElement //DOM2
        : o &&
            typeof o === "object" &&
            o.nodeType === 1 &&
            typeof o.nodeName === "string";
    }

    function handleDiffSetAttr(message, el) {
      const { attr } = message;

      if (attr.Name === "value" && el.value) {
        el.value = attr.Value;
      } else {
        el.setAttribute(attr.Name, attr.Value);
      }
    }

    function handleDiffRemoveAttr(message, el) {
      const { attr } = message;

      el.removeAttribute(attr.Name);
    }
    
    function handleDiffReplace(message, el) {
      const { content } = message;

      const wrapper = document.createElement("div");
      wrapper.innerHTML = content;

      el.parentElement.replaceChild(wrapper.firstChild, el);
    }
    
    function handleDiffRemove(message, el) {
      el.parentElement.removeChild(el);
    }
    
    function handleDiffSetInnerHTML(message, el, componentId) {
      const { content } = message;

      if (el.nodeType === Node.TEXT_NODE) {
        el.textContent = content;
        return;
      }

      el.innerHTML = content;

      goLive.connectElement(componentId, el);
    }
    
    function handleDiffAppend(message, el, componentId) {
      const { content } = message;

      const wrapper = document.createElement("div");
      wrapper.innerHTML = content;
      const child = wrapper.firstChild;
      el.appendChild(child);

      goLive.connectElement(componentId, el);
    }

    const handleChange = {
      "{{ .Enum.DiffSetAttr }}": handleDiffSetAttr,
      "{{ .Enum.DiffRemoveAttr }}": handleDiffRemoveAttr,
      "{{ .Enum.DiffReplace }}": handleDiffReplace,
      "{{ .Enum.DiffRemove }}": handleDiffRemove,
      "{{ .Enum.DiffSetInnerHtml }}": handleDiffSetInnerHTML,
      "{{ .Enum.DiffAppend }}": handleDiffAppend,
    };

    const goLive = {
      server: new WebSocket(["ws://", window.location.host, "/ws"].join("")),

      handlers: [],
      onceHandlers: {},

      getLiveComponent(id) {
        return document.querySelector(
          ["*[",GO_LIVE_COMPONENT_ID, "=", id, "]"].join("")
        );
      },

      on(name, handler) {
        const newSize = this.handlers.push({
          name,
          handler,
        });
        return newSize - 1;
      },

      emitOnce(name) {
        const handler = this.onceHandlers[name];
        if (!handler) {
          this.createOnceHandler(name, true);
          return;
        }
        for (const cb of handler.cbs) {
          cb();
        }
      },

      createOnceHandler(name, called) {
        this.onceHandlers[name] = {
          called,
          cbs: [],
        };

        return this.onceHandlers[name];
      },

      once(name, cb) {
        let handler = this.onceHandlers[name];

        if (!handler) {
          handler = this.createOnceHandler(name, false);
        }

        handler.cbs.push(cb);
      },

      findHandler(name) {
        return this.handlers.filter((i) => i.name === name);
      },

      emit(name, message) {
        for (const handler of this.findHandler(name)) {
          handler.handler(message);
        }
      },

      off(index) {
        this.handlers.splice(index, 1);
      },

      send(message) {
        goLive.server.send(
            JSON.stringify(message))
      },

      connectChilds(viewElement) {
        const liveChilds = viewElement.querySelectorAll(
          "*[" + GO_LIVE_COMPONENT_ID + "]"
        );

        liveChilds.forEach((child) => {
          const componentId = child.getAttribute(GO_LIVE_COMPONENT_ID);
          this.connectElement(componentId, child);
        });
      },

      connectElement(componentId, viewElement) {
        if (typeof viewElement === "string") {
          return;
        }

        if (!isElement(viewElement)) {
          return;
        }

        const liveInputs = findLiveInputsFromElement(viewElement);
        const clickElements = findLiveClicksFromElement(viewElement);

        clickElements.forEach(function (element) {
          element.addEventListener("click", function (_) {
            goLive.send({
                name: "{{ .Enum.EventLiveMethod }}",
                component_id: componentId,
                method_name: element.getAttribute("go-live-click"),
                value: String(element.value),
            })
          });
          element.setAttribute(GO_LIVE_CONNECTED, true);
        });

        liveInputs.forEach(function (element) {
          const type = element.getAttribute("type");

          element.addEventListener("input", function (_) {
            let value = element.value;

            if (type === "checkbox") {
              value = element.checked;
            }

            goLive.send({
                name: "{{ .Enum.EventLiveInput }}",
                component_id: componentId,
                key: element.getAttribute("go-live-input"),
                value: String(value),
            })
        
          });
          element.setAttribute(GO_LIVE_CONNECTED, true);
        });
      },

      connect(id) {
        const element = goLive.getLiveComponent(id);

        goLive.connectElement(id, element);

        goLive.on("{{ .Enum.EventLiveDom }}", function handleLiveDom(message) {
          if (id === message.cid) {
            for (const instruction of message.i) {
              const type = instruction.t;
              const content = instruction.c;
              const attr = instruction.a;
              const selector = instruction.s;

              const element = document.querySelector(selector)

              handleChange[type](
                {
                  content: content,
                  attr: attr,
                },
                element,
                id
              );
            }
          }
        });
      },
    };

    goLive.server.onmessage = (rawMessage) => {
      try {
        const message = JSON.parse(rawMessage.data);
        goLive.emit(message.n, message);
      } catch (e) {
        console.log("Error", e);
        console.log("Error message", rawMessage.data);
      }
    };

    goLive.server.onopen = () => {
      goLive.emitOnce("WS_CONNECTION_OPEN");
    };
  </script>

  <body>
    {{ .Body }}
  </body>
</html>
`