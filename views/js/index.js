function removeElement(albumID) {
  document.getElementById(albumID).remove();
}

async function copyErrorMessage(albumID) {
  const el = document.getElementById(albumID + "-error");
  await navigator.clipboard.writeText(el.innerText);
  window.alert("Error log of album " + albumID + " copied");
}

class ScrollManager {
  constructor() {
    this.items = [];
  }

  addSelector(s) {
    this.items.push({
      selector: s,
      scrollPos: 0,
    });
  }

  saveState() {
    this.items.forEach((item) => {
      item.scrollPos = document.querySelector(item.selector).scrollTop;
    });
  }

  restoreState() {
    this.items.forEach((item) => {
      document.querySelector(item.selector).scrollTop += item.scrollPos;
    });
  }
}

const sm = new ScrollManager();
sm.addSelector("body");
sm.addSelector("#queue");
sm.addSelector("#active");
sm.addSelector("#ended");

// TODO: make it less react-like
async function updateContent() {
  const tasksContainer = document.querySelector("#tasks");

  const res = await fetch(document.location.origin + "/api/task/render");
  const newContent = await res.text();

  tasksContainer.innerHTML = newContent;
  htmx.process(tasksContainer);

  sm.restoreState();
}
// async function updateContent() {
//   const tasksContainer = document.querySelector("#tasks");
//
//   let bodyScroll = document.body.scrollTop;
//   let queueScroll = document.getElementById("queue").scrollTop;
//   let activeScroll = document.getElementById("active").scrollTop;
//   let endedScroll = document.getElementById("ended").scrollTop;
//
//   const res = await fetch(document.location.origin + "/api/task/render");
//   const newContent = await res.text();
//
//   tasksContainer.innerHTML = newContent;
//   htmx.process(tasksContainer);
//
//   document.body.scrollTop += bodyScroll;
//   document.getElementById("queue").scrollTop += queueScroll;
//   document.getElementById("active").scrollTop += activeScroll;
//   document.getElementById("ended").scrollTop += endedScroll;
// }

setInterval(updateContent, 1000 * 5);
