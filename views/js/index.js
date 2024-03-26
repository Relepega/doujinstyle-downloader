const updateInterval = await fetch(document.location.origin + "/updateInterval")
    .then((res) => res.text())
    .then((res) => parseFloat(res));

function removeElement(albumID) {
    document.getElementById(albumID).remove();
}

async function copyErrorMessage(albumID) {
    const el = document.getElementById(albumID + "-error");
    await navigator.clipboard.writeText(el.innerText);
    window.alert("Error log of album " + albumID + " copied");
}

// TODO: make it less react-like
async function updateContent() {
    const tasksContainer = document.querySelector("#tasks");

    let bodyScroll = document.body.scrollTop;
    let queueScroll = document.getElementById("queue").scrollTop;
    let activeScroll = document.getElementById("active").scrollTop;
    let endedScroll = document.getElementById("ended").scrollTop;

    const res = await fetch(document.location.origin + "/api/task/render");
    const newContent = await res.text();

    tasksContainer.innerHTML = newContent;
    htmx.process(tasksContainer);

    document.body.scrollTop += bodyScroll;
    document.getElementById("queue").scrollTop += queueScroll;
    document.getElementById("active").scrollTop += activeScroll;
    document.getElementById("ended").scrollTop += endedScroll;
}

// setInterval(updateContent, 1000 * updateInterval);


// partial renderer
const sseSource = new EventSource(document.location.origin + "/event-stream")

/** 
 *
 * @typedef {Object} SSEEvent
 * @property {string} target - ID of node that will be replaced
 * @property {string} receiver - ID of node that will receive the new node
 * @property {string} node - the newly rendered node
 *
 * @param {SSEEvent} evt
 */
sseSource.addEventListener("update", (evt) => {
    const {receiver, target, node} = JSON.parse(evt.data)

    document.getElementById(target).remove()
    document.getElementById(receiver).appendChild(node)

    htmx.process(tasksContainer)
})
