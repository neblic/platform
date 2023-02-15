FROM alpine:3.17

RUN apk --update add ca-certificates sudo
RUN addgroup -S otelcol && adduser -D -S otelcol -G otelcol

COPY --chown=755 neblictl /bin/neblictl
COPY --chown=755 otelcol /bin/otelcol
COPY dist/otelcol/otelcol.yaml /etc/neblic/otelcol/otelcol.yaml

COPY --chmod=755 dist/otelcol/entrypoint.sh /bin/entrypoint.sh

CMD ["/bin/entrypoint.sh"]