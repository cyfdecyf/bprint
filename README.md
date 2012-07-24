# Overview

hexdump's format string is hard to use IMO, and the shipped version on Debian 6
and OS X doesn't support 64-bit integer. So I created this little tool in Go to
replace hexdump for my own need.

Limitations:

- Does not support floating point number and string
- Does not have special print specifier to control where to print offset, etc.
- Others features that I do not use

# Usage

In `bprint`, the binary data specification and how the they are going to be
printed is specified separately using the `-e` and `-p` option.

- `-e` specifies binary field. Using the same syntax as Ruby's `Array.unpack`.
  - `c`, `s`, `l`, `q` stands for signed 8,16,32,64-bit integer
  - `C`, `S`, `L`, `Q` stands for unsigned 8,16,32,64-bit integer
  - A number following the type specifier repeats that specifier. For example, `c4` is equivalent to `cccc`
- `-p` specifies how to print the binary data. It uses C printf style field specifier
  - `%c`, `%d`, `%x`, `%o` are supported, size and signess information is implicit from the binary field information
  - **if not specified, defaults to `%02x` for each binary field**
  - field can be followed by an optional seperator and count to repeat. For example, `%2d-3#` is eqivalent to `%2d-%2d-%2d`
- `-o` print offset at the left most column
- `-c` print how many record has been read (right after offset column)
- `--version` print version information

# Example

Suppose each record in the binary file contains 2 byte and 1 64-bit integer, here's invocation to print it's content:

    bprint -e 'c2q' -p '%x %d2#' bindata