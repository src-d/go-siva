
## Specification

This is the specification of the siva format version 1.

A siva file is composed of a sequence of one or more blocks. Blocks are just
concatenated without any additional delimiter.

```
block 1
...
[block n]
```

### Block

Each block has the following structure:

```
[file content 1]
...
[file content n]
index
```

The content of each file is concatenated without any delimiter. After the last
file content, there is an index of the block.

It is possible to have a block with no file contents at all. That is the case
for a block that deletes a file. In any case, the index must be present.

### Index

The index has the following structure:

```
signature
version
[index entry 1]
...
[index entry n]
[index footer]
```

The `signature` field is a sequence of 3 bytes (Go implementation use uint8 for this. Go byte is an alias for uint8 type) with the value `IBA`. If the
signature does not match this sequence, it is considered an error.

The `version` field is an uint8 with the value `1`. If the version contains an
unknown value, the implementation is not expected to be able to read the file
at all.

Each index entry has the following fields:

* Byte length of the entry name (uint32).
* Entry name (UTF-8 string in UNIX format).
* UNIX mode (uint32), [see below](#unix-mode-format).
* Modification time as UNIX time in nanoseconds (int64).
* Offset of the file content, relative to the beginning of the block (uint64).
* Size of the file content (uint64).
* CRC32 (uint32) (Integrity of the file content this entry points to).
* Flags (uint32), supported flags: 0x0 (no flags), 0x1 (deleted).

The index footer consists of:

* Number of entries in the block (uint32).
* Index size in bytes (uint64).
* Block size in bytes (uint64).
* CRC32 (uint32) (Integrity of: Signature + Version + Entries).

All integers are encoded as big endian.

### Unix Mode Format

The UNIX mode field has the following format:

```
     3 3 2 2 2 2 2 2 2 2 2 2 1 1 1 1 1 1 1 1 1 1
bit  1 0 9 8 7 6 5 4 3 2 1 0 9 8 7 6 5 4 3 2 1 0 9 8 7 6 5 4 3 2 1 0
  M +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+ L
  S |d|a|l|T|L|D|p|S|u|g|c|t|?| - - - - - - - - - |r|w|x|r|w|x|r|w|x| S
  B +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+ B
     @ _ _ _ @ @ @ @ _ _ @ _ @  ← file type bits  |owner|group|other|
     | | | | | | | | | | | | | 	                     access perms
     | | | | | | | | | | | | |- non-regular file
     | | | | | | | | | | | \--- sticky
     | | | | | | | | | | \----- character device (when D is set)
     | | | | | | | | | \------- setgid
     | | | | | | | | \--------- setuid
     | | | | | | | \----------- Unix-domain socket
     | | | | | | \------------- named pipe (FIFO)
     | | | | | \--------------- device file
     | | | | \----------------- symbolic link
     | | | \------------------- temporary file (Plan 9 only)
     | | \--------------------- exclusive use
     | \----------------------- append-only
     \------------------------- directory
```

All the bits not otherwise labelled are reserved for future use.

**Note:** This layout is defined by the [Go `os` package](https://godoc.org/os#FileMode)
and is _not_ the same layout as POSIX. The same layout is used on all
systems. See [issue #11](https://github.com/src-d/go-siva/issues/11) for
context.

## Limitations

The following limits apply to the format as of version 1:

* File name length: 2<sup>32</sup>-1 bytes.
* Number of blocks: no limit.
* Number of entries per block: 2<sup>32</sup>-1
* Number of total entries: no limit (reference implementation does not support more than 2<sup>63</sup>-1).
* Block index size: 2<sup>64</sup>-1 bytes.
* Block size: 2<sup>64</sup>-1 bytes.
* File entry size: 2<sup>64</sup>-1 bytes.
* Archive file size: no limit.
