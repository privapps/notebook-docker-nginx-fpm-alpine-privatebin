FROM --platform=$BUILDPLATFORM alpine:3.15

ARG TARGETOS TARGETARCH

RUN  \
   if [ "$TARGETARCH" = "amd64" ] ; then TGT='amd64'; else TGT='arm64'; fi \
   && wget https://github.com/privapps/notebook-docker-nginx-fpm-alpine-privatebin/raw/latest-binaries/privapps-notebook-server-linux-$TGT.gz \
   && gzip -d privapps-notebook-server-linux-$TGT.gz \
   && mv privapps-notebook-server-linux-$TGT notebook-server \
   && chmod u+x notebook-server \
   && mkdir -p notes/data static

COPY /static /static
COPY /run.sh .

EXPOSE 8080

ENTRYPOINT ["sh" ,"run.sh"]
