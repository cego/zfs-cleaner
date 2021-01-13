#!/bin/bash
set -x pipefail

REMOTE=$(cat remotehost)

ssh "$REMOTE" zpool destroy datastore0
ssh "$REMOTE" zpool destroy datastore1
ssh "$REMOTE" zpool destroy datastore2

ssh "$REMOTE" losetup -v -d /dev/loop0
ssh "$REMOTE" losetup -v -d /dev/loop1
ssh "$REMOTE" losetup -v -d /dev/loop2
ssh "$REMOTE" losetup -v -d /dev/loop3
ssh "$REMOTE" losetup -v -d /dev/loop4
ssh "$REMOTE" losetup -v -d /dev/loop5
ssh "$REMOTE" losetup -v -d /dev/loop6
ssh "$REMOTE" losetup -v -d /dev/loop7

ssh "$REMOTE" rm -rf /zfsmnt

ssh "$REMOTE" rm -f /root/*.conf
ssh "$REMOTE" rm -f /root/*.protect
ssh "$REMOTE" rm -f /usr/local/bin/zfs-cleaner
