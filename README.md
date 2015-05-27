========
| v3.0 |
========
TODO
====
* Use user certificates for multiuser support

DONE v3.0
=========
* Using grpc for both RPCs and configuration.

DONE v2.5
=========
* Allow to download multiple episodes of the same show without confusing the dB
* Allow to donwload shows in tvd frontend that don't have the proper name structure (Sleepy Hollow)
* Use numbers in hooks to properly sort them
* Fix bug that kills the backend when deleting twice a download in a slow network from web frontend
* Download progress reporting
  * Use plowprobe --printf=%f%t%s%n <URL>
  * Then os.Stat() corresponding .part file in dir, and divide/normalize.
  * os.Stat() calls can happen in a goroutine that lives in paralel with download().

    var size int64
    if fi, err := os.Stat(dl.Name); err != nil {
      size = fi.Size()
    }

  * Probably a dB change to keep the % of data downloaded.

DONE v2.0
=========
* Mirrored in github.
* Make dl into a backend with an RPC API.
  * GetAll
  * Download
  * Get
  * Del
  * HookNames
* Run Downloader.recovery() in New()
* Add destination to Download
* Create RENAME hook
* Make tvd into a frontend
* Merged with tvd
* Made downloader into a backend
* Two different frontends:
  * web for:
    * manual downloads
    * visualization
  * XML scrapper for automatic TV Shows download


DONE v1.0
=========
* Each download may have several links
* 4 statuses: QUEUED, RUNNING, SUCCESS, ERROR (and a error explanation in another field)
* downloads table:
  * ID
  * name (optional, final file will be renamed to this)
  * status (string, not need for purity here)
  * error (text field, empty unless status is ERROR)
  * posthook (unrar, delete downloaded files)
  * timestamp
  create table downloads (
    id INT UNSIGNED NOT NULL AUTO_INCREMENT,
    name varchar(255) NOT NULL,
    status varchar(10) NOT NULL DEFAULT "QUEUED",
    error varchar(2048) NOT NULL DEFAULT "No errors",
    posthook varchar(255) NOT NULL DEFAULT "",
    destination varchar(1024) NOT NULL DEFAULT "",
    created_at DATETIME NOT NULL,
    modified_at DATETIME NOT NULL,
    PRIMARY KEY(id));
  alter table downloads add index status (status);
* links table:
  * ID
  * download_id
  * url
  * status
  create table links (
    id INT UNSIGNED NOT NULL AUTO_INCREMENT,
    download_id INT UNSIGNED NOT NULL,
    url varchar(255) NOT NULL,
    status varchar(10) NOT NULL DEFAULT "QUEUED",
    percent float NOT NULL DEFAULT 0.0,
    created_at DATETIME NOT NULL,
    modified_at DATETIME NOT NULL,
    PRIMARY KEY(id),
    FOREIGN KEY(download_id) REFERENCES downloads(id));
  alter table links add index status (status);
* REST interface with mysql backend
* Run in some port, use apache proxy to show it in a particular URL diektronics.com/downloader
* Link per line
* Do we need notifier?
* After download hook for unrar
* Optional name to rename downloaded file?
* Gorilla mux for API
* Parallel workers for downloading? Yes, queue is scheduler
* Consider having a big queue channel, and resize it if full. In order for this to work, only one goroutine may add stuff to the channel...
* Downloads to temporary dir. Moves final file to ~/Downloads
* Name is for a dir in ~/Downloads. All downloaded stuff goes in there
* We can have different queues/workers for posthooks (only implement unrar and delete)
* We can create a channel/worker per download. As downloader.workers finish with their links, they send the result to that channel. The worker in there updates the types.Download until an error is received or all links have been downloaded. Then executes the posthooks (chan to workers again)
* Use SSL by default
* Refresh page with a timer.
* AngularJS frontend.
* Update links status.
* Make a better DELETE button/link.
* Recovery mechanism: get everything in dB that is QUEUED and RUNNING
