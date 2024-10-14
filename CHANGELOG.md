
<a name="v0.3.2"></a>
## [v0.3.2](https://github.com/Relepega/doujinstyle-downloader/compare/v0.3.1...v0.3.2)

> 2024-10-14

### Build

* differentiated build-all script and build only for local machine
* When building, generate new changelog
* **Makefile:** Make all go run & test commands to run with race condition checker
* **air:** removed app config from watchlist to avoid infinite loop

### Chore

* updated deps

### Docs

* Added go docs for the remaining source code files
* Added go comment docs for the librarycode file. Also renamed the wrapper struct

### Feat

* adding tempdir to global store to make its access easier
* Added skeletons for v2 impl of webserver and task modules
* added more selectors to fetch a download url
* Merge branch 'new-queue'
* Added an alternative temp dir
* added function to check if value if tq already holds a value
* Added a library-sort of wrapper to keep synced both the queue and the tracker
* Impl new version of queue / generic code
* added in-memory store
* **appUtils:** Hacked a way through storing the tempdir as global var :eyes:
* **main:** limited playwright browser choice & fixed config load error on non-existing file
* **services:** Added new function that returns new Service from url

### Fix

* removed runner file because it should be implemented outside the module
* **configManager:** Changed actual order of config settings
* **configManager:** Fixed bug where an updated config is never saved to file
* **mediafire:** Fixed download link fetching error
* **queue:** Fixed errors on app close

### Refactor

* Changes done to reflect the correct behavior of things as done in the tests
* Changed some logic and added new "Lib functions"
* Changed some struct field types & impl
* Changed runner function args & consequent impl
* use config from inmemory store
* **GetAppTempDir:** Return the value stored in global store

### Test

* Finished writing test suite for the package
* Finished fixing tests
* Added multi-threading test
* Added some basic usage test, will need to add some more advanced ones


<a name="v0.3.1"></a>
## [v0.3.1](https://github.com/Relepega/doujinstyle-downloader/compare/v0.3.0...v0.3.1)

> 2024-07-02

### Build

* **air:** added app config file to realod watchlist

### Feat

* Added application restart api route & UI button
* improved logging
* **mega:** Added "low quota" case handling
* **sukidesuost:** Added more queries for fetching album name
* **sukidesuost:** Added more selectors for fetching download url

### Fix

* likely to fix the task duplication issue
* **mega:** Fixed download progress report
* **sukidesuost:** Fixed filename not using correct separator

### Refactor

* Early template rendering to save time


<a name="v0.3.0"></a>
## [v0.3.0](https://github.com/Relepega/doujinstyle-downloader/compare/v0.3.0-b2...v0.3.0)

> 2024-06-19

### Docs

* **changelog:** Updated changelog to latest version

### Feat

* Added field for media name to be displayed
* added ID field to the event
* Clear screen after playwright complains about your linux distro
* Added a general-purpose queue implementation

### Refactor

* Make auto fetch current tag


<a name="v0.3.0-b2"></a>
## [v0.3.0-b2](https://github.com/Relepega/doujinstyle-downloader/compare/v0.3.0-b1...v0.3.0-b2)

> 2024-06-13

### Chore

* fix typo
* Updated deps

### Docs

* Updated README
* updated README

### Feat

* Improved download URL and audio format type recognizers
* added changelog
* **sukidesuost:** Can use url substring
* **webUI:** Save last used service to localstorage

### Fix

* **EvaluateFilename:** wrong selectors when evaluating audio filetypes

### Refactor

* Strip whitespaces by default


<a name="v0.3.0-b1"></a>
## [v0.3.0-b1](https://github.com/Relepega/doujinstyle-downloader/compare/v0.2.0...v0.3.0-b1)

> 2024-05-01

### Build

* **deps:** Bump golang.org/x/net from 0.22.0 to 0.23.0

### Build

* Added changelog generation script
* bumped build number

### Chore

* updated deps
* updated dependencies
* **deps:** updated dependencies

### Docs

* **README:** updated README
* **readme:** Added branch todo list

### Feat

* setting custom download folder inside host
* Readded all hosts
* impl host service
* removed unused static file
* Added batch albumID processing via delimiter ([#6](https://github.com/Relepega/doujinstyle-downloader/issues/6))
* Readded sukidesuost service handler
* added .prettierrc
* added "Downloads" and subdirs into exclusions
* Reset input value on form submit
* publish dl update to queue evt subscriber
* Added queue subscriber support
* Added global publishers
* Added host download logic
* Added hosts, mediafire is first implemented
* impl service interface and doujinstyle logic
* **webserver:** support for multiple open connections

### Fix

* Fixed Jottacloud parsing wrong file extension
* Fixed UI responsiveness
* Fixed UI css
* logging hardcoded service name
* Implemented OpenServicePage function
* Graceful shutdown for queue & playwright
* Fixed 'node not found' error
* Fixed js switch - cases didn't have a scope
* **configManager:** Capitalized downloads dir
* **mediafire:** folder downlaod misplaced files
* **templates:** hacked problematic "%" in json parse
* **webserver:** graceful shutdown

### Refactor

* Renamed module
* commented out all console.log lines
* removed debug print
* Removed useless code
* Reverted sse route to the original thing

### Style

* removed leftover comments

### Pull Requests

* Merge pull request [#7](https://github.com/Relepega/doujinstyle-downloader/issues/7) from Relepega/event-driven-rewrite
* Merge pull request [#5](https://github.com/Relepega/doujinstyle-downloader/issues/5) from Relepega/dependabot/go_modules/golang.org/x/net-0.23.0

### BREAKING CHANGE


Please backup your existing config file, let the app
create a new one and then restore your settings in the new file. this
ensures more flexibility in the app structure.


<a name="v0.2.0"></a>
## [v0.2.0](https://github.com/Relepega/doujinstyle-downloader/compare/v0.2.0-b1...v0.2.0)

> 2024-04-15

### Feat

* **appUtils:** Added more helper methods
* **mediafire:** Folder download feature added
* **mediafire:** Added API types

### Fix

* removed unused variable
* Added checks for edge cases

### Pull Requests

* Merge pull request [#4](https://github.com/Relepega/doujinstyle-downloader/issues/4) from Relepega/mediafire-folder


<a name="v0.2.0-b1"></a>
## [v0.2.0-b1](https://github.com/Relepega/doujinstyle-downloader/compare/v0.1.0...v0.2.0-b1)

> 2024-03-17

### Build

* Incremented minor feat build number

### Chore

* Updated packages
* **deps:** Updated dependencies

### Feat

* Initial support for different services
* Multiple services support
* Added sukidesuost in the services list
* **Makefile:** Added run command
* **Makefile:** update dependencies
* **downloader:** New helper functions
* **queue:** Channel to quit the queue

### Fix

* **downloadFile:** Incomplete downloads
* **form:** Form service select reset
* **playwrightWrapper:** Interrupt handling

### Refactor

* Removed dead code
* duplicate vars & switch improvements
* **appUtils:** Moved function to utils
* **downloader:** Moved each site to submodule
* **playwrightWrapper:** merged to single file
* **task:** Moved logic to new Run method
* **webserver:** Moved code

### Style

* **downloader:** formatted function definition


<a name="v0.1.0"></a>
## [v0.1.0](https://github.com/Relepega/doujinstyle-downloader/compare/v0.1.0-b5...v0.1.0)

> 2024-01-14

### Build

* Releasing stable version 0.1.0 🎉

### Docs

* **README:** Polished first elements's structure

### Feat

* **mediafire:** Added isFolder chech


<a name="v0.1.0-b5"></a>
## [v0.1.0-b5](https://github.com/Relepega/doujinstyle-downloader/compare/v0.1.0-b4...v0.1.0-b5)

> 2024-01-09


<a name="v0.1.0-b4"></a>
## [v0.1.0-b4](https://github.com/Relepega/doujinstyle-downloader/compare/v0.1.0-b3...v0.1.0-b4)

> 2023-12-31


<a name="v0.1.0-b3"></a>
## [v0.1.0-b3](https://github.com/Relepega/doujinstyle-downloader/compare/v0.1.0-b2...v0.1.0-b3)

> 2023-12-15


<a name="v0.1.0-b2"></a>
## [v0.1.0-b2](https://github.com/Relepega/doujinstyle-downloader/compare/v0.1.0-b1...v0.1.0-b2)

> 2023-11-30

### Build

* **deps:** Bump github.com/go-jose/go-jose/v3 in /internal/downloader
* **deps:** Bump github.com/go-jose/go-jose/v3

### Pull Requests

* Merge pull request [#1](https://github.com/Relepega/doujinstyle-downloader/issues/1) from Relepega/dependabot/go_modules/internal/playwrightWrapper/github.com/go-jose/go-jose/v3-3.0.1
* Merge pull request [#2](https://github.com/Relepega/doujinstyle-downloader/issues/2) from Relepega/dependabot/go_modules/internal/downloader/github.com/go-jose/go-jose/v3-3.0.1

### BREAKING CHANGE


Removed windows i386 target due to being sunsetted a
long time ago

Removed windows i386 target due to being sunsetted a
long time ago


<a name="v0.1.0-b1"></a>
## v0.1.0-b1

> 2023-11-18

