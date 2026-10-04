#!/bin/bash
# exit-on-failure=yes
set -e

sudo apt-get update -y
sudo apt-get install -y postgresql postgresql-contrib

sudo systemctl enable --now postgresql
