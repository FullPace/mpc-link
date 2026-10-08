# Programs MPC Link puts on the MPC

`sshfs` is [SSHFS](https://github.com/libfuse/sshfs) 3.7.3, unmodified, built for 32-bit ARM (armv7) and statically
linked with libfuse 3 and GLib (Alpine Linux, musl) by [`build-sshfs.sh`](build-sshfs.sh):

```sh
docker run --rm --platform linux/arm/v7 -v "$PWD":/out arm32v7/alpine:3.20 sh /out/build-sshfs.sh
```

SSHFS and libfuse are licensed under the GNU General Public License, version 2 (libfuse's library part under the LGPL
2.1); GLib under the LGPL 2.1. Their source code: https://github.com/libfuse/sshfs/releases/tag/sshfs-3.7.3,
https://github.com/libfuse/libfuse and https://gitlab.gnome.org/GNOME/glib (the versions Alpine 3.20 ships). MPC Link
copies this binary to the MPC as is, as a separate program (`/data/hacks/link/sshfs`); MPC Link's own code is MIT.
