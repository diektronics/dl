TODO
====
* Recovery mechanism: get everything in dB that is QUEUED and RUNNING
* Use SSL by default
* Use user certificates for multiuser support
* AngularJS frontend.

DONE
====
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
    error varchar(255) NOT NULL DEFAULT "",
    posthook varchar(255) NOT NULL DEFAULT "",
    created_at DATETIME NOT NULL,
    modified_at DATETIME NOT NULL,
    PRIMARY KEY(id));
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
    created_at DATETIME NOT NULL,
    modified_at DATETIME NOT NULL,
    PRIMARY KEY(id),
    FOREIGN KEY(download_id) REFERENCES downloads(id));
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