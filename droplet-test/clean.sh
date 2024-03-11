#!/bin/bash
set -x pipefail

REMOTE=$(cat remotehost)

ssh "$REMOTE" zpool destroy datastore0
ssh "$REMOTE" zpool destroy datastore1
ssh "$REMOTE" zpool destroy datastore2

ssh "$REMOTE" losetup -v -d /dev/loop10
ssh "$REMOTE" losetup -v -d /dev/loop11
ssh "$REMOTE" losetup -v -d /dev/loop12
ssh "$REMOTE" losetup -v -d /dev/loop13
ssh "$REMOTE" losetup -v -d /dev/loop14
ssh "$REMOTE" losetup -v -d /dev/loop15
ssh "$REMOTE" losetup -v -d /dev/loop16
ssh "$REMOTE" losetup -v -d /dev/loop17

ssh "$REMOTE" rm -rf /zfsmnt

ssh "$REMOTE" rm -f /root/*.conf
ssh "$REMOTE" rm -f /root/*.protect
ssh "$REMOTE" rm -f /usr/local/bin/zfs-cleaner
