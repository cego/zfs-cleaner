#!/bin/bash
set -exo pipefail

REMOTE=$(cat remotehost)

ssh "$REMOTE" mkdir -p /zfsmnt/

ssh "$REMOTE" dd if=/dev/zero of=/zfsmnt/disk0 bs=1M count=128
ssh "$REMOTE" dd if=/dev/zero of=/zfsmnt/disk1 bs=1M count=128
ssh "$REMOTE" dd if=/dev/zero of=/zfsmnt/disk2 bs=1M count=128

ssh "$REMOTE" losetup /dev/loop0 /zfsmnt/disk0
ssh "$REMOTE" losetup /dev/loop1 /zfsmnt/disk1
ssh "$REMOTE" losetup /dev/loop2 /zfsmnt/disk2

ssh "$REMOTE" zpool create -f datastore0 raidz /dev/loop0 /dev/loop1 /dev/loop2

ssh "$REMOTE" zfs snapshot datastore0@0
ssh "$REMOTE" zfs snapshot datastore0@1
ssh "$REMOTE" zfs snapshot datastore0@2

ssh "$REMOTE" dd if=/dev/zero of=/zfsmnt/disk3 bs=1M count=128
ssh "$REMOTE" dd if=/dev/zero of=/zfsmnt/disk4 bs=1M count=128

ssh "$REMOTE" losetup /dev/loop3 /zfsmnt/disk3
ssh "$REMOTE" losetup /dev/loop4 /zfsmnt/disk4

ssh "$REMOTE" zpool create -f datastore1 raidz /dev/loop3 /dev/loop4

ssh "$REMOTE" zfs snapshot datastore1@0
ssh "$REMOTE" zfs snapshot datastore1@1

ssh "$REMOTE" bash -c "echo @1 > /root/1.protect"
ssh "$REMOTE" bash -c "echo @1 > /root/2.protect"

ssh "$REMOTE" dd if=/dev/zero of=/zfsmnt/disk5 bs=1M count=128
ssh "$REMOTE" dd if=/dev/zero of=/zfsmnt/disk6 bs=1M count=128
ssh "$REMOTE" dd if=/dev/zero of=/zfsmnt/disk7 bs=1M count=128

ssh "$REMOTE" losetup /dev/loop5 /zfsmnt/disk5
ssh "$REMOTE" losetup /dev/loop6 /zfsmnt/disk6
ssh "$REMOTE" losetup /dev/loop7 /zfsmnt/disk7

ssh "$REMOTE" zpool create -f datastore2 raidz /dev/loop5 /dev/loop6 /dev/loop7
