#!/bin/bash
# exit-on-failure=yes

sudo -u #SITE_SLUG# bash -c "ssh-keyscan github.com gitlab.com bitbucket.org >> #USER_DIRECTORY#/.ssh/known_hosts" || true
