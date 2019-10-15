# dedup

## Simple, but fast and multithreaded deduplication system

## Either prints duplicates or removes them

## Download - get it [here](https://github.com/bestmethod/dedup/releases)

```
$ mv dedup-osx /usr/local/bin/dedup
$ chmod 755 /usr/local/bin/dedup
$ dedup -help
```

## Usage

```
$ dedup -help
Usage:
  dedup [-detail] [-rm|-rm-all] [-quiet] [-dryrun] directory
  dedup [-detail] [-rm|-rm-all] [-quiet] [-dryrun] directory1 directory2 directory...

Optional Arguments:
  -detail
        if set, will print each file it's processing
  -dryrun
        if set, doesn't perform actual remove actions
  -multithread
        if set, enable multithreading - one thread per listed directory; will sort by specified directory order, then alphabetically
  -quiet
        if set, will not print duplicate information
  -rm
        if set, will remove all duplicates except the first one for each file
  -rm-all
        if set, will remove all files that have duplicates (all copies)
  -sort
        employs a file sorting mechanism like multithreading: by specified directory order, then alphabetically; useful when with -rm option (first file remains)
```

## Example output

```
$ ./dedup -rm -multithread 1 1 2 3 4
Generating sums, dryrun=false
Generating sums on 1
Generating sums on 1
Generating sums on 2
Generating sums on 3
Generating sums on 4
Enumerating table and printing duplicates
DUPLICATE: size+sha=503+ea836217d793dd6c158d141a7ea8d0bd824ccc68e7914bc7ddf7c9ae5a16e1a4
                inode=16777220+3391255  name=a.json     path=1/a.json
        remove  inode=16777220+3391255  name=launch.json        path=1/launch.json
        remove  inode=16777220+3391255  name=b.json     path=1/b.json
        remove  inode=16777220+3391256  name=launch.json        path=2/launch.json
        remove  inode=16777220+3392380  name=launch.json        path=3/launch.json
        remove  inode=16777220+3399297  name=b.json     path=4/11/b.json
        remove  inode=16777220+3399296  name=launch.json        path=4/11/launch.json
        remove  inode=16777220+3399295  name=a.json     path=4/11/a.json
        remove  inode=16777220+3392382  name=launch.json        path=4/launch.json
```
