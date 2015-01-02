octopus
=======

[![Build Status](https://travis-ci.org/tidepool-org/octopus.png)](https://travis-ci.org/tidepool-org/octopus)

This is a data query API for Tidepool.

It was initially created by [anderspitman](https://github.com/anderspitman) upon request by Tidepool.

The purpose of this API is to provide the framework for queries against the Tidepool platform.

The first version of this API will be largely faked, designed to mimic the eventual API but with a tightly constrained set of functionality.

The version that was originally created was specified to do a simple query against the data stream to return the date and time of the last known data item for a given device. The next version will also include a constrained version of a general-purpose query API.

Authenticated calls must use a standard Tidepool authentication token in the headers. See the login function within [platform-client](https://github.com/tidepool-org/platform-client/blob/master/index.js) for details on how to log in and get one.

Here is what we expect to implement for the next version:

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

    POST /query

Requires authentication. The body of the post is the query text.

The result will be 200 response with the MIME type of application/json, containing a JSON object with the results. If the query generates an empty set, the result will be 200 with an empty array. If the query fails to parse, the result will be 400.


## Supported Query Formats:

The following queries are written as if they were using [TQL](http://developer.tidepool.io/queries-and-notifications/), but they are not -- the query parser for this release expects these queries and no others, and only parts of these queries are recognized. Please don’t expect to experiment with the query language yet, as it won’t accept variation.

Query to get all of a user’s update records:

    METAQUERY
        WHERE userid IS 12d7bc90fa

    QUERY
        TYPE IN update
        SORT BY time AS Timestamp REVERSED

Query to get a block of records in a given time range:

    METAQUERY
        WHERE userid IS 12d7bc90fa

    QUERY
        TYPE IN cbg, smbg, bolus, wizard
        WHERE time > starttime AND time < endtime
        SORT BY time AS Timestamp REVERSED

Result will be a JSON array with individual records corresponding to the selected types, reverse sorted by date (from newest to oldest).

The only acceptable `METAQUERY` is to query for a single userid. Aggregate metaqueries are not supported, and only the userids we give you will work.

Results are sorted by the time field. `REVERSED` can be specified or omitted to control the sort, but you cannot sort by anything other than time (the rest of the sort clause is ignored).

`TYPE IN` must be followed by a comma-separated list of types as defined in the [data formats documentation](http://developer.tidepool.io/data-model/v1/).

The `WHERE` clause for time must either be:

    WHERE time > starttime
or

    WHERE time > starttime AND time < endtime

Both starttime and endtime must be ISO 8601 timestamps referenced to UTC (example: `2014-12-11T04:44:16Z`).

Whitespace and upper/lower case are ignored; the formatting above makes it easier to read but it’s unimportant.

