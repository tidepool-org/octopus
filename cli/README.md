octopus cli
===========

This is basic tool for executing the query endpoint on octopus

### Options

* -e : the api url for your environment e.g. http://localhost:8009, the default is "https://devel-api.tidepool.io"

* -t : the types of data wanted, the default is "cbg, smbg, bolus, wizard"

* -w : id of who we are fetching data for

### Example

```
go run cli/cli.go -w 7e881aec8f -t smbg
```
