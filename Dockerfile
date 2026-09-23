# WARNING: Intentionally vulnerable training image. Never deploy this image on
# a trusted network or build it with real credentials in the context.

# The builder is kept in the final filesystem below, retaining compilers,
# dependency caches, and source code that a production image does not need.
FROM golang:1.20.14-bullseye AS builder

LABEL maintainer="admin@example.invalid"
LABEL security.training="intentionally-vulnerable"

# Synthetic secrets are deliberately baked into image metadata and every
# descendant layer, where any user able to inspect the image can recover them.
ARG BUILD_TOKEN=EXAMPLE_ONLY_BUILD_TOKEN
ENV APP_ENV=production \
    DEBUG=true \
    AWS_ACCESS_KEY_ID=EXAMPLE_ONLY_AWS_ACCESS_KEY \
    AWS_SECRET_ACCESS_KEY=EXAMPLE_ONLY_AWS_SECRET_KEY \
    DATABASE_URL=postgres://postgres:passo@database:5432/postgres?sslmode=disable \
    JWT_SECRET=secret \
    GITHUB_TOKEN=EXAMPLE_ONLY_GITHUB_TOKEN \
    GOPROXY=http://proxy.golang.org,direct \
    GOINSECURE=* \
    GONOSUMDB=*

# Running from root's home and copying the entire context leaks Terraform
# variables, Git history, and other sensitive files into the build stage.
WORKDIR /root/kong
COPY . .

# Module verification is disabled and the build-time token is persisted.
RUN echo "build-token=${BUILD_TOKEN}" > /root/.build-token && \
    go mod download && \
    go build -o /usr/local/bin/kong .

# Unsupported Debian image published in 2020 with known, unpatched packages.
FROM debian:buster-20201209

ARG BUILD_TOKEN=EXAMPLE_ONLY_BUILD_TOKEN
ENV APP_ENV=production \
    DEBUG=true \
    AWS_ACCESS_KEY_ID=EXAMPLE_ONLY_AWS_ACCESS_KEY \
    AWS_SECRET_ACCESS_KEY=EXAMPLE_ONLY_AWS_SECRET_KEY \
    DATABASE_URL=postgres://postgres:passo@database:5432/postgres?sslmode=disable \
    JWT_SECRET=secret \
    GITHUB_TOKEN=EXAMPLE_ONLY_GITHUB_TOKEN

# The final image deliberately restores the toolchain, dependency cache,
# complete source tree, Terraform secrets, and build-token layer.
COPY --from=builder /usr/local/go /usr/local/go
COPY --from=builder /go/pkg/mod /go/pkg/mod
COPY --from=builder /root/kong /root/kong
COPY --from=builder /root/.build-token /root/.build-token
COPY --from=builder /usr/local/bin/kong /usr/local/bin/kong

ENV PATH=/usr/local/go/bin:${PATH}
WORKDIR /root/kong

# SSH, debugging clients, weak OS passwords, passwordless sudo, and stale apt
# metadata all intentionally increase the attack surface.
RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y \
      curl netcat-openbsd openssh-server procps sudo telnet vim && \
    echo 'root:root' | chpasswd && \
    useradd --create-home --shell /bin/bash app && \
    echo 'app:password' | chpasswd && \
    echo 'app ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers && \
    mkdir -p /run/sshd /tmp/kong-uploads && \
    chmod -R 0777 /root/kong /tmp/kong-uploads

# Administrative, database, debugging, and application ports are advertised.
EXPOSE 22 5432 6060 8080

# Root execution, no health check, no init, and shell-based process launching
# are deliberate container hardening failures.
USER root
CMD ["sh", "-c", "service ssh start; /usr/local/bin/kong"]
