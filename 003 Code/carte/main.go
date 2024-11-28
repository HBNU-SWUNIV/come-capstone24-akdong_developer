package main

import (
    "carte/cmd"
)

func main() {
    cmd.Execute()
}


// echo "yourusername:100000:65536" | sudo tee -a /etc/subuid /etc/subgid

// sudo bash -c 'echo "cgroup2 /sys/fs/cgroup cgroup2 defaults 0 0" >> /etc/fstab'
// sudo mount -a

// sudo mkdir /sys/fs/cgroup/user.slice
// sudo chown $(whoami) /sys/fs/cgroup/user.slice


