#!/bin/bash
# exit-on-failure=yes
set -e

curl -fsSL https://deb.nodesource.com/setup_lts.x | sudo -E bash -
sudo apt-get install -y nodejs

node --version
npm --version
