#!/bin/bash

# start snapd
sudo service snapd start

# use snap to install multipass
sudo snap install multipass

# start a mutlipass vm
multipass launch --name uberbase --cpus 2 --mem 4G --disk 20G --cloud-init ./openfass/cloud-config.yaml
