from privatebin/nginx-fpm-alpine:1.3.4

ENV S6_READ_ONLY_ROOT 0

USER 0

RUN  \
   sed -i 's/index\ index.php\ index.html/index\ index.html\ index.php/g' /etc/nginx/http.d/site.conf \
   && sed -i 's/listen\ \[:/#listen\ \[:/g' /etc/nginx/http.d/site.conf \
   && wget https://github.com/privapps/notebook/releases/download/v1.0.0/notebook-privapps.tar.xz \
   && tar xf notebook-privapps.tar.xz \
   && mv notebook-privapps/* /var/www/


COPY config.json /var/www/assets/

RUN \
  find /var/www -type f | xargs chmod -x \
  && rm -f notebook-privapps.tar.xz \
  && rmdir notebook-privapps/

ENTRYPOINT ["/init"]
