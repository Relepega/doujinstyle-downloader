async function updateContent() {
  const tasksContainer = document.querySelector("#tasks");

  let bodyScroll = document.body.scrollTop;
  let queueScroll = document.getElementById("queue").scrollTop;
  let activeScroll = document.getElementById("active").scrollTop;
  let endedScroll = document.getElementById("ended").scrollTop;

  // const res = await fetch("http://127.0.0.1:5522/renderTasks");
  const res = await fetch(document.location.origin + "/renderTasks");
  const newContent = await res.text();

  tasksContainer.innerHTML = newContent;
  htmx.process(tasksContainer);

  document.body.scrollTop += bodyScroll;
  document.getElementById("queue").scrollTop += queueScroll;
  document.getElementById("active").scrollTop += activeScroll;
  document.getElementById("ended").scrollTop += endedScroll;
}

setInterval(updateContent, 1000 * 5);
