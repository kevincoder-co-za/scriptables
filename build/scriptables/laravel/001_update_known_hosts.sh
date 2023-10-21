#!/bin/bash
# exit-on-failure=yes
set +e

eval `ssh-agent`

sudo -u #SITE_SLUG# ssh-keyscan github.com >> ~/.ssh/known_hosts
sudo -u #SITE_SLUG# ssh-keyscan gitlab.com >> ~/.ssh/known_hosts
sudo -u #SITE_SLUG# ssh-keyscan gitlab.org >> ~/.ssh/known_hosts
sudo -u #SITE_SLUG# ssh-keyscan bitbucket.org >> ~/.ssh/known_hosts