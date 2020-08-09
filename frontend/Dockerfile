FROM node:alpine as build

WORKDIR /app
COPY . .

RUN npm ci && npm run build

FROM nginx:alpine
COPY --from=0 /app/dist /var/www/html
