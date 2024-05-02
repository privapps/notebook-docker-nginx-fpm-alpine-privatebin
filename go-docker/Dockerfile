FROM --platform=$BUILDPLATFORM alpine:3.13

ARG TARGETOS TARGETARCH

COPY /static /static
COPY /privapps-notebook-server-linux-${TARGETARCH} /

RUN \
   mv privapps-notebook-server-linux-${TARGETARCH} notebook-server \
   && chmod u+x notebook-server \
   && mkdir -p notes/data

COPY /config.json /static/assets/

EXPOSE 8080

ENTRYPOINT ["/notebook-server"]
