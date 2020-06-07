FROM nginx:1.17.10-alpine

CMD ["nginx", "-g", "daemon off;"]
RUN ln -sf /dev/stdout /var/log/nginx/access.log && ln -sf /dev/stderr /var/log/nginx/error.log

COPY ["nginx.conf", "/etc/nginx/nginx.conf"]
COPY ["index.html", "/usr/share/nginx/html/"]
