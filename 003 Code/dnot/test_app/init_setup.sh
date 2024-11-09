#!/bin/bash

# 사용자 ID와 그룹 ID를 설정합니다.
echo "$(whoami):100000:65536" | sudo tee -a /etc/subuid /etc/subgid

# 사용자 네임스페이스를 활성화합니다.
sudo sysctl -w kernel.unprivileged_userns_clone=1

# cgroup2 설정
sudo mkdir -p /sys/fs/cgroup/user.slice
sudo chown $(whoami) /sys/fs/cgroup/user.slice

# cgroup2 마운트
if ! mountpoint -q /sys/fs/cgroup; then
    sudo mount -t cgroup2 none /sys/fs/cgroup
fi

echo "Init setup completed."


