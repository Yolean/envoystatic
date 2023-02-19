ARG envoystatic_tooling_image
ARG envoystatic_image

FROM ${envoystatic_tooling_image}

COPY . /workspace

RUN [ "/usr/local/bin/envoystatic", \
  "route", \
  "--in=/workspace", \
  "--out=/tmp/docroot", \
  "--rdsyaml=/tmp/route.yaml" ]

FROM ${envoystatic_image}

COPY --from=0 /tmp/route.yaml /etc/envoy/rds/
COPY --from=0 /tmp/docroot /var/docroot
