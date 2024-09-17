#!/bin/sh

# Add user namespace mappings
echo "$(whoami):100000:65536" | sudo tee -a /etc/subuid /etc/subgid

# Enable unprivileged user namespaces
sudo sysctl -w kernel.unprivileged_userns_clone=1

echo "$(whoami):100000:65536"
echo "kernel.unprivileged_userns_clone = 1"
echo "Init setup completed."

