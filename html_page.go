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
		const goLive = {
			server: new WebSocket(['ws://', window.location.host, "/ws"].join("")),

			handlers: [],
			onceHandlers: {},

			getLiveComponent(id) {
				return document.querySelector(['*[go-live-component-id=', id, ']'].join(''));
			},

			on(name, handler) {
				const newSize = this.handlers.push({
					name,
					handler
				});
				return newSize - 1;
			},

			emitOnce(name) {
				const handler = this.onceHandlers[name]
				if( !handler ) {
					this.createOnceHandler(name, true);
					return;
				}
				for( const cb of handler.cbs ){
					cb();
				}
			},

			createOnceHandler(name, called) {
				this.onceHandlers[name] = {
					called,
					cbs: []
				};

				return this.onceHandlers[name];
			},

			once(name, cb) {
				let handler = this.onceHandlers[name];

				if( !handler ) {
					handler = this.createOnceHandler(name, false);
				}

				handler.cbs.push(cb);
			},

			findHandler(name) {
				return this.handlers.filter( i => i.name === name );
			},

			emit(name, message) {
				for (const handler of this.findHandler(name)) {
					handler.handler(message);
				}
			},

			off(index) {
				this.handlers.splice(index, 1);
			},

			connectElement(scopeId, viewElement) {

				const clickElements = findLiveClicksFromElement(viewElement);
				clickElements.forEach(function(element) {
					if (!element) {
						return;
					}

					element.addEventListener('click', function(_) {
						goLive.server.send(JSON.stringify({
							name: "{{ .Enum.EventLiveMethod }}",
							component_id: scopeId,
							method_name: element.getAttribute("go-live-click"),
							method_data: dataFromElementAttributes(element),
						}));
					});
				});

				const keydownElements = findLiveKeyDownFromElement(viewElement);
				keydownElements.forEach(function(element) {
					if (!element) {
						return;
					}

					const method = element.getAttribute("go-live-keydown");

					const attrs = element.attributes;
					let filterKeys = [];
					for(let i = 0; i < attrs.length; i++) {
						if (attrs[i].name === "go-live-key" || attrs[i].name.startsWith("go-live-key-")) {
							filterKeys.push(attrs[i].value);
						}
					}

					element.addEventListener('keydown', function(event) {
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
							goLive.server.send(JSON.stringify({
								name: "{{ .Enum.EventLiveMethod }}",
								component_id: scopeId,
								method_name: method,
								method_data: dataFromElementAttributes(element),
								dom_event: {
									keyCode: code,
								},
							}));
						}
					});
				});

				const liveInputs = findLiveInputsFromElement(viewElement);
				liveInputs.forEach(function(element) {

					if (!element) {
						return;
					}

					const type = element.getAttribute("type");

					element.addEventListener('input', function(_) {

						let value = element.value;

						if (type === "checkbox") {
							value = element.checked;
						}

						goLive.server.send(JSON.stringify({
							name: "{{ .Enum.EventLiveInput }}",
							component_id: scopeId,
							key: element.getAttribute("go-live-input"),
							value: String(value)
						}));
					});
				});

				goLive.on('{{ .Enum.EventLiveDom }}', (message) => {

					const handleChange = {
						'{{ .Enum.DiffSetAttr }}': (message) => {
							const {
								attr,
								element
							} = message;

							const el = viewElement.querySelector(element);

							if( !el ) {
								console.error("Element not found", element);
								return;
							}

							if (attr.Name === "value" && el.value) {
								el.value = attr.Value;
							}

							else {
								el.setAttribute(attr.Name, attr.Value);
							}

						},
						'{{ .Enum.DiffRemoveAttr }}': (message) => {
							const {
								attr,
								element
							} = message;

							const el = viewElement.querySelector(element);

							if( !el ) {
								console.error("Element not found", element);
								return;
							}

							el.removeAttribute(attr.Name);

						},
						'{{ .Enum.DiffReplace }}': (message) => {
							const {
								content,
								element
							} = message;

							const el = viewElement.querySelector(element);

							if (!el) {
								console.warn("Element not found with selector", element)
								return
							}

							const wrapper = document.createElement('div');
							wrapper.innerHTML = content;

							el.parentElement.replaceChild(wrapper.firstChild, el)
						},
						'{{ .Enum.DiffRemove }}': (message) => {
							const {
								element
							} = message

							const el = viewElement.querySelector(element)

							if (!el) {
								console.warn("Element not found with selector", element)
								return
							}

							el.parentElement.removeChild(el)
						},
						'{{ .Enum.DiffSetInnerHtml }}': (message) => {
							const {
								content,
								element
							} = message;

							const el = viewElement.querySelector(element);

							if (!el) {
								console.warn("Element not found with selector", element);
								return;
							}

							if (content.trim().length === 0) {
								return;
							}

							el.innerHTML = content;

							// Add listeners to new elements
							goLive.connectElement(scopeId, el);

						},
						'{{ .Enum.DiffAppend }}': (message) => {
							const {
								content,
								element
							} = message;

							const el = viewElement.querySelector(element);

							if (!el) {
								console.warn("Element not found with selector", element);
								return;
							}

							const wrapper = document.createElement('div');
							wrapper.innerHTML = content;

							if (content.trim().length > 0) {
								el.appendChild(wrapper.firstChild);
							}

							goLive.connectElement(scopeId, el.lastChild);
						}
					}

					if (viewElement.getAttribute("go-live-component-id") === message.component_id) {
						handleChange[message.type.toLowerCase()](message);
					}
				})
			},

			connect(id) {
				goLive.connectElement(id, goLive.getLiveComponent(id));
			}
		}

		goLive.server.onmessage = (rawMessage) => {
			const message = JSON.parse(rawMessage.data);

			goLive.emit(message.name, message);
		}

		goLive.server.onopen = () => {
			goLive.emitOnce('WS_CONNECTION_OPEN');
		}

		const findLiveInputsFromElement = (el) => {
			return el.querySelectorAll('*[go-live-input]');
		}

		const findLiveClicksFromElement = (el) => {
			return el.querySelectorAll('*[go-live-click]');
		}

		const findLiveKeyDownFromElement = (el) => {
			return el.querySelectorAll('*[go-live-keydown]');
		}

		const dataFromElementAttributes = (el) => {
			const attrs = el.attributes;
			let data = {};
			for(let i = 0; i < attrs.length; i++) {
				if (attrs[i].name.startsWith("go-live-data-")) {
					data[attrs[i].name.substring(13)] = attrs[i].value
				}
			}

			return data
		}
	</script>
	<body>
	{{ .Body }}
	</body>
	</html>
`
