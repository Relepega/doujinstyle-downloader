function removeElement(albumID) {
  document.getElementById(albumID).remove();
}

async function copyErrorMessage(albumID) {
  const el = document.getElementById(albumID + "-error");
  await navigator.clipboard.writeText(el.innerText);
  window.alert("Error log of album " + albumID + " copied");
}
