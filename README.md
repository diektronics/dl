* REST interface with mysql backend
* Run in some port, use apache proxy to show it in a particular URL diektronics.com/downloader
* Link per line
* Maybe use tvd.downloader for plowdown?
* Do we need notifier?
* After download hook for unrar
* Optional name to rename downloaded file
* Gorilla mux for API
* Each download may have several links
* 4 statuses: QUEUE, RUNNING, SUCCESS, ERROR (and a error explanation in another field)
* downloads table:
  * ID
  * name (optional, final file will be renamed to this)
  * timestamp
  * links (list in a text field)
  * status (string, not need for purity here)
  * error (text field, empty unless status is ERROR)
  * posthook (unrar, delete downloaded files)
