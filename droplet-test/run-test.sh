#!/bin/bash
set -exo pipefail

REMOTE=$(cat remotehost)

scp ../zfs-cleaner "$REMOTE:/usr/local/bin"
scp *.conf "$REMOTE:"

echo "without ignore empty"

ssh "$REMOTE" zfs-cleaner plancheck cleaner_datastore_0-1.conf

echo "with ignore empty"

ssh "$REMOTE" zfs-cleaner plancheck --ignore-empty cleaner_datastore_0-1.conf

echo "all"

ssh "$REMOTE" zfs-cleaner plancheck all.conf
