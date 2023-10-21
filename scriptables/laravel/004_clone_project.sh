#!/bin/bash
# exit-on-failure=yes
set +e

cd #USER_DIRECTORY#

sudo -u #SITE_SLUG# /bin/bash -c eval `ssh-agent`
sudo -u #SITE_SLUG# /bin/bash -c ssh-add #KEY_PATH#

echo "Cloning project..."
errorMessage=""
if [ -d "#USER_DIRECTORY#/#SITE_SLUG#" ]; then
    echo "Directory exists - doing just a pull."
    errorMessage=$(cd #USER_DIRECTORY#/#SITE_SLUG# && sudo -u #SITE_SLUG# GIT_SSH_COMMAND="ssh -o StrictHostKeyChecking=no -i #KEY_PATH#" git pull 2>&1)
else
    errorMessage=$(cd #USER_DIRECTORY# && sudo -u #SITE_SLUG# GIT_SSH_COMMAND="ssh -o StrictHostKeyChecking=no -i #KEY_PATH#" git clone #GIT_URL# #SITE_SLUG# 2>&1)
fi


if [ ! -d "#USER_DIRECTORY#/#SITE_SLUG#" ]; then
    set -e
    echo "Failed to clone the project: $errorMessage. Please check that you have added the deploy key. If the problem persists, try to ssh into the server and run: sudo -u #SITE_SLUG# GIT_SSH_COMMAND=\"ssh -o StrictHostKeyChecking=no -i #KEY_PATH#\" git clone #GIT_URL# #SITE_NAME# "
    exit 1
fi

cd #USER_DIRECTORY#/#SITE_SLUG#

echo "Checking out #BRANCH# branch"
sudo -u #SITE_SLUG# git checkout #BRANCH#

if [ -f ".env" ]; then
    sudo rm ".env"
fi