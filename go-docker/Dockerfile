FROM --platform=$BUILDPLATFORM alpine:3.15

ARG TARGETOS TARGETARCH

COPY /*.gz /
COPY /static /static
COPY /run.sh .

RUN  \
   if [ "$TARGETARCH" = "amd64" ] ; then TGT='amd64'; else TGT='arm64'; fi \
   && gzip -d privapps-notebook-server-linux-$TGT.gz \
   && mv privapps-notebook-server-linux-$TGT notebook-server \
   && chmod u+x notebook-server \
   && mkdir -p notes/data static && rm *.gz


EXPOSE 8080

ENTRYPOINT ["sh" ,"run.sh"]
