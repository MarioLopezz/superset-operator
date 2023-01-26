FROM rockylinux:8 AS builder

ARG SUPERSET_VERSION=2.0.1

ENV LANG=C.UTF-8 \
    LC_ALL=C.UTF-8 \
    FLASK_ENV=production \
    FLASK_APP="superset.app:create_app()" \
    PYTHONPATH="/app/pythonpath" \
    SUPERSET_HOME="/app/superset_home" \
    SUPERSET_PORT=8088

RUN dnf install -y sudo
RUN sudo dnf install -y gcc gcc-c++ libffi-devel python38-devel python38-pip python38-wheel openssl-devel cyrus-sasl-devel openldap-devel
RUN pip3 install --upgrade pip
RUN pip install virtualenv

RUN python3 -m venv venv && \
  . venv/bin/activate

RUN pip install impyla && \
  pip install pyhive && \
  pip install sqlalchemy-solr

COPY requirements/*.txt app/requirements/
RUN cd /app && \
  pip install --no-cache -r requirements/base.txt && \
  pip install --no-cache -r requirements/development.txt

RUN pip install apache-superset==$SUPERSET_VERSION
RUN superset db upgrade

COPY ./run-server.sh /usr/bin/

RUN chmod a+x /usr/bin/run-server.sh

WORKDIR /app

USER superset

EXPOSE ${SUPERSET_PORT}

CMD /usr/bin/run-server.sh