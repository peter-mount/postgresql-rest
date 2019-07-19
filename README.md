# postgresql-rest

postgresql-rest is a standalone application that allows PostgreSQL functions in one or more databases to be
exposed as a REST API.

## History

It started off being part of the nre-feeds project when it was noticed it should be more generic for exposing other
opendata than just the NRE data feeds. I then started to include it within the uktransport project but it's turned out
to be deserving it's own project.

## Overview

The utility currently has 3 modes of operation:
1. Exposing functions over a rest api
1. Invoking functions with no parameters on a Cron schedule
1. Invoking functions when a message is received from a RabbitMQ queue

You can run the utility with any or all of these modes at the same time - although if you are receiving from RabbitMQ
as well as hosting REST I suggest you run them in two separate instances.

Full documentation will appear within the project's wiki.
