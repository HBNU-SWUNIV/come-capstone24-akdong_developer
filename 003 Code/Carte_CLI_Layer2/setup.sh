#!/bin/bash

echo "Setting up environment..."

# Update package lists and install necessary packages
apt-get update && apt-get install -y golang mysql-server mongodb

# Setting up Go environment
mkdir -p ~/go/bin
echo "export GOPATH=$HOME/go" >> ~/.bashrc
echo "export PATH=$PATH:$GOPATH/bin" >> ~/.bashrc
source ~/.bashrc

# Setting up MySQL
systemctl start mysql
systemctl enable mysql
mysql_secure_installation <<EOF

y
y
y
y
EOF

# Setting up MongoDB
systemctl start mongodb
systemctl enable mongodb

# Verify installations
go version
mysql --version
mongo --version

echo "Setup complete."
