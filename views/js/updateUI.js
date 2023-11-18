async function updateContent() {
  const tasksContainer = document.querySelector("#tasks");

  let bodyScroll = document.body.scrollTop;
  let queueScroll = document.getElementById("queue").scrollTop;
  let activeScroll = document.getElementById("active").scrollTop;
  let endedScroll = document.getElementById("ended").scrollTop;

  const res = await fetch("http://127.0.0.1:42069/renderTasks");
  const newContent = await res.text();
  const parsed = new DOMParser().parseFromString(newContent, "text/html");

  tasksContainer.innerHTML = newContent;
  htmx.process(tasksContainer);

  document.body.scrollTop += bodyScroll;
  document.getElementById("queue").scrollTop += queueScroll;
  document.getElementById("active").scrollTop += activeScroll;
  document.getElementById("ended").scrollTop += endedScroll;
}

setInterval(updateContent, 1000 * 2);
