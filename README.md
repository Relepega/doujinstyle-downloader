# doujinstyle-downloader

A User-friendly app made for effortlessly downloading music from
[doujinstyle](https://doujinstyle.com/) written in Go, using HTML Templates and
[htmx.js](https://htmx.org/).

![GitHub tag checks state](https://img.shields.io/github/checks-status/relepega/doujinstyle-downloader/HEAD)
![GitHub Release](https://img.shields.io/github/v/release/Relepega/doujinstyle-downloader)
[![Go Report Card](https://goreportcard.com/badge/github.com/relepega/doujinstyle-downloader)](https://goreportcard.com/report/github.com/relepega/doujinstyle-downloader)
[![codebeat badge](https://codebeat.co/badges/4fd08f5a-b2d8-49b9-b0d4-a37cd6f6c3dc)](https://codebeat.co/projects/github-com-relepega-doujinstyle-downloader-main)

## Table of content

- [doujinstyle-downloader](#doujinstyle-downloader)
  - [Table of content](#table-of-content)
  - [Why](#why)
  - [Installation](#installation)
  - [Usage](#usage)
  - [Build](#build)
  - [Contributing](#contributing)
    - [instructions](#instructions)
  - [QnA](#QnA)

## Why

Due to the fact that during conventions and other events there is a surge in
album publications, I wasn't able to keep up with the releases while downloading
manually.

## Installation

I will "officially" build for these OSes and architectures: Windows (x64), Linux
(x64) and MacOS (ARM).

If you want to run this program on another os/arch, you have to build it
yourself. More informations [in the build section](#Build).

You can download it
[in the releases page](https://github.com/Relepega/doujinstyle-downloader/releases).

To install the app, you just need to download the release zip, unzip it inside a
folder with a name you can remember (optional), then open the app! No
installation required!

## Usage

**Prerequisite for Arch Linux users**: you need to install the `playwright`
package from the AUR.

1. Open the application.
2. Navigate to [http://127.0.0.1:5522/](http://127.0.0.1:5522/).
3. Get the ID of the music you want to download from the page url and copy it
   (e.g: in this url `https://doujinstyle.com/?p=page&type=1&id=22816` the id is
   `22816`)
4. Paste the ID into the input field of the WebUI and press the "Add download
   task" button.
5. Wait for the download to complete. After that, the box should be moved into
   the Ended Tasks column.
6. If the download succeds, the box will have a green-ish background color. If
   it fails, said color will be red-ish.
   - If the download fails, there will be a box that displays the encountered
     error with a button that allows you to copy it. You can use it to fill a
     bug report later!
7. Profit!

## Build

To build the app yourself, follow these steps:

0. (Arch Linux only): install the `playwright` package from the AUR.
1. Install these packages to get started: `git`, `go`, `make`.
2. Clone the repo
   `git clone https://github.com/Relepega/doujinstyle-downloader.git`.
3. Run the command `go get && make`. You will find the built binaries into the
   Build folder.
4. Done!

## Contributing

I welcome any and all contributions! Here are some ways you can get started:

1. Report bugs: If you encounter any bugs, please let us know. Open up an issue
   and let us know the problem.
2. Contribute code: If you are a developer and want to contribute, follow the
   instructions below to get started!
3. Suggestions: If you don't want to code but have some awesome ideas, open up
   an issue explaining some updates or imporvements you would like to see!
4. Documentation: If you see the need for some additional documentation, feel
   free to add some!

### instructions

1. Fork this repository
2. Clone the forked repository
3. Add your contributions (code or documentation)
4. Commit and push
5. Create a Pull Request
6. Wait for it to be merged

### QnA

**Q**. I get the following error:
`listen tcp ...: bind: An attempt was made to access a socket in a way forbidden by its access permissions.`.
How do i fix it?

**A**. The error is either caused by another process using the port 5522, or by
HyperV. To fix it you have 3 choices:

- By stopping the process that is using that port app.
- (Hyper-V only) By using the command `Restart-Service hns` in a powershell
  prompt.
- Change the application port in the `config.toml` file (you can edit it with a
  notepad app)

If none of these helps you out, you can open a new issue. You need to accurately
describe what the issue is, your os, os version, app version and the steps to
reproduce the issue (if any).

**Q**. I get the following error:
`listen tcp ... bind: The requested address is not valid in its context.` How do
i fix it?

**A**. Change the `Host` value in the [server] section in the config file to a
valid IP address. Either change it to 127.0.0.1 if you're unsure of what to do
or use the valid current network IP address of your device.
