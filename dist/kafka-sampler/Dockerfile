FROM scratch

ARG USER_UID=10001
USER ${USER_UID}

COPY --chmod=755 kafka-sampler /kafka-sampler
COPY dist/kafka-sampler/config.yaml /etc/neblic/kafka-sampler/config.yaml
ENTRYPOINT ["/kafka-sampler"]