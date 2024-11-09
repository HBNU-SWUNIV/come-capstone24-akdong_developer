# BusyBox 다운로드 및 압축 해제
cd /CarteTest/test

# BusyBox 다운로드
wget https://busybox.net/downloads/busybox-1.36.0.tar.bz2
tar -xvjf busybox-1.36.0.tar.bz2
cd /CarteTest/test/busybox-1.36.0
make distclean  # 이전 빌드 정리
make defconfig  # 기본 구성 설정
echo "CONFIG_IP=y" >> .config  # 'ip' 지원이 포함되도록 설정
make -j$(nproc)  # BusyBox 빌드


# 필요한 디렉토리 생성 및 파일 복사
mkdir -p /CarteTest/image/test_tt/bin
cp busybox /CarteTest/image/test_tt/bin/

# 기본 디렉토리 구조 생성
mkdir -p /CarteTest/image/test_tt/{bin,sbin,etc,proc,sys,usr/bin,usr/sbin,dev,tmp,lib,lib64}

# 패스워드 및 그룹 파일 생성
echo "root:x:0:0:root:/root:/bin/sh" > /CarteTest/image/test_tt/etc/passwd
echo "root:x:0:" > /CarteTest/image/test_tt/etc/group
echo "nameserver 8.8.8.8" > /CarteTest/image/test_tt/etc/resolv.conf

# 필요한 라이브러리 복사
# lib 디렉토리에 라이브러리 복사
cp /lib/x86_64-linux-gnu/libc.so.6 /CarteTest/image/test_tt/lib/
cp /lib/x86_64-linux-gnu/libm.so.6 /CarteTest/image/test_tt/lib/
cp /lib/x86_64-linux-gnu/libresolv.so.2 /CarteTest/image/test_tt/lib/

# lib64 디렉토리에 라이브러리 복사
mkdir -p /CarteTest/image/test_tt/lib64
cp /lib64/ld-linux-x86-64.so.2 /CarteTest/image/test_tt/lib64/

# 권한 설정
chmod +x /CarteTest/image/test_tt/bin/busybox

# /bin/sh를 busybox로 연결 (심볼릭 링크)
ln -s /CarteTest/image/test_tt/bin/busybox /CarteTest/image/test_tt/bin/sh

# /bin/sh, /bin/ps, /bin/ls를 busybox로 복사 (실제 파일 복사)
cp /CarteTest/image/test_tt/bin/busybox /CarteTest/image/test_tt/bin/ps
cp /CarteTest/image/test_tt/bin/busybox /CarteTest/image/test_tt/bin/ls
cp /CarteTest/image/test_tt/bin/busybox /CarteTest/image/test_tt/bin/ip
cp /CarteTest/image/test_tt/bin/busybox /CarteTest/image/test_tt/bin/sh

# 심볼릭 링크 생성
ln -s /CarteTest/image/test_tt/bin/busybox /CarteTest/image/test_tt/bin/ps
ln -s /CarteTest/image/test_tt/bin/busybox /CarteTest/image/test_tt/bin/ls
# ln -s /CarteTest/image/test_tt/bin/busybox /CarteTest/image/test_tt/bin/sh


# 이미지 압축
cd /CarteTest/image
tar -cvf test_tt.tar -C test_tt .

# Cleanup
cd /CarteTest/test || exit
rm -rf busybox-1.36.0 busybox-1.36.0.tar.bz2

echo "BusyBox image creation complete."

# cp /CarteTest/container/test_tt/bin/busybox /CarteTest/container/test_tt/bin/ip
