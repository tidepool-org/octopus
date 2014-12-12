octopus
=======

This is a data query API for Tidepool.

It was initially created by [anderspitman](https://github.com/anderspitman) upon request by Tidepool.

The purpose of this API is to provide the framework for queries against the Tidepool platform.

The first version of this API will be largely a fa&cedille;ade, designed to mimic the eventual API but with a tightly constrained set of functionality.

Here is what we expect to do for the first version:

## Query submission

To submit a query, POST to:

/query

where the body of the post is the query text. You must use a standard Tidepool authentication token in the headers.

The result will be 200 response with the MIME type of application/json, containing a JSON object with the results. If the query generates an empty set, the result will be 200 with an empty array. If the query fails to parse, the result will be 400.


## Supported Query Formats:

The following queries are written as if they were using [TQL](http://developer.tidepool.io/queries-and-notifications/), but they are not -- the query parser for this release expects these queries and no others, and only parts of these queries are recognized. Please don’t expect to experiment with the query language yet, as it won’t accept variation.

Query to get all of a user’s update records:

    METAQUERY
        WHERE userid IS "12d7bc90fa"

    QUERY
        TYPE IN update
        SORT BY time AS Timestamp REVERSED

Query to get a block of records in a given time range:

    METAQUERY
        WHERE userid IS "12d7bc90fa"

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

