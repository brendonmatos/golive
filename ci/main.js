const GO_LIVE_CONNECTED = "go-live-connected";
const GO_LIVE_COMPONENT_ID = "go-live-component-id";
const EVENT_LIVE_DOM_COMPONENT_ID_KEY = "cid";
const EVENT_LIVE_DOM_INSTRUCTIONS_KEY = "i";
const EVENT_LIVE_DOM_TYPE_KEY = "t";
const EVENT_LIVE_DOM_CONTENT_KEY = "c";
const EVENT_LIVE_DOM_ATTR_KEY = "a";
const EVENT_LIVE_DOM_SELECTOR_KEY = "s";
const EVENT_LIVE_DOM_INDEX_KEY = "i";

const handleChange = {
    "{{ .Enum.DiffSetAttr }}": handleDiffSetAttr,
    "{{ .Enum.DiffRemoveAttr }}": handleDiffRemoveAttr,
    "{{ .Enum.DiffReplace }}": handleDiffReplace,
    "{{ .Enum.DiffRemove }}": handleDiffRemove,
    "{{ .Enum.DiffSetInnerHTML }}": handleDiffSetInnerHTML,
    "{{ .Enum.DiffAppend }}": handleDiffAppend,
    "{{ .Enum.DiffMove }}": handleDiffMove,
};

const goLive = {
    server: createConnection(),

    handlers: [],

    once: createOnceEmitter(),

    getLiveComponent(id) {
        return document.querySelector(
            ["*[", GO_LIVE_COMPONENT_ID, "=", id, "]"].join("")
        );
    },

    on(name, handler) {
        const newSize = this.handlers.push({
            name,
            handler,
        })
        return newSize - 1;
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
        goLive.server.send(JSON.stringify(message));
    },

    connectChildren(viewElement) {
        const liveChildren = viewElement.querySelectorAll(
            "*[" + GO_LIVE_COMPONENT_ID + "]"
        );

        liveChildren.forEach((child) => {
            this.connectElement(child);
        });
    },

    connectElement(viewElement) {
        if (typeof viewElement === "string") {
            console.warn("is string")
            return;
        }

        if (!isElement(viewElement)) {
            console.warn("not element")
            return;
        }

        const connectedElements = []

        const clickElements = findLiveClicksFromElement(viewElement);
        clickElements.forEach(function (element) {

            const componentId = getComponentIdFromElement(element);

            element.addEventListener("click", function (_) {
                goLive.send({
                    name: "{{ .Enum.EventLiveMethod }}",
                    component_id: componentId,
                    method_name: element.getAttribute("go-live-click"),
                    method_data: dataFromElementAttributes(element),
                });
            });

            connectedElements.push(element)
        });

        const keydownElements = findLiveKeyDownFromElement(viewElement);
        keydownElements.forEach(function (element) {

            const componentId = getComponentIdFromElement(element);
            const method = element.getAttribute("go-live-keydown");

            const attrs = element.attributes;
            let filterKeys = [];
            for (let i = 0; i < attrs.length; i++) {
                if (
                    attrs[i].name === "go-live-key" ||
                    attrs[i].name.startsWith("go-live-key-")
                ) {
                    filterKeys.push(attrs[i].value);
                }
            }

            element.addEventListener("keydown", function (event) {
                const code = String(event.code);
                let hit = true;

                if (filterKeys.length !== 0) {
                    hit = false;
                    for (let i = 0; i < filterKeys.length; i++) {
                        if (filterKeys[i] === code) {
                            hit = true;

                            break;
                        }
                    }
                }

                if (hit) {
                    goLive.send({
                        name: "{{ .Enum.EventLiveMethod }}",
                        component_id: componentId,
                        method_name: method,
                        method_data: dataFromElementAttributes(element),
                        dom_event: {
                            keyCode: code,
                        },
                    });
                }
            });

            connectedElements.push(element)
        });

        const liveInputs = findLiveInputsFromElement(viewElement);
        liveInputs.forEach(function (element) {

            const type = element.getAttribute("type");
            const componentId = getComponentIdFromElement(element);

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
                });
            });

            connectedElements.push(element)
        });


        for( const el of connectedElements ) {
            el.setAttribute(GO_LIVE_CONNECTED, true);
        }
    },

    connect(id) {
        const element = goLive.getLiveComponent(id);

        goLive.connectElement(element);

        goLive.on(
            "{{ .Enum.EventLiveDom }}",
            function handleLiveDom(message) {
                if (id === message[EVENT_LIVE_DOM_COMPONENT_ID_KEY]) {
                    for (const instruction of message[
                        EVENT_LIVE_DOM_INSTRUCTIONS_KEY
                        ]) {
                        const type = instruction[EVENT_LIVE_DOM_TYPE_KEY];
                        const content = instruction[EVENT_LIVE_DOM_CONTENT_KEY];
                        const attr = instruction[EVENT_LIVE_DOM_ATTR_KEY];
                        const selector = instruction[EVENT_LIVE_DOM_SELECTOR_KEY];
                        const index = instruction[EVENT_LIVE_DOM_INDEX_KEY]

                        const element = document.querySelector(selector);

                        if (!element) {
                            console.error("Element not found", selector);
                            return;
                        }

                        handleChange[type](
                            {
                                content: content,
                                attr: attr,
                                index: index
                            },
                            element,
                            id
                        );
                    }
                }
            }
        );
    },
};

goLive.once.on("WS_CONNECTION_OPEN", () => {
    goLive.on("{{ .Enum.EventLiveConnectElement }}", (message) => {
        const cid = message[EVENT_LIVE_DOM_COMPONENT_ID_KEY];
        goLive.connect(cid);
    });
    goLive.on("{{ .Enum.EventLiveError }}", (message) => {
        console.error("message", message.m)
        if (
            message.m ===
            '{{ index .EnumLiveError ` + "`LiveErrorSessionNotFound`" + `}}'
        ) {
            window.location.reload(false);
        }
    });
});

goLive.server.onmessage = (rawMessage) => {
    try {
        const message = JSON.parse(rawMessage.data);
        goLive.emit(message.t, message);
    } catch (e) {
        console.log("Error", e);
        console.log("Error message", rawMessage.data);
    }
};

goLive.server.onopen = () => {
    goLive.once.emit("WS_CONNECTION_OPEN");
};

function createConnection() {
    const path = [];

    if (window.location.protocol === "https:") {
        path.push("wss");
    } else {
        path.push("ws");
    }

    path.push("://", window.location.host, "/ws");

    return new WebSocket(path.join(""));
}

function createOnceEmitter() {
    const handlers = {};
    const createHandler = (name, called) => {
        handlers[name] = {
            called,
            cbs: [],
        };

        return handlers[name];
    };

    return {
        on(name, cb) {
            let handler = handlers[name];

            if (!handler) {
                handler = createHandler(name, false);
            }

            handler.cbs.push(cb);
        },
        emit(name, ...attrs) {
            const handler = handlers[name];

            if (!handler) {
                createHandler(name, true);
                return;
            }

            for (const cb of handler.cbs) {
                cb();
            }
        },
    };
}

const findLiveInputsFromElement = (el) => {
    return el.querySelectorAll(
        ["*[go-live-input]:not([", GO_LIVE_CONNECTED, "])"].join("")
    );
};

const findLiveClicksFromElement = (el) => {
    return el.querySelectorAll(
        ["*[go-live-click]:not([", GO_LIVE_CONNECTED, "])"].join("")
    );
};

const findLiveKeyDownFromElement = (el) => {
    return el.querySelectorAll(
        ["*[go-live-keydown]:not([", GO_LIVE_CONNECTED, "])"].join("")
    );
};

const dataFromElementAttributes = (el) => {
    const attrs = el.attributes;
    let data = {};
    for (let i = 0; i < attrs.length; i++) {
        if (attrs[i].name.startsWith("go-live-data-")) {
            data[attrs[i].name.substring(13)] = attrs[i].value;
        }
    }

    return data;
};

function getElementChild(element, index) {
    let el = element.firstElementChild;

    while (index > 0) {
        if (!el) {
            console.error("Element not found in path", element);
            return;
        }

        el = el.nextSibling;

        if (el.nodeType !== Node.ELEMENT_NODE) {
            continue
        }

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

    const parent = el.parentElement
    parent.replaceChild(wrapper.firstChild, el);

    goLive.connectElement(parent)
}

function handleDiffRemove(message, el) {
    const parent = el.parentElement
    parent.removeChild(el);
}

function handleDiffSetInnerHTML(message, el) {
    let { content } = message;

    if (content === undefined) {
        content = "";
    }

    if (el.nodeType === Node.TEXT_NODE) {
        el.textContent = content;
        return;
    }

    el.innerHTML = content;

    goLive.connectElement(el);
}

function handleDiffAppend(message, el) {
    const { content } = message;

    const wrapper = document.createElement("div");
    wrapper.innerHTML = content;

    const child = wrapper.firstChild;
    el.appendChild(child);
    goLive.connectElement(el);
}

function handleDiffMove(message, el) {
    const parent = el.parentNode
    parent.removeChild(el)

    const child = getElementChild(parent, message.index)
    parent.replaceChild(el, child)
}
const getComponentIdFromElement = (element) => {
    const attr = element.getAttribute("go-live-component-id");
    if (attr) {
        return attr;
    }

    if (element.parentElement) {
        return getComponentIdFromElement(element.parentElement);
    }

    return undefined;
};
