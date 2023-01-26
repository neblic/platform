FROM alpine:3.17

RUN apk --update add ca-certificates sudo
RUN addgroup -S otelcol && adduser -D -S otelcol -G otelcol

COPY --chown=755 otelcol /bin/otelcol
COPY dist/otelcol/config.yaml /etc/neblic/otelcol/config.yaml

COPY --chmod=755 dist/otelcol/entrypoint.sh /bin/entrypoint.sh

CMD ["/bin/entrypoint.sh"]