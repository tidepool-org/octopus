[DEPRECATED] octopus
=======

[![Build Status](https://travis-ci.org/tidepool-org/octopus.png)](https://travis-ci.org/tidepool-org/octopus)

**NOTE:** This was meant to be the data query J-API for Tidepool, but now after learning some hard lessons we are in the process of stepping away from that idea.

It was initially created by [anderspitman](https://github.com/anderspitman) upon request by Tidepool.

The purpose of this API is to provide the framework for queries against the Tidepool platform.

The first version of this API will be largely faked, designed to mimic the eventual API but with a tightly constrained set of functionality.

The version that was originally created was specified to do a simple query against the data stream to return the date and time of the last known data item for a given device. The next version will also include a constrained version of a general-purpose query API.

Authenticated calls must use a standard Tidepool authentication token in the headers. See the login function within [platform-client](https://github.com/tidepool-org/platform-client/blob/master/index.js) for details on how to log in and get one.

As of Jan 15, 2015 here is what we have implemented for the current version:

## Status

    GET /status

Does not require authentication. Returns 200 and OK if the status is good, 500 if the data store is unreachable.

## Last Entry for a user

    GET /upload/lastentry/{userid}

Requires authentication. Returns 200 and an ISO8601 timestamp of the last data record for a given userid.

## Last Entry for a user's device

    GET /upload/lastentry/{userid}/{deviceid}

Requires authentication. Returns 200 and an ISO8601 timestamp of the last data record for a given userid / deviceid combination.

## Query submission

    POST /query/data

Requires authentication. The body of the post is the query text.

The result will be 200 response with the MIME type of application/json, containing a JSON object with the results. If the query generates an empty set, the result will be 200 with an empty array. If the query fails to parse, the result will be 400.


## Supported Query Formats:

The following queries are written as if they were using [TQL](http://developer.tidepool.io/queries-and-notifications/), but they are not -- the query parser for this release expects these queries and no others, and only parts of these queries are recognized. Please don’t expect to experiment with the query language yet, as it won’t accept variation.

### Supported Query Formats:

These are the two supported metaquery formats where you can use either the tidepool user's id or the email address associated with the users account

METAQUERY
    WHERE userid IS 12d7bc90fa
    ...

METAQUERY
    WHERE emails CONTAINS foo@bar.com
    ...


### Query Examples:

Query to get all of a user’s update records:

    METAQUERY
        WHERE userid IS 12d7bc90fa

    QUERY
        TYPE IN update

Query to get all of a user’s update records using the email address

    METAQUERY
        WHERE emails CONTAINS foo@bar.com

    QUERY
        TYPE IN update

Query to get a block of records in a given time range:

    METAQUERY
        WHERE userid IS 12d7bc90fa

    QUERY
        TYPE IN cbg, smbg, bolus, wizard
        WHERE time > starttime AND time < endtime

Query to get a block of records within a set of one or more upload IDs:
    METAQUERY
        WHERE userid IS 12d7bc90fa

    QUERY
        TYPE IN cbg
        WHERE uploadId IN 4oiyhsdkh, 23498jsjsaf, ljlsadjfljasdf

You can also say NOT IN to reverse the sense of the test.

You CANNOT currently combine the two types of WHERE clauses.

Result will be a JSON array with individual records corresponding to the selected types, reverse sorted by date (from newest to oldest).

The only acceptable `METAQUERY` is to query for a single userid. Aggregate metaqueries are not supported, and only the userids we give you will work.

**NOTE:** Results are sorted by the `time` field.

`TYPE IN` must be followed by a comma-separated list of types as defined in the [data formats documentation](http://developer.tidepool.io/data-model/v1/).

The `WHERE` clause for time must either be:

    WHERE time > starttime
or

    WHERE time > starttime AND time < endtime

Both starttime and endtime must be ISO 8601 timestamps referenced to UTC (example: `2014-12-11T04:44:16Z`).

Whitespace and upper/lower case are ignored; the formatting above makes it easier to read but it’s unimportant.

The `WHERE` clause for containment must look like this:

    WHERE fieldname [IN|NOT IN] listOfValues

The `fieldname` can be any supported fieldname in the record; listOfValues is a comma-separated or space-separated list of values for that field. It's intended that fieldname is uploadId and that the values are ID strings; other values may not give the desired results but you're welcome to try.

## Running Queries:

-- use "source" (also known as ".") to load it, as in ```. query_cli```

There is a packaged set of scripts, the [query_cli](http://developer.tidepool.io/octopus/query_cli), that allows a user to


* login using an existing tidepool-account
```
tp_login <user_name>
```
* with that account query the data of accounts that you have
```
tp_query <userid_to_query> "smbg, cbg" "WHERE time > 2014-11-23T10:25:16"
```

There are many more features with this tool -- read the source if you're interested.
