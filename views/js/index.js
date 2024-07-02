const serviceSelect = document.querySelector('#ServiceNumber')

document.addEventListener('DOMContentLoaded', () => {
	const value = localStorage.getItem('LastSelectedService')

	if (value) {
		serviceSelect.value = value
	}
})

serviceSelect.addEventListener('change', () => {
	localStorage.setItem('LastSelectedService', serviceSelect.value)
})

/**
 *
 * @param {string} method
 * @param {string} albumID
 * @param {string} groupAction
 *
 */
async function taskAction(method, albumID, groupAction) {
	const res = await fetch('/api/task', {
		method: method,
		body: JSON.stringify({
			AlbumID: albumID,
			GroupAction: groupAction,
		}),
	})

	if (!res.ok) {
		const text = await res.text()
		window.alert(text)
	}
}

document.addEventListener('click', async function (evt) {
	// console.log(evt)
	switch (evt.target.id) {
		case 'clear-queued': {
			await taskAction('DELETE', '', 'clear-queued')
			break
		}

		case 'clear-all-completed': {
			taskAction('DELETE', '', 'clear-all-completed')
			break
		}

		case 'clear-success-completed': {
			await taskAction('DELETE', '', 'clear-success-completed')
			break
		}

		case 'clear-fail-completed': {
			await taskAction('DELETE', '', 'clear-fail-completed')
			break
		}

		case 'retry-fail-completed': {
			await taskAction('DELETE', '', 'retry-fail-completed')
			break
		}

		case 'task-ctrl-remove-task': {
			const albumID = evt.target.attributes['data-id'].value
			await taskAction('DELETE', albumID, '')
			break
		}

		case 'task-ctrl-copy-error': {
			const albumID = evt.target.attributes['data-id'].value
			const el = document.getElementById(albumID + '-error')

			await navigator.clipboard.writeText(el.innerText)
			window.alert('Error log of album ' + albumID + ' copied')

			break
		}

		case 'task-ctrl-retry': {
			const albumID = evt.target.attributes['data-id'].value

			const formData = new FormData()
			formData.append('AlbumID', albumID)

			await fetch('/api/task', { method: 'PATCH', body: formData }).then(
				async (res) => {
					if (!res.ok) {
						const text = await res.text()
						window.alert(text)
					}
				},
			)

			break
		}

		default:
			break
	}
})

// document.querySelector("#clear-queued-btn").addEventListener("click", function() {
//     await taskAction("DELETE", "", "clear")
//     console.log("button pressed")
// })

document
	.querySelector('form > button')
	.addEventListener('click', async function (e) {
		e.preventDefault()

		const form = document.querySelector('form')

		const formData = new FormData(form)
		await fetch('/api/task', { method: 'POST', body: formData })

		form.AlbumID.value = ''
	})

document
	.querySelector('#restart-btn')
	.addEventListener('click', async function (e) {
		const res = window.confirm(
			'WARNING: this operation will restart the application. All unsaved progress will be discarded.\n\n Continue?',
		)

		if (res) {
			try {
				await fetch('/api/internal/restart', {
					method: 'POST',
				})
			} catch (error) {}

			window.location.reload()
		}
	})

// SSE things
const source = new EventSource(window.location.origin + '/events-stream')

source.addEventListener('message', function (event) {
	console.log('new message from server: ', event.data)
	// let node = document.createElement("p")
	// node.innerHTML = event.data
	// document.getElementById("content").prepend(node)
})

source.addEventListener('new-task', function (event) {
	// console.log('new task', event)
	// https://developer.mozilla.org/en-US/docs/Web/API/Element/insertAdjacentHTML
	document
		.getElementById('queued')
		.insertAdjacentHTML('beforeend', event.data)
})

source.addEventListener('remove-task', function (event) {
	// console.log('to remove: ', event.data)
	const node = document.getElementById(event.data)

	if (!node) {
		return
	}

	// document.getElementById("content").removeChild(node)
	node.remove()
})

source.addEventListener('replace-node', function (event) {
	const data = JSON.parse(event.data)
	// console.log('replace-node parsed data: ', data)

	const node = document.getElementById(data.targetNodeID)
	if (node) {
		node.remove()
	}

	document
		.querySelector(data.receiverNode)
		.insertAdjacentHTML(data.position, data.newContent)
})

source.addEventListener('replace-node-content', function (event) {
	const data = JSON.parse(event.data)
	// console.log('replace-node-content parsed data: ', data)

	document.getElementById(data.receiverNode).innerHTML = data.newContent
})

source.addEventListener('error', async function (event) {
	if (event.data == undefined) {
		return
	}

	console.error(event.data)
	window.alert(event.data)
})
