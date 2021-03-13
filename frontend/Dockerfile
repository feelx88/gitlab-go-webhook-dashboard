FROM node:alpine as build

ARG backend_url
ARG icon_src
ENV VUE_APP_BACKEND_URL=$backend_url
ENV VUE_APP_ICON_SRC=$icon_src

WORKDIR /app
COPY . .

RUN npm ci && npm run build

FROM nginx:alpine
COPY --from=0 /app/dist /usr/share/nginx/html
COPY default.conf /etc/nginx/conf.d/
