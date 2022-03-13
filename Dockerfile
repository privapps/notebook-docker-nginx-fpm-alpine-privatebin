FROM privatebin/nginx-fpm-alpine:1.3.5

ENV S6_READ_ONLY_ROOT 0

USER 0

RUN  \
   sed -i 's/index\ index.php\ index.html/index\ index.html\ index.php/g' /etc/nginx/http.d/site.conf \
   && sed -i 's/listen\ \[:/#listen\ \[:/g' /etc/nginx/http.d/site.conf \
   && wget https://github.com/privapps/notebook/releases/download/v1.1.0/notebook.tar.gz \
   && tar xf notebook.tar.gz && rm notebook.tar.gz \   
   && mv notebook/* /var/www/ && rmdir notebook-privapps/ \
   && cd notebook && sed -i "s|/index.html|/notebook/index.html|g" index.html


COPY config.json /var/www/assets/

USER 65534:82

ENTRYPOINT ["/init"]
