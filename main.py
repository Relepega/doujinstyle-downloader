import asyncio
import os
from urllib.parse import urlparse

from playwright.async_api import Page, Playwright, async_playwright

DEBUG = True
DOWNLOAD_ROOT = os.path.join(os.getcwd(), "Downloads")


async def usePlaywright(
    browser: str = "chromium",
    headless: bool = not DEBUG,
    timeout: int = 0,
):
    pw = await async_playwright().start()
    bw = None

    match browser:
        case "chromium":
            bw = pw.chromium
        case "firefox":
            bw = pw.firefox
        case "webkit":
            bw = pw.webkit
        case _:
            raise Exception("browser is not of type chromium, firefox, or webkit")

    bw = await bw.launch(headless=headless, timeout=timeout)
    ctx = await bw.new_context()

    ctx.set_default_timeout(timeout)

    return (pw, bw, ctx)


async def mediafire(ds_page: Page, dl_page: Page):
    filename = "test mediafire filename"
    extension = "." + dl_page.url.split(".")[-1].split("/")[0]

    async with dl_page.expect_download() as dl_info:
        await dl_page.evaluate("document.querySelector('#downloadButton').click()")

    dl_handler = await dl_info.value

    await dl_handler.save_as(os.path.join(DOWNLOAD_ROOT, filename + extension))


async def mega(ds_page: Page, dl_page: Page):
    while (
        await dl_page.evaluate("document.querySelector('.js-default-download')") is None
    ):
        await dl_page.wait_for_timeout(500)

    filename = "test mega filename"
    extension = await dl_page.evaluate("document.querySelector('.extension').innerText")

    async with dl_page.expect_download() as dl_info:
        await dl_page.evaluate("document.querySelector('.js-default-download').click()")

    dl_handler = await dl_info.value

    await dl_handler.save_as(os.path.join(DOWNLOAD_ROOT, filename + extension))


async def main():
    url = "https://doujinstyle.com/?p=page&type=1&id=22378"
    # url = "https://doujinstyle.com/?p=page&type=1&id=16315"

    print(DOWNLOAD_ROOT)
    if not os.path.exists(DOWNLOAD_ROOT) or not os.path.isdir(DOWNLOAD_ROOT):
        os.makedirs(DOWNLOAD_ROOT)

    pw, browser, ctx = await usePlaywright()

    page = await ctx.new_page()
    await page.goto(url)

    async with ctx.expect_page() as p_info:
        await page.evaluate("document.querySelector('#downloadForm').click()")

    dl_page = await p_info.value
    await dl_page.wait_for_load_state()

    match urlparse(dl_page.url).netloc:
        case "www.mediafire.com":
            await mediafire(page, dl_page)
        case "mega.nz":
            await mega(page, dl_page)
        case _:
            pass

    for p in ctx.pages:
        if "doujinstyle.com" not in p.url:
            await p.close()

    await ctx.close()
    await browser.close()
    await pw.stop()


if __name__ == "__main__":
    asyncio.run(main())
