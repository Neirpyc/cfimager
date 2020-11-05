FROM nginx:1.19.2-alpine

RUN touch /var/run/nginx.pid && \
  chown -R 500:500 /var/run/nginx.pid && \
  chown -R 500:500 /var/cache/nginx

USER 500

WORKDIR /cfimager_server

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY ./web ./web
COPY config.toml .
COPY src/server server

EXPOSE 80/tcp

CMD /bin/sh -c "/usr/sbin/nginx && ./server"