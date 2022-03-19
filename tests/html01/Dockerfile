FROM yolean/envoystatic:tooling

COPY . /workspace

RUN [ "/usr/local/bin/envoystatic", \
  "route", \
  "--in=/workspace", \
  "--out=/tmp/docroot", \
  "--rdsyaml=/tmp/route.yaml" ]

FROM yolean/envoystatic:envoy

COPY --from=0 /tmp/route.yaml /etc/envoy/rds/
COPY --from=0 /tmp/docroot /var/docroot
