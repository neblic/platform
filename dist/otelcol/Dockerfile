FROM alpine:3.17
RUN apk --update add ca-certificates

ARG USER_UID=10001
USER ${USER_UID}

COPY --chown=755 otelcol /bin/otelcol
COPY dist/otelcol/config.yaml /etc/neblic/otelcol/config.yaml

COPY --chmod=755 dist/otelcol/entrypoint.sh /bin/entrypoint.sh
CMD ["/bin/entrypoint.sh"]
EXPOSE 4317 8899