# This is a renovate-friendly source of Docker images.
FROM davidanson/markdownlint-cli2:v0.24.1@sha256:caf0698eaea0268bcf35b5e2d7ba946b5899d323e765bc2862054405658cd948 AS markdown
FROM python:3.13.6-slim-bullseye@sha256:e98b521460ee75bca92175c16247bdf7275637a8faaeb2bcfa19d879ae5c4b9a AS python
