import os
import re
from typing import Union
from urllib.parse import urlparse

from flask import Flask, request
from playwright.async_api import Page, async_playwright

PLAYWRIGHT_DEBUG = True
FLASK_PORT = 9750
DOWNLOAD_ROOT = os.path.join(os.getcwd(), "Downloads")

app = Flask(__name__)


async def usePlaywright(
    browser: str = "chromium",
    headless: bool = not PLAYWRIGHT_DEBUG,
    timeout: int = 0,
):
    pw = await async_playwright().start()

    match browser:
        case "chromium":
            bt = pw.chromium
        case "firefox":
            bt = pw.firefox
        case "webkit":
            bt = pw.webkit
        case _:
            raise Exception("browser is not of type chromium, firefox, or webkit")

    bw = await bt.launch(headless=headless, timeout=timeout)
    ctx = await bw.new_context()

    ctx.set_default_timeout(timeout)

    return (pw, bw, ctx)


async def handle_popup(popup: Page) -> None:
    await popup.wait_for_load_state()
    await popup.close()


async def craft_filename(page: Page) -> str:
    album = await page.evaluate("document.querySelector('h2').innerText")
    artist = await page.evaluate("document.querySelectorAll('.pageSpan2')[0].innerText")

    el = await page.query_selector('text="Format:"')
    format = await page.evaluate(
        """
            (element) => {
                let sibling = element.nextElementSibling;
                return sibling.innerText
            }
        """,
        el,
    )

    event: Union[str, None] = None
    try:
        event = re.findall(
            r"C\d+|M\d\-\d+",
            await page.evaluate("document.querySelectorAll('.pageSpan2')[1].innerText"),
        )[0]
    except Exception:
        event = None

    return f"{artist} — {album}{f' [{event}]' if event is not None else ''} [{format}]"


async def mediafire(album_name: str, dl_page: Page) -> None:
    extension: str = re.findall(
        r"\.[a-zA-Z0-9]+",
        await dl_page.evaluate("document.querySelector('.filetype').innerText"),
    )[0].lower()

    async with dl_page.expect_download() as dl_info:
        await dl_page.evaluate("document.querySelector('#downloadButton').click()")

    dl_handler = await dl_info.value

    await dl_handler.save_as(os.path.join(DOWNLOAD_ROOT, album_name + extension))


async def mega(album_name: str, dl_page: Page) -> None:
    while (
        await dl_page.evaluate("document.querySelector('.js-default-download')") is None
    ):
        await dl_page.wait_for_timeout(500)

    extension = await dl_page.evaluate("document.querySelector('.extension').innerText")

    async with dl_page.expect_download() as dl_info:
        await dl_page.evaluate("document.querySelector('.js-default-download').click()")

    dl_handler = await dl_info.value

    await dl_handler.save_as(os.path.join(DOWNLOAD_ROOT, album_name + extension))


async def main(url: str) -> None:
    # url = "https://doujinstyle.com/?p=page&type=1&id=22378"
    # url = "https://doujinstyle.com/?p=page&type=1&id=16315"

    if not os.path.exists(DOWNLOAD_ROOT) or not os.path.isdir(DOWNLOAD_ROOT):
        os.makedirs(DOWNLOAD_ROOT)

    pw, browser, ctx = await usePlaywright()

    page = await ctx.new_page()
    await page.goto(url)

    album_name = await craft_filename(page)

    async with ctx.expect_page() as p_info:
        await page.evaluate("document.querySelector('#downloadForm').click()")

    dl_page = await p_info.value
    await dl_page.wait_for_load_state()

    dl_page.on("popup", handle_popup)

    match urlparse(dl_page.url).netloc:
        case "www.mediafire.com":
            await mediafire(album_name, dl_page)
        case "mega.nz":
            await mega(album_name, dl_page)
        case _:
            pass

    await dl_page.close()

    await ctx.close()
    await browser.close()
    await pw.stop()


@app.route("/", methods=["GET"])
async def server_root():
    album_id = int(request.args.get("id", -1))

    if album_id == -1:
        return "URL is not set"

    await main(f"https://doujinstyle.com/?p=page&type=1&id={album_id}")

    return f"Got album ID {album_id}"


if __name__ == "__main__":
    print(f"Opening web server at http://127.0.0.1:{FLASK_PORT}/")
    app.run(port=FLASK_PORT)
