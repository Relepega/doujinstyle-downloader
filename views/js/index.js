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

setInterval(updateContent, 1000 * updateInterval);
