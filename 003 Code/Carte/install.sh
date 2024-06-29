#!/bin/bash

# 'carte' 그룹 생성
if ! getent group carte > /dev/null 2>&1; then
    sudo groupadd carte
    echo "Group 'carte' created."
else
    echo "Group 'carte' already exists."
fi

# 현재 사용자를 'carte' 그룹에 추가
sudo usermod -aG carte $USER
echo "User '$USER' added to 'carte' group."

# 데몬 실행 파일 설치 (예: /usr/local/bin에 복사)
sudo cp carte-daemon /usr/local/bin/carte-daemon
sudo chmod +x /usr/local/bin/carte-daemon

# 데몬 서비스 설정 (systemd 사용)
sudo bash -c 'cat > /etc/systemd/system/carte.service <<EOF
[Unit]
Description=Carte Daemon
After=network.target

[Service]
ExecStart=/usr/local/bin/carte-daemon
Restart=always
User=root
Group=carte

[Install]
WantedBy=multi-user.target
EOF'

# 서비스 시작 및 부팅 시 자동 시작 설정
sudo systemctl daemon-reload
sudo systemctl start carte
sudo systemctl enable carte

echo "Carte daemon installed and started."
