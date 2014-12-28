package cli

import (
	"flag"
	"fmt"
	"log"
)

func main() {

	const (
		DEFAULT_QUERY = "METAQUERY WHERE userid IS %s QUERY TYPE IN cbg, smbg SORT BY time AS Timestamp REVERSED"
	)

	who := flag.String("w", "", "who we are fetching data for")
	query := flag.String("q", DEFAULT_QUERY, "the query to execute")
	usr := flag.String("u", "", "tidepool username")
	pw := flag.String("p", "", "tidepool password")

	flag.Parse()

	log.Printf("for [%s] run [%s] logged in as [%s]", *who, fmt.Sprintf(*query, *who), *usr)

}
