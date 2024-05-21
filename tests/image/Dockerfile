FROM --platform=amd64 node:16-bookworm-slim AS builder

ARG SUPERSET_VERSION=4.0.0
ENV BUILD_CMD="build" \
    PUPPETEER_SKIP_CHROMIUM_DOWNLOAD=true
WORKDIR /app

RUN apt-get update -qq \
    && apt-get install -yqq --no-install-recommends \ 
        curl \
        tar \
        ca-certificates \
        build-essential \
        python3

RUN curl -L https://github.com/apache/superset/archive/refs/tags/${SUPERSET_VERSION}.tar.gz | tar xz

WORKDIR /app/superset-frontend

RUN mkdir docker superset-frontend
RUN mv /app/superset-${SUPERSET_VERSION}/docker/frontend-mem-nag.sh docker/
RUN ./docker/frontend-mem-nag.sh

RUN mv /app/superset-${SUPERSET_VERSION}/superset-frontend/package.json ./ \
    && mv /app/superset-${SUPERSET_VERSION}/superset-frontend/package-lock.json ./ \
    && npm ci

RUN cp -r /app/superset-${SUPERSET_VERSION}/superset-frontend/* ./

RUN npm run ${BUILD_CMD}

FROM python:3.9-slim

ARG SUPERSET_VERSION=4.0.0
WORKDIR /app
ENV LANG=C.UTF-8 \
    LC_ALL=C.UTF-8 \
    SUPERSET_ENV=production \
    FLASK_APP="superset.app:create_app()" \
    PYTHONPATH="/app/pythonpath" \
    SUPERSET_HOME="/app/superset_home" \
    SUPERSET_PORT=8088

RUN mkdir -p ${PYTHONPATH} superset/static superset-frontend apache_superset.egg-info requirements \
    && useradd --user-group -d ${SUPERSET_HOME} -m --no-log-init --shell /bin/bash superset \
    && apt-get update -qq && apt-get install -yqq --no-install-recommends \
        build-essential \
        default-libmysqlclient-dev \
        libsasl2-dev \
        libsasl2-modules-gssapi-mit \
        libpq-dev \
        libecpg-dev \
        libldap2-dev \
    && touch superset/static/version_info.json \
    && chown -R superset:superset ./* \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/superset-${SUPERSET_VERSION}/setup.py . 
COPY --from=builder /app/superset-${SUPERSET_VERSION}/MANIFEST.in ./
COPY --from=builder /app/superset-${SUPERSET_VERSION}/README.md ./
COPY --from=builder /app/superset-${SUPERSET_VERSION}/requirements/*.txt requirements/ 
COPY --from=builder /app/superset-frontend/package.json superset-frontend/ 

RUN pip install -r requirements/local.txt

COPY --from=builder /app/superset/static/assets superset/static/assets
COPY --from=builder /app/superset-${SUPERSET_VERSION}/superset superset/

RUN --mount=type=cache,target=/root/.cache/pip pip install -e . \
    && flask fab babel-compile --target superset/translations \
    && chown -R superset:superset superset/translations

COPY --from=builder /app/superset-${SUPERSET_VERSION}/docker/run-server.sh /usr/bin/run-server.sh
RUN chmod a+x /usr/bin/run-server.sh

USER superset 
EXPOSE ${SUPERSET_PORT}
CMD ["/usr/bin/run-server.sh"]