#!/usr/bin/env bash

mkdir -p /opt/testing
cp -f /vddk-test-plugin.so /opt/testing/libvddk-test-plugin.so
cp -f /lib64/nbdkit/filters/nbdkit-xz-filter.so /opt/testing/nbdkit-xz-filter.so
mv /Fedora-Cloud-Base-32-1.6.x86_64.raw.xz /opt/testing/nbdtest.xz
