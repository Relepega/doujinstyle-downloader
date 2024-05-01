
<a name="v0.3.0-b1"></a>
## [v0.3.0-b1](https://github.com/Relepega/Doujinstyle-downloader/compare/v0.2.0...v0.3.0-b1)

> 2024-05-01

### ✨ New Features ✨

* setting custom download folder inside host
* Readded all hosts
* impl host service
* removed unused static file
* Added batch albumID processing via delimiter ([#6](https://github.com/Relepega/Doujinstyle-downloader/issues/6))
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

### 🐛 Bugfixes 🐛

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

### 🛠️ Code Refactoring 🛠️

* Renamed module
* commented out all console.log lines
* removed debug print
* Removed useless code
* Reverted sse route to the original thing

### 🧹 Chores 🧹

* updated deps
* updated dependencies
* **deps:** updated dependencies

### 🪄 Style 🪄

* removed leftover comments

### Pull Requests

* Merge pull request [#7](https://github.com/Relepega/Doujinstyle-downloader/issues/7) from Relepega/event-driven-rewrite
* Merge pull request [#5](https://github.com/Relepega/Doujinstyle-downloader/issues/5) from Relepega/dependabot/go_modules/golang.org/x/net-0.23.0


<a name="v0.2.0"></a>
## [v0.2.0](https://github.com/Relepega/Doujinstyle-downloader/compare/v0.2.0-b1...v0.2.0)

> 2024-04-15

### ✨ New Features ✨

* **appUtils:** Added more helper methods
* **mediafire:** Folder download feature added
* **mediafire:** Added API types

### 🐛 Bugfixes 🐛

* removed unused variable
* Added checks for edge cases

### Pull Requests

* Merge pull request [#4](https://github.com/Relepega/Doujinstyle-downloader/issues/4) from Relepega/mediafire-folder


<a name="v0.2.0-b1"></a>
## [v0.2.0-b1](https://github.com/Relepega/Doujinstyle-downloader/compare/v0.1.0...v0.2.0-b1)

> 2024-03-17

### ✨ New Features ✨

* Initial support for different services
* Multiple services support
* Added sukidesuost in the services list
* **Makefile:** Added run command
* **Makefile:** update dependencies
* **downloader:** New helper functions
* **queue:** Channel to quit the queue

### 🐛 Bugfixes 🐛

* **downloadFile:** Incomplete downloads
* **form:** Form service select reset
* **playwrightWrapper:** Interrupt handling

### 🛠️ Code Refactoring 🛠️

* Removed dead code
* duplicate vars & switch improvements
* **appUtils:** Moved function to utils
* **downloader:** Moved each site to submodule
* **playwrightWrapper:** merged to single file
* **task:** Moved logic to new Run method
* **webserver:** Moved code

### 🧹 Chores 🧹

* Updated packages
* **deps:** Updated dependencies

### 🪄 Style 🪄

* **downloader:** formatted function definition


<a name="v0.1.0"></a>
## [v0.1.0](https://github.com/Relepega/Doujinstyle-downloader/compare/v0.1.0-b5...v0.1.0)

> 2024-01-14

### ✨ New Features ✨

* **mediafire:** Added isFolder chech


<a name="v0.1.0-b5"></a>
## [v0.1.0-b5](https://github.com/Relepega/Doujinstyle-downloader/compare/v0.1.0-b4...v0.1.0-b5)

> 2024-01-09


<a name="v0.1.0-b4"></a>
## [v0.1.0-b4](https://github.com/Relepega/Doujinstyle-downloader/compare/v0.1.0-b3...v0.1.0-b4)

> 2023-12-31


<a name="v0.1.0-b3"></a>
## [v0.1.0-b3](https://github.com/Relepega/Doujinstyle-downloader/compare/v0.1.0-b2...v0.1.0-b3)

> 2023-12-15


<a name="v0.1.0-b2"></a>
## [v0.1.0-b2](https://github.com/Relepega/Doujinstyle-downloader/compare/v0.1.0-b1...v0.1.0-b2)

> 2023-11-30

### Pull Requests

* Merge pull request [#1](https://github.com/Relepega/Doujinstyle-downloader/issues/1) from Relepega/dependabot/go_modules/internal/playwrightWrapper/github.com/go-jose/go-jose/v3-3.0.1
* Merge pull request [#2](https://github.com/Relepega/Doujinstyle-downloader/issues/2) from Relepega/dependabot/go_modules/internal/downloader/github.com/go-jose/go-jose/v3-3.0.1


<a name="v0.1.0-b1"></a>
## v0.1.0-b1

> 2023-11-18

