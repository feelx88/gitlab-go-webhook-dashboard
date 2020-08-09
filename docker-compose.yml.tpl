version: '3.5'
services:
  backend:
    build:
      context: backend
    ports:
      - 8081:8081
  frontend:
    build:
      context: frontend
    ports:
      - 80:80
