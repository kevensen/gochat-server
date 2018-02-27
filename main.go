//The entry point for the gochat program
package main

import (
	"flag"
	"net/http"

	"github.com/golang/glog"
)

/*Main entry point.  Flags, Handlers, and authentication providers configured here.
*
* Basic usage: gochat -host=0.0.0.0:8080
*
* Help: gochat --help
*
 */
func main() {
	var host = flag.String("host", ":8080", "The host address of the application.")
	flag.Parse()

	r := newRoom()

	http.Handle("/room", r)

	go r.run()
	glog.Infoln("Starting the web server on", *host)
	if err := http.ListenAndServe(*host, nil); err != nil {
		glog.Errorln("ListenAndServe: ", err)
	}

}
