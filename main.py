from typing import List
from quart import Quart, render_template, request, abort, make_response, Response
import asyncio
from dataclasses import dataclass

from .fuck import server_root

app = Quart(__name__)

task_list: List[str] = []
active_tasks: int = 0

queue = asyncio.Queue(maxsize=3)
mutex = asyncio.Lock()


@dataclass
class ServerSentEvent:
    data: str
    event: str | None = None
    id: int | None = None
    retry: int | None = None

    def encode(self) -> bytes:
        message = f"data: {self.data}"
        if self.event is not None:
            message = f"{message}\nevent: {self.event}"
        if self.id is not None:
            message = f"{message}\nid: {self.id}"
        if self.retry is not None:
            message = f"{message}\nretry: {self.retry}"
        message = f"{message}\n\n"
        return message.encode("utf-8")


async def render_task_list_ui() -> str:
    elms = []

    if len(task_list) == 0:
        elms.append("Nothing to see here...")

    for index in range(0, len(task_list)):
        elms.append(
            render_template("thing.html", item=task_list[index], elementIndex=index)
        )

    return "".join(elms)


async def process_album(album_id: str):
    global active_tasks

    while True:
        await mutex.acquire()

        if active_tasks >= 4:
            mutex.release()
            continue

        active_tasks += 1
        mutex.release()
        await server_root(album_id)

        await mutex.acquire()
        task_list.remove(album_id)
        mutex.release()

        return


@app.route("/remove-queue-element")
async def remove_queue_element() -> str:
    task_list.pop(int(request.args["index"]))
    return await render_task_list_ui()


@app.route("/do-the-thing", methods=["POST"])
async def add_task() -> str:
    album_id = (await request.form)["AlbumID"]

    await mutex.acquire()
    task_list.append(album_id)
    mutex.release()

    # await process_album(album_id)

    return await render_template(
        "thing.html",
        item=(await request.form)["AlbumID"],
        elementIndex=len(task_list) - 1,
    )


# TODO
@app.route("/stream")
async def sse():
    if "text/event-stream" not in request.accept_mimetypes:
        abort(400)

    async def send_events():
        while True:
            # event = ServerSentEvent(data=render_task_list_ui(), event="list-reload")
            # yield event.encode()
            yield f"data: {await render_task_list_ui()} \n\n"

    events = [event async for event in send_events()]
    return Response(events, mimetype="text/event-stream")


@app.route("/")
async def hello() -> str:
    return await render_template("index.html", name="bocchi")


if __name__ == "__main__":
    app.run(port=5500)
