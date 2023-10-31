#!/bin/bash
# exit-on-failure=yes

sudo apt install apt-transport-https ca-certificates curl software-properties-common -y
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" -y

sudo apt install docker-ce -y

sudo groupadd docker
sudo usermod -aG docker "#username#"