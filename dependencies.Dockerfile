# This is a renovate-friendly source of Docker images.
FROM davidanson/markdownlint-cli2:v0.24.2@sha256:c768e94d8bb161c449bde28f60656929ca1f02c3010a405f65c9cac71563a889 AS markdown
FROM python:3.13.6-slim-bullseye@sha256:e98b521460ee75bca92175c16247bdf7275637a8faaeb2bcfa19d879ae5c4b9a AS python
