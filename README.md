

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

 - Run: curl -fsSL https://raw.githubusercontent.com/plexcorp-pty-ltd/scriptables/main/autoprovision.sh | bash
 - Navigate to: http://127.0.0.1:3012/users/register

## Known issues

 - Multiple PHP versions are currently not supported on Ubuntu 23.04.
 - CSRF random expiry warning - just hard refresh the page if you see a "session expired" message.
 - Service workers currently not 100% implemented.

## Getting help

Please use the issue tracker to report bugs and new feature requests.

## Supporting the development of Scriptables

If you use Scriptables commercially or just love this awesome tool, please consider supporting us [here](https://store.plexscriptables.com/buy/09049952-97d6-4d36-83ad-b7fe01ad732f) . Every dollar is appreciated and will go a long way toward making Scriptables that much better.

