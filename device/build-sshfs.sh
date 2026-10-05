set -e
apk add -q build-base meson ninja pkgconf fuse3-dev fuse3-static glib-dev glib-static gettext-static pcre2-dev zlib-static libffi-dev curl >/dev/null
cd /tmp && curl -sL https://github.com/libfuse/sshfs/releases/download/sshfs-3.7.3/sshfs-3.7.3.tar.xz | tar xJ
cd sshfs-3.7.3 && meson setup build --prefer-static --default-library=static -Dc_link_args=-static >/dev/null && ninja -C build 2>&1 | tail -n 15
strip build/sshfs && cp build/sshfs /out/sshfs && file /out/sshfs 2>/dev/null || ls -la /out/sshfs
