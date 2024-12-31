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
 * @param {string[]} ids
 * @param {string} mode
 *
 */
async function taskAction(method, ids, mode) {
    let data = new FormData()
    data.append("IDs", ids)
    data.append("Mode", mode)

	const res = await fetch('/api/task', { method: method, body: data })

	if (!res.ok) {
		const text = await res.text()
		window.alert(text)
	}
}

function aggregateNodeIDFromEnded() {
    let ids = ''

    const nodes = document.querySelector('#ended').childNodes

    nodes.forEach((node, idx) => {
        if (idx == 0 || idx == nodes.length - 1) return

        if (ids === '') {
            ids = node.id
        } else {
            ids += '|' + node.id
        }
    })

    return ids
}

document.addEventListener('click', async function (evt) {
	// console.log(evt)
	switch (evt.target.id) {
		case 'clear-queued': {
			await taskAction('DELETE', '', 'queued')
			break
		}

		case 'clear-all-completed': {
			await taskAction('DELETE', '', 'completed')
			break
		}

		case 'clear-success-completed': {
			await taskAction('DELETE', '', 'succeeded')
			break
		}

		case 'clear-fail-completed': {
			await taskAction('DELETE', '', 'failed')
			break
		}

		case 'retry-fail-completed': {
			await taskAction('DELETE', '', 'retry-fail-completed')
			break
		}

		case 'task-ctrl-remove-task': {
			const taskID = evt.target.attributes['data-id'].value
			await taskAction('DELETE', taskID, 'single')
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

		form.Slugs.value = ''
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
	 //console.log('new task', event)
	// https://developer.mozilla.org/en-US/docs/Web/API/Element/insertAdjacentHTML
	document
		.getElementById('queued')
		.insertAdjacentHTML('beforeend', event.data)
})

source.addEventListener('remove-node', function (event) {
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
	//console.log('replace-node parsed data: ', data)

	const node = document.getElementById(data.TargetNodeID)
	if (node) {
		node.remove()
	}

	document
		.querySelector(data.ReceiverNodeSelector)
		.insertAdjacentHTML(data.Position, data.NewContent)
})

source.addEventListener('update-node-content', function (event) {
	const data = JSON.parse(event.data)
	//console.log('replace-node-content parsed data: ', data)

	document.getElementById(data.ReceiverNodeSelector).innerHTML = data.NewContent
})

source.addEventListener('error', async function (event) {
	if (event.data == undefined) {
		return
	}

	console.error(event.data)
	window.alert(event.data)
})
