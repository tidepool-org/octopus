package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func runQuery(queryToRun string) {

	url := "http://localhost:8009/query/data"

	var jsonStr = []byte(queryToRun)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("Query status: ", resp.Status)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("Query result: ", string(body))

}

func main() {

	const (
		DEFAULT_QUERY = "METAQUERY WHERE userid IS %s QUERY TYPE IN cbg, smbg SORT BY time AS Timestamp REVERSED"
	)

	who := flag.String("w", "", "who we are fetching data for")
	query := flag.String("q", DEFAULT_QUERY, "the query to execute")
	usr := flag.String("u", "", "tidepool username")
	//pw := flag.String("p", "", "tidepool password")

	flag.Parse()

	log.Printf("for [%s] run [%s] logged in as [%s]", *who, fmt.Sprintf(*query, *who), *usr)

	runQuery(fmt.Sprintf(*query, *who))

}
