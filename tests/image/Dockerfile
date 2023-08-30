ARG PY_VER=3.9-slim
ARG SUPERSET_VERSION=2.0.1

FROM node:16-slim AS superset-node

ENV PUPPETEER_SKIP_CHROMIUM_DOWNLOAD=true

RUN apt-get update && apt-get install -y git \
    && git clone https://github.com/apache/superset.git /app \
    && cd /app \
    && git checkout ${SUPERSET_VERSION} \
    && apt-get remove -y git && apt-get autoremove -y

WORKDIR /app/superset-frontend
RUN npm ci
RUN npm run build

FROM python:${PY_VER} AS production

ENV LANG=C.UTF-8 \
    LC_ALL=C.UTF-8 \
    FLASK_DEBUG=1 \
    FLASK_APP="superset.app:create_app()" \
    PYTHONPATH="/app/pythonpath" \
    SUPERSET_HOME="/app/superset_home" \
    SUPERSET_PORT=8088 

RUN mkdir -p ${PYTHONPATH} \
    && useradd --user-group -d ${SUPERSET_HOME} -m --no-log-init --shell /bin/bash superset \
    && apt-get update -y \
    && apt-get install -y --no-install-recommends \
        build-essential \
        curl \
        default-libmysqlclient-dev \
        libsasl2-dev \
        libsasl2-modules-gssapi-mit \
        libpq-dev \
        libecpg-dev \
        libldap2-dev \
        libsasl2-dev \
    && rm -rf /var/lib/apt/lists/*

COPY --from=superset-node /app/requirements /app/requirements
COPY --from=superset-node /app/setup.py /app/
COPY --from=superset-node /app/MANIFEST.in /app/
COPY --from=superset-node /app/README.md /app/

COPY --from=superset-node /app/superset-frontend/package.json /app/superset-frontend/
RUN mkdir -p /app/superset/static && touch /app/superset/static/version_info.json
COPY --from=superset-node /app/superset/static/assets /app/superset/static/assets

RUN pip install --upgrade pip

COPY --from=superset-node /app/superset /app/superset
RUN cd /app \
    && chown -R superset:superset * \
    && pip install --no-cache -r requirements/local.txt \
    && pip install -e . \
    && flask fab babel-compile --target superset/translations
    
COPY --from=superset-node /app/docker/run-server.sh /usr/bin/
RUN chmod a+x /usr/bin/run-server.sh
WORKDIR /app
USER superset
EXPOSE ${SUPERSET_PORT}
CMD /usr/bin/run-server.sh
