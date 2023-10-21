#!/bin/bash

sudo mkdir -p #USER_DIRECTORY#/.ssh/

eval `ssh-agent`

if ! id "#SITE_SLUG#" &>/dev/null; then
  sudo  useradd #SITE_SLUG# --shell /bin/bash
fi

sudo chown -R #SITE_SLUG#:#SITE_SLUG# #USER_DIRECTORY#
sudo chmod  0700 #USER_DIRECTORY#/.ssh/

sudo -u #SITE_SLUG# ssh-keygen -b 2048 -t rsa -f #KEY_PATH# -q -N ""