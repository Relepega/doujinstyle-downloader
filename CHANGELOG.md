
<a name="v0.4.0-a1"></a>
## [v0.4.0-a1](https://github.com/Relepega/doujinstyle-downloader/compare/v0.3.2...v0.4.0-a1)

> 2025-01-01

### Chore

* removed debugging strings
* updated app version
* updated deps
* Removed old leftovers
* Fixed tests and empty file errors
* **taskRunner:** error messages now point to correct pkg and not to taskRunner

### Docs

* Added contacts on QnA section in the readme

### Feat

* getting places
* added proper debug commands and config
* Implemented all the remaining modes for task DELETE method
* Restructured crud api
* Make use of SSE to send live updates to all connected clients
* re-added data-agnostic observer module
* added a service validator function
* re-added Mega
* embed filehost url to task
* Incomplete impl for UI messaging
* added the find method
* added a unique IDentifier private field with relative getter method
* added the Slug getter method
* Added a proper Unique ID
* update taskData with new intelligible data
* Filehost constructor function now accepts a page as a parameter
* deferring playwright context close on task completion
* check internet connection availability
* moving logic to a single module to mimic a package
* Implemented all remaining modes for task PATCH method
* added task already present in DB checks
* Readded a non-race condition checking run script
* fixed task progression impl
* added tests & functionalities
* Added a GetAll method to Tracker and fixed some compiler errors
* handleTaskUpdateState route is finished
* new function that returns node selectively
* added partial impl to task state update endpoint
* Added queue runner logic
* introduced main application logic
* Almost completed mediafire core logic
* added guard clause for empty registration lists
* renamed "aggregator" dir to plural & added sukidesuost
* Added Doujinstyle aggregator
* **engine:** added function that returns all nodes with same progress state
* **engine:** Added method to get all tasks with the same download state
* **initters.engine:** Completed implementation of taskRunner function
* **playwrightWrapper:** Added support for custom downloadPath

### Fix

* full url being built even if slug is already a full url
* removed unused import
* removed end task on error when only first getfilename has been processed
* fixed comparator function filters
* changed queryselector for file format evaluation
* changed some old refs to new ones
* fixed erroneous task render & event send on bad request
* **engine:** Fixed logic errors in half-implemented functions and fixed nullptr dereference
* **engine:** fixed local error where task wasn't resetted correctly
* **sse:** Fixed edge case in UIEventBuilder on replaceNode event
* **taskRunner:** fixed logical error that sets the page as not found when it's not
* **taskRunner:** now detecting file name & extension through both aggregator and filehost
* **webserver:** fixed logic errors while comparing completed failed and succeeded tasks
* **webserver:** fixed logic issues on task PATCH method

### Refactor

* removed old or unused code
* changed data type for aggregator and filehost
* setting taskdata.displayname at different points if one fails
* changed utils' download function signature to add flexibility
* changed helper function name to MkdirAll to reflect the std package naming
* Filehost constructor function now accepts a page as a parameter (follows #c17fd0b)
* General interface and method updates
* changed parameter from *int8 to func(int 8)
* Removed webserver initter because of its uselessness
* updated AggregatorConstrFn parameters
* Changed interface names & interface function names
* removed unused return value
* Moved init functions to a separate module
* **dsdl:** made error message constants

### Tests

* fixed tests to reflect changes in filehost impl
* Fixed implementation due to signature changes


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
* publish dl update to queue evt subscriber
* impl host service
* removed unused static file
* Reset input value on form submit
* Readded sukidesuost service handler
* added .prettierrc
* Readded all hosts
* Added batch albumID processing via delimiter ([#6](https://github.com/Relepega/doujinstyle-downloader/issues/6))
* added "Downloads" and subdirs into exclusions
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

### BREAKING CHANGE


Removed windows i386 target due to being sunsetted a
long time ago

Removed windows i386 target due to being sunsetted a
long time ago


<a name="v0.1.0-b1"></a>
## v0.1.0-b1

> 2023-11-18

