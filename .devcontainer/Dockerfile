FROM node:22.23.1-bookworm-slim AS node

FROM golang:1.26.5-bookworm

ARG USER_ID=1000
ARG GROUP_ID=1000

COPY --from=node /usr/local/ /usr/local/

RUN apt-get update \
    && apt-get install --yes --no-install-recommends \
        ca-certificates \
        fonts-liberation \
        fonts-noto-color-emoji \
        libasound2 \
        libatk-bridge2.0-0 \
        libatk1.0-0 \
        libcups2 \
        libdbus-1-3 \
        libdrm2 \
        libgbm1 \
        libglib2.0-0 \
        libnspr4 \
        libnss3 \
        libpango-1.0-0 \
        libx11-6 \
        libxcb1 \
        libxcomposite1 \
        libxdamage1 \
        libxext6 \
        libxfixes3 \
        libxkbcommon0 \
        libxrandr2 \
    && rm -rf /var/lib/apt/lists/* \
    && groupadd --gid "${GROUP_ID}" auditor \
    && useradd --uid "${USER_ID}" --gid "${GROUP_ID}" --create-home --shell /bin/bash auditor \
    && rm -f /usr/local/bin/yarn /usr/local/bin/yarnpkg /usr/local/bin/pnpm /usr/local/bin/pnpx \
    && corepack enable pnpm \
    && mkdir -p /workspace /home/auditor/.cache/go-build /home/auditor/.local/share/pnpm/store \
    && chown -R auditor:auditor /workspace /home/auditor

ENV PATH="/usr/local/go/bin:/home/auditor/go/bin:/home/auditor/.local/share/pnpm:${PATH}"
ENV CGO_ENABLED=0
ENV PLAYWRIGHT_BROWSERS_PATH=/workspace/.cache/ms-playwright

USER auditor
WORKDIR /workspace

CMD ["sleep", "infinity"]
