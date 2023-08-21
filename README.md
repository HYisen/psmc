# PSMC

A [ps_mem](https://github.com/pixelb/ps_mem) client that provides interfaces to monitor process memory usage on Linux.

**Deprecated** as suggests to use [procfs](https://pkg.go.dev/github.com/prometheus/procfs) instead.

## Story

It's designed to provide abilities to get process information for other programs as a library.

I Google it as "arch linux get memory usage of PID" and got ps_mem from
[this](https://www.2daygeek.com/linux-commands-check-memory-usage/)
as a per process one-shot(i.e. not interactive) solution.

The original idea includes to play as a CLI wrapper of ps_mem.

There is already a golang version [psm](https://pkg.go.dev/github.com/bpowers/psm),
but it's a CLI program rather than a library API.

Once I checked its source code, I decided to remaster it in code level rather than create a parser as it's quite simple.

Having finished most of the structure, I think it's better to rename to something contains `/proc` as basically it's
a specific linux file reader and parser. And then I found [procfs](https://pkg.go.dev/github.com/prometheus/procfs),
which is actually a better version (as generic for various platforms) of what I have done.

Name really matters.

## Usage

The real interface is not exposed. Check main function to see how to use it and expose the API as you want. 