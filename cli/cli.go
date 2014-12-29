package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func runQuery(queryToRun, env string) {

	url := env + "/query/data"

	var jsonStr = []byte(queryToRun)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("Query [%s] %s ", resp.Status, string(body))

}

func main() {

	const (
		DEFAULT_QUERY = "METAQUERY WHERE userid IS %s QUERY TYPE IN %s SORT BY time AS Timestamp REVERSED"
		DEFAULT_TYPES = "cbg, smbg, bolus, wizard"
		DEFAULT_ENV   = "https://devel-api.tidepool.io"
	)

	who := flag.String("w", "", "who we are fetching data for")
	query := flag.String("q", DEFAULT_QUERY, "the query to execute")
	types := flag.String("t", DEFAULT_TYPES, "the types of data wanted")
	env := flag.String("e", DEFAULT_ENV, "the api url for your environment e.g. http://localhost:8009")
	//usr := flag.String("u", "", "tidepool username")
	//pw := flag.String("p", "", "tidepool password")

	flag.Parse()

	log.Printf("for [%s] run [%s] in env [%s]", *who, fmt.Sprintf(*query, *who, *types), *env)

	runQuery(fmt.Sprintf(*query, *who, *types), *env)

}
