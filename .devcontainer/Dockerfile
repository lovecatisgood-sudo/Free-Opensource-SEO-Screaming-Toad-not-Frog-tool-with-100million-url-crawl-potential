FROM node:22.23.1-bookworm-slim AS node

FROM golang:1.26.5-bookworm

ARG USER_ID=1000
ARG GROUP_ID=1000

COPY --from=node /usr/local/ /usr/local/

RUN groupadd --gid "${GROUP_ID}" auditor \
    && useradd --uid "${USER_ID}" --gid "${GROUP_ID}" --create-home --shell /bin/bash auditor \
    && rm -f /usr/local/bin/yarn /usr/local/bin/yarnpkg /usr/local/bin/pnpm /usr/local/bin/pnpx \
    && corepack enable pnpm \
    && mkdir -p /workspace /home/auditor/.cache/go-build /home/auditor/.local/share/pnpm/store \
    && chown -R auditor:auditor /workspace /home/auditor

ENV PATH="/usr/local/go/bin:/home/auditor/go/bin:/home/auditor/.local/share/pnpm:${PATH}"
ENV CGO_ENABLED=0

USER auditor
WORKDIR /workspace

CMD ["sleep", "infinity"]
