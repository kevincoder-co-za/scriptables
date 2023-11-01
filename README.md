

# Scriptables

Scriptables is an open source orchestration tool that takes away the pain of setting up and managing production servers. In just a few minutes you can build app servers, deploy code from GIT, manage your firewall, setup crons and more - all while using a friendly web interface.

While Scriptables is platform agnostic, we love PHP and offer full support for Laravel. This includes all the essential components necessary for running a production server; including MySQL, Nginx, Redis, Multiple PHP versions and more.


**Screenshot**:

![](https://plexscriptables.com/static/img/build-server.png)


## Dependencies

Scriptables is built using the GIN framework. A popular Golang framework for building web apis. Scriptables uses both redis and MySQL. We provide a convenient docker-compose file to run everything.

## Documentation & Installation

Detailed documentation and instructions on how to install can be found: https://scriptables.gitbook.io/

## Quick start

 - Clone this repository locally and navigate to the build folder.  Next
 -  Rename the "example.env" to ".env" . Change the settings accordingly (see docs above for help). Initially the most important setting to change is the encryption key.
 - Run:  `docker-compose up -d --build`
 - Navigate to : http://127.0.0.1:3001/users/register

## Known issues

 - Multiple PHP versions is currently not supported on Ubuntu 23.04.
 - CSRF random expiry warning - just hard refresh the page if you see a "session expired" message.
 - DB - need to move to app level instead of per request.
 - Service workers currently not 100% implemented.
 - Need to upgrade the template to use tailwind and remove bower.

## Getting help

Please use the issue tracker to report bugs and new feature requests. You can also visit us at https://plexscriptables.com


## Getting involved

We currently are working on a structured way to contribute. In the interim, simply branch of main and submit a pull request.

