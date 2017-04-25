# WARNING
This is alpha quality software and will probably destroy all your ZFS pools
and kill your dog. Use with caution.

# zfs-cleaner

zfs-cleaner is a simple tool for destroying ZFS snapshots after predefined
retention periods. It aims to do one thing only - and do it well.

## Installation

If you have a working Go environment, it should be as simple as:

    go install github.com/cego/zfs-cleaner

### Installation from pre-built package

Binary releases are available here: https://github.com/cego/zfs-cleaner/releases

## Common usage scenario

zfs-cleaner was developed to assist in an environment where ZFS snapshots is
used as backup. Snapshots is transferred to remote hosts. Different
retention policies must be applied to senders and receivers.

A common pattern used at CEGO is having application servers perform ZFS
snapshots often, and use `zfs send` to transfer snapshots to a remote backup
host.

On the application server itself it is often desired to only keep a few
snapshots for space saving. On the receiving backup host, it is often preferred
to have longer retention. zfs-cleaner should solve both cases.

## Configuration

zfs-cleaner is configured through a configuration file resembling the nginx
configuration format without semicolons.

A configuration consists of one of more *plans*. A plan defines how to clean
one or more ZFS datasets.

A plan is defined like this:

    plan planA {
        path pool/dataset1
        path pool/dataset2

		keep latest 2

		keep 1m for 2h
		keep 1h for 2d
	}

    plan planB {
        path pool/dataset3

		keep latest 10
	}

    plan planC {
        path pool/dataset4

		keep 0s for 1h
	}

*planA* will keep snapshots one minute apart for two hours and one hour apart
for two days. This will be applied to the dataset *pool/dataset1* and
*pool/dataset2*. Besides that the latest two snapshots will be kept.

*planB* will keep the ten latest snapshots from *pool/dataset3*.

*planC* will keep all snapshots for an hour.

Path must refer to one of the results from `sudo zfs list -t filesystem -o name`.

### Units

All periods consits of a positive integer and a unit. A special case is `0s` which when used as a frequency means "everything". The Following units are supported:

| Unit | Description    |
|------|----------------|
| `s`  | Second         |
| `m`  | Minute         |
| `h`  | Hour           |
| `d`  | Day (24 hours) |
| `y`  | Year           |

### Command line arguments

zfs-cleaner has a single mandatory argument. The path to a configuration file.

A few flags are provided as well.

| Short | Long        | Description                                  |  
|-------|-------------|----------------------------------------------|
| `-n`  | `--dryrun`  | Do nothing, print what could have been done. |
| `-v`  | `--verbose` | Do everything, print what's done.            |
