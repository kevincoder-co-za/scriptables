#!/bin/bash
# exit-on-failure=yes
set -e

sudo apt-get update -y
sudo apt-get install -y supervisor

sudo systemctl enable --now supervisor
