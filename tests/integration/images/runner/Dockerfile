FROM golang:1.11.1-alpine

# install dependencies
RUN apk update \
    && apk add --no-cache --virtual .build-deps \
    && apk add bash make curl openssh git \
    && ln -sf /usr/share/zoneinfo/Etc/UTC /etc/localtime \
    && apk -Uuv add groff less vim \
    && apk add python3 \
    && python3 -m ensurepip \
    && rm -r /usr/lib/python*/ensurepip \
    && pip3 install --upgrade pip setuptools \
    && if [ ! -e /usr/bin/pip ]; then ln -s pip3 /usr/bin/pip ; fi \
    && if [[ ! -e /usr/bin/python ]]; then ln -sf /usr/bin/python3 /usr/bin/python; fi \
    && pip3 install awscli \
    && apk --purge -v del py-pip \
    && rm -rf /var/cache/apk/* \
    && go get github.com/lmammino/invok \
    && go install github.com/lmammino/invok

# copy test runner scripts
COPY ./src /opt/scripts/runner
COPY ./start.sh /opt/scripts/start.sh

# install python dependencies
RUN pip install -r /opt/scripts/runner/requirements.txt \
    && chmod +x /opt/scripts/runner/runner.py \
    && chmod +x /opt/scripts/start.sh

# copy test config
COPY ./test-config.json /etc/test-config.json

# copy fixtures
COPY ./fixtures/ /var/storage/

CMD [ "bash" ]
