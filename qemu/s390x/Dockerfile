FROM ubuntu:22.04

ENV RUNC_ROOT /runq/runc
ENV RUNQ_ROOT /var/lib/runq
ENV QEMU_ROOT /var/lib/runq/qemu

ENV DEBIAN_FRONTEND noninteractive
ENV GO_URL https://go.dev/dl/go1.22.3.linux-s390x.tar.gz
ENV GO_SHA256 2aba796417a69be5f3ed489076bac79c1c02b36e29422712f9f3bf51da9cf2d4
ENV PATH /usr/local/go/bin:$PATH

WORKDIR /runq

RUN echo "do_initrd = no" >> /etc/kernel-img.conf \
    && apt-get update \
    && apt-get install -y --no-install-recommends \
        build-essential \
        ca-certificates \
        cpio \
        e2fsprogs \
        git \
        libseccomp-dev \
        linux-image-virtual \
        linux-image-extra-virtual \
        pkg-config \
        qemu-system-s390x \
        rsync \
        wget \
        xz-utils

RUN set -eu; \
    wget -q -O go.tar.gz $GO_URL; \
    echo "$GO_SHA256 go.tar.gz" | sha256sum -c -; \
    tar -C /usr/local -xzf go.tar.gz; \
    rm -f go.tar.gz

RUN mkdir -p \
    $QEMU_ROOT/etc \
    $QEMU_ROOT/dev \
    $QEMU_ROOT/proc \
    $QEMU_ROOT/rootfs \
    $QEMU_ROOT/share \
    $QEMU_ROOT/sys


RUN    echo base   /lib/modules/*/kernel/fs/netfs/netfs.ko                                   >  $QEMU_ROOT/kernel.conf \
    && echo base   /lib/modules/*/kernel/fs/fscache/fscache.ko                               >> $QEMU_ROOT/kernel.conf \
    && echo base   /lib/modules/*/kernel/net/9p/9pnet.ko                                     >> $QEMU_ROOT/kernel.conf \
    && echo base   /lib/modules/*/kernel/fs/9p/9p.ko                                         >> $QEMU_ROOT/kernel.conf \
    && echo base   /lib/modules/*/kernel/net/9p/9pnet_virtio.ko                              >> $QEMU_ROOT/kernel.conf \
    && echo base   /lib/modules/*/kernel/net/core/failover.ko                                >> $QEMU_ROOT/kernel.conf \
    && echo base   /lib/modules/*/kernel/drivers/net/net_failover.ko                         >> $QEMU_ROOT/kernel.conf \
    && echo base   /lib/modules/*/kernel/drivers/block/virtio_blk.ko                         >> $QEMU_ROOT/kernel.conf \
    && echo base   /lib/modules/*/kernel/drivers/net/virtio_net.ko                           >> $QEMU_ROOT/kernel.conf \
    && echo vsock  /lib/modules/*/kernel/net/vmw_vsock/vsock.ko                              >> $QEMU_ROOT/kernel.conf \
    && echo vsock  /lib/modules/*/kernel/net/vmw_vsock/vmw_vsock_virtio_transport_common.ko  >> $QEMU_ROOT/kernel.conf \
    && echo vsock  /lib/modules/*/kernel/net/vmw_vsock/vmw_vsock_virtio_transport.ko         >> $QEMU_ROOT/kernel.conf \
    && echo btrfs  /lib/modules/*/kernel/arch/s390/crypto/crc32-vx_s390.ko                   >> $QEMU_ROOT/kernel.conf \
    && echo btrfs  /lib/modules/*/kernel/lib/libcrc32c.ko                                    >> $QEMU_ROOT/kernel.conf \
    && echo btrfs  /lib/modules/*/kernel/lib/raid6/raid6_pq.ko                               >> $QEMU_ROOT/kernel.conf \
    && echo btrfs  /lib/modules/*/kernel/lib/zstd/zstd_compress.ko                           >> $QEMU_ROOT/kernel.conf \
    && echo btrfs  /lib/modules/*/kernel/crypto/xor.ko                                       >> $QEMU_ROOT/kernel.conf \
    && echo btrfs  /lib/modules/*/kernel/fs/btrfs/btrfs.ko                                   >> $QEMU_ROOT/kernel.conf \
    && echo xfs    /lib/modules/*/kernel/arch/s390/crypto/crc32-vx_s390.ko                   >> $QEMU_ROOT/kernel.conf \
    && echo xfs    /lib/modules/*/kernel/lib/libcrc32c.ko                                    >> $QEMU_ROOT/kernel.conf \
    && echo xfs    /lib/modules/*/kernel/fs/xfs/xfs.ko                                       >> $QEMU_ROOT/kernel.conf \
    && echo zcrypt /lib/modules/*/kernel/drivers/s390/crypto/zcrypt.ko                       >> $QEMU_ROOT/kernel.conf \
    && echo zcrypt /lib/modules/*/kernel/drivers/s390/crypto/zcrypt_cex4.ko                  >> $QEMU_ROOT/kernel.conf \
    && echo z13    /lib/modules/*/kernel/drivers/char/hw_random/virtio-rng.ko                >> $QEMU_ROOT/kernel.conf \
    && echo z14+   /lib/modules/*/kernel/drivers/char/hw_random/s390-trng.ko                 >> $QEMU_ROOT/kernel.conf


RUN cp /boot/vmlinuz-*-generic $QEMU_ROOT/kernel

RUN cp -d --preserve=all --parents \
    /usr/lib/s390x-linux-gnu/* \
    $QEMU_ROOT/ 2>&1 | grep -v 'omitting directory';:

 RUN cp -a --parents \
     /usr/bin/qemu-system-s390x \
     /usr/bin/rsync \
     /usr/lib/ld64.so.1 \
     /usr/lib/modules \
     /usr/lib/s390x-linux-gnu/qemu \
     /usr/sbin/e2fsck \
     /usr/sbin/fsck.ext2 \
     /usr/sbin/fsck.ext4 \
     /usr/sbin/mke2fs \
     /usr/sbin/mkfs.ext2 \
     /usr/sbin/mkfs.ext4 \
     /usr/share/qemu \
     $QEMU_ROOT/

RUN cd $QEMU_ROOT \
    && ln -s ../proc/self/mounts etc/mtab \
    && ln -s usr/bin bin \
    && ln -s usr/lib64 lib64 \
    && ln -s usr/lib lib \
    && ln -s usr/sbin sbin

RUN    git config --global --add safe.directory /runq \
    && git config --global --add safe.directory /runq/runc
