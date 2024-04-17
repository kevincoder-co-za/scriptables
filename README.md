

# Scriptables

Scriptables is an open-source orchestration tool that takes away the pain of setting up and managing production servers. In just a few minutes you can build app servers, deploy code from GIT, manage your firewall, set up CRONs, and more - all while using a friendly web interface.

While Scriptables is platform agnostic, we love PHP and offer full support for Laravel. This includes all the essential components necessary for running a production server; including MySQL, Nginx, Redis, Multiple PHP versions, and more.


**Screenshot**:

![](https://plexscriptables.com/static/img/build-server.png)

## Dependencies

Scriptables are built using the GIN framework. A popular Golang framework for building web APIs. Scriptables uses both Redis and MySQL. We provide a convenient docker-compose file to run everything.

## Documentation & Installation

Detailed documentation and instructions on how to install can be found: https://scriptables.gitbook.io/

## Quick start

 - Clone this repository locally.
 -  Rename the "example.env" to ".env". Change the settings accordingly (see docs above for help). Initially, the most important setting to change is the encryption key.
 - Run:  `docker-compose up -d --build`
 - Navigate to: http://127.0.0.1:3001/users/register

## Known issues

 - Multiple PHP versions are currently not supported on Ubuntu 23.04.
 - CSRF random expiry warning - just hard refresh the page if you see a "session expired" message.
 - DB - needs to move to app level instead of per request.
 - Service workers currently not 100% implemented.
 - Need to upgrade the template to use tailwind and remove bower.

## Getting help

Please use the issue tracker to report bugs and new feature requests.

## Supporting the development of Scriptables

If you use Scriptables commercially or just love this awesome tool, please consider supporting us here: https://store.plexscriptables.com/checkout . Every dollar is appreciated and will go a long way toward making Scriptables that much better.

