FROM registry.fedoraproject.org/fedora-minimal:32

COPY src /src
COPY entrypoint.sh /entrypoint.sh

RUN microdnf update -y && \
    microdnf install -y dnf && \
    dnf install -y \
        wget \
        gcc \
        golang \
        libnbd-devel \
        nbdkit \
        nbdkit-devel \
        nbdkit-basic-plugins \
        nbdkit-xz-filter \
        && \
    mkdir /plugin && mv /src /plugin/src/ && (cd /plugin && GOPATH=$(pwd) go build -o /vddk-test-plugin.so -buildmode=c-shared vddk) && \
    wget https://download.fedoraproject.org/pub/fedora/linux/releases/32/Cloud/x86_64/images/Fedora-Cloud-Base-32-1.6.x86_64.raw.xz && \
    dnf remove -y \
        wget \
	gcc \
	golang \
	libnbd-devel \
	nbdkit \
	nbdkit-devel \
	nbdkit-basic-plugins \
	&& \
    dnf clean all && \
    microdnf remove dnf && \
    microdnf clean all

ENTRYPOINT /entrypoint.sh
