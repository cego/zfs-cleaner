# -*- mode: ruby -*-
# vi: set ft=ruby :


Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/bionic64"

  config.vm.provision "shell", inline: <<-SHELL
    locale-gen en_GB.UTF-8
    add-apt-repository ppa:longsleep/golang-backports
    apt-get update
    apt-get install -y zfsutils-linux zfs-initramfs

    mkdir -p /zfsmnt/

    dd if=/dev/zero of=/zfsmnt/disk0 bs=1M count=128
    dd if=/dev/zero of=/zfsmnt/disk1 bs=1M count=128
    dd if=/dev/zero of=/zfsmnt/disk2 bs=1M count=128
    losetup /dev/loop0 /zfsmnt/disk0
    losetup /dev/loop1 /zfsmnt/disk1
    losetup /dev/loop2 /zfsmnt/disk2

    zpool create -f datastore0 raidz /dev/loop0 /dev/loop1 /dev/loop2

    dd if=/dev/zero of=/zfsmnt/disk3 bs=1M count=128
    dd if=/dev/zero of=/zfsmnt/disk4 bs=1M count=128
    losetup /dev/loop3 /zfsmnt/disk3
    losetup /dev/loop4 /zfsmnt/disk4

    zpool create -f datastore1 raidz /dev/loop3 /dev/loop4

    dd if=/dev/zero of=/zfsmnt/disk5 bs=1M count=128
    dd if=/dev/zero of=/zfsmnt/disk6 bs=1M count=128
    dd if=/dev/zero of=/zfsmnt/disk7 bs=1M count=128
    losetup /dev/loop5 /zfsmnt/disk5
    losetup /dev/loop6 /zfsmnt/disk6
    losetup /dev/loop7 /zfsmnt/disk7

    zpool create -f datastore2 raidz /dev/loop5 /dev/loop6 /dev/loop7

    cat << EOF > plan
plan planA {
  path datastore0
  path datastore1

  keep latest 2

  keep 1m for 2h
  keep 1h for 2d
}

plan planB {
  path datastore2

  keep latest 10
}

EOF

  echo 'You can now run "GOOS=linux go build && GOOS=linux go test -c" from the host'
  echo 'and run "/vagrant/zfs-cleaner plan" in the box'
  SHELL
end
