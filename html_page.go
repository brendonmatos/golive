package golive

var BasePageString = `
	<!DOCTYPE html>
<html lang="{{ .Lang }}">
<head>
    <meta charset="UTF-8">
    <title>{{ .Title }}</title>
    {{ .Head }}
</head>
<script type="application/javascript">

    const findLiveInputsFromElement = (el) => {
        return el.querySelectorAll('*[go-live-input]')
    }

    const findLiveClicksFromElement = (el) => {
        return el.querySelectorAll('*[go-live-click]')
    }

    function getElementChild (element, index) {
        let el = element.firstChild
        while( index > 0 ) {
            el = el.nextSibling
            index --
        }

        return el;
    }

    function getElementByIndexPath(indexPath, viewElement) {

        let parent = viewElement
        for ( const index of indexPath.slice(1) ) {

            parent = getElementChild(parent, index)

            if (!parent) {
                console.log(viewElement, indexPath)
                throw new Error("Path not found in element")
            }
        }

        return parent;
    }

    function isElement(o){
        return (
            typeof HTMLElement === "object" ? o instanceof HTMLElement : //DOM2
                o && typeof o === "object"  && o.nodeType === 1 && typeof o.nodeName==="string"
        );
    }


    const goLive = {
        server: new WebSocket(['ws://', window.location.host, "/ws"].join("")),

        handlers: [],
        onceHandlers: {},

        getLiveComponent(id) {
            return document.querySelector(['*[go-live-component-id=', id, ']'].join('') )
        },
        
        on(name, handler) {
            const newSize = this.handlers.push({
                name,
                handler
            })
            return newSize - 1
        },

        emitOnce(name) {
            const handler = this.onceHandlers[name]
            if( !handler ) {
                this.createOnceHandler(name, true)
                return
            }
            for( const cb of handler.cbs ){
                cb()
            }
        },

        createOnceHandler(name, called) {
            this.onceHandlers[name] = {
                called,
                cbs: []
            }

            return this.onceHandlers[name]
        },

        once(name, cb) {
            let handler = this.onceHandlers[name]

            if( !handler ) {
                handler = this.createOnceHandler(name, false)
            }

            handler.cbs.push(cb)
        },

        findHandler(name) {
            return this.handlers.filter( i => i.name === name )
        },

        emit(name, message) {
            for (const handler of this.findHandler(name)) {
                handler.handler(message)
            }
        },

        off(index) {
            this.handlers.splice(index, 1)
        },

        connectElement(scopeId, viewElement) {

            if ( typeof viewElement === 'string' ) {
                return
            }

            if ( !isElement(viewElement) ) {
                return
            }

            const liveInputs = findLiveInputsFromElement(viewElement)
            const clickElements = findLiveClicksFromElement(viewElement)

            clickElements.forEach(function(element) {
                if (!element) {
                    return
                }

                element.addEventListener('click', function(_) {
                    goLive.server.send(JSON.stringify({
                        name: "{{ .Enum.EventLiveMethod }}",
                        component_id: scopeId,
                        method_name: element.getAttribute("go-live-click"),
                        value: String(element.value)
                    }))
                })
            })

            liveInputs.forEach(function(element) {

                if (!element) {
                    return
                }

                const type = element.getAttribute("type")

                element.addEventListener('input', function(_) {
                    let value = element.value

                    if (type === "checkbox") {
                        value = element.checked
                    }

                    goLive.server.send(JSON.stringify({
                        name: "{{ .Enum.EventLiveInput }}",
                        component_id: scopeId,
                        key: element.getAttribute("go-live-input"),
                        value: String(value)
                    }))
                })
            })

            goLive.on('{{ .Enum.EventLiveDom }}', (message) => {

                const handleChange = {
                    '{{ .Enum.DiffSetAttr }}': (message) => {
                        const {
                            attr,
                            path
                        } = message

                        const el = getElementByIndexPath(path, viewElement)

                        if( !el ) {
                            console.error("Path not found", path)
                            return
                        }

                        if (attr.Name === "value" && el.value) {
                            el.value = attr.Value
                        }

                        else {
                            el.setAttribute(attr.Name, attr.Value)
                        }

                    },
                    '{{ .Enum.DiffRemoveAttr }}': (message) => {
                        const {
                            attr,
                            path
                        } = message

                        const el = getElementByIndexPath(path, viewElement)

                        if( !el ) {
                            console.error("Path not found", path)
                            return
                        }

                        el.removeAttribute(attr.Name)

                    },
                    '{{ .Enum.DiffReplace }}': (message) => {
                        const {
                            content,
                            path
                        } = message

                        const el = getElementByIndexPath(path, viewElement)

                        if (!el) {
                            console.warn("Path not found with selector", path)
                            return
                        }

                        const wrapper = document.createElement('div');
                        wrapper.innerHTML = content;

                        el.parentElement.replaceChild(wrapper.firstChild, el)
                    },
                    '{{ .Enum.DiffRemove }}': (message) => {
                        const {
                            path
                        } = message

                        const el = getElementByIndexPath(path, viewElement)

                        if (!el) {
                            console.warn("Path not found with selector", path)
                            return
                        }

                        el.parentElement.removeChild(el)
                    },
                    '{{ .Enum.DiffSetInnerHtml }}': (message) => {
                        const {
                            content,
                            path
                        } = message

                        const el = getElementByIndexPath(path, viewElement)

                        if (!el) {
                            console.warn("Path not found with selector", path)
                            return
                        }

                        if (el.nodeType === Node.TEXT_NODE) {
                            el.textContent = content
                            return
                        }

                        el.innerHTML = content

                        // Add listeners to new elements
                        goLive.connectElement(scopeId, el)
                    },
                    '{{ .Enum.DiffAppend }}': (message) => {
                        const {
                            content,
                            path
                        } = message

                        const el = getElementByIndexPath(path, viewElement)

                        if (!el) {
                            console.warn("Path not found with selector", path)
                            return
                        }

                        const wrapper = document.createElement('div');
                        wrapper.innerHTML = content;

                        if (content.trim().length > 0) {
                            el.appendChild(wrapper.firstChild)
                        }

                        goLive.connectElement(scopeId, el.lastChild)
                    }
                }

                if (viewElement.getAttribute("go-live-component-id") === message.component_id) {
                    handleChange[message.type.toLowerCase()](message)
                }
            })
        },

        connect(id) {
            goLive.connectElement(id, goLive.getLiveComponent(id))
        },
    }

    goLive.server.onmessage = (rawMessage) => {
        const message = JSON.parse(rawMessage.data)
        goLive.emit(message.name, message)
    }

    goLive.server.onopen = () => {
        goLive.emitOnce('WS_CONNECTION_OPEN')
    }

</script>
<body>
{{ .Body }}
</body>
</html>
`
