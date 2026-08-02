# Non-root revision of the project-local IDA Pro 9.4 tool image.
#
# The inherited image keeps its license under /root/.idapro, so invoking IDA
# with the repository owner's UID fails before analysis begins.  Copy only the
# runtime configuration into the image's existing ubuntu account; project data
# remains mounted separately and the runtime still sets an explicit UID/GID.
FROM ida-pro-9.4-ver2

RUN install -d -o 1000 -g 1000 /home/ubuntu/.idapro \
 && cp /root/.idapro/ida.reg /root/.idapro/ida-config.json \
       /root/.idapro/idapro.hexlic /home/ubuntu/.idapro/ \
 && chown 1000:1000 /home/ubuntu/.idapro/* \
 && chmod 600 /home/ubuntu/.idapro/idapro.hexlic \
 && chmod 644 /home/ubuntu/.idapro/ida.reg \
              /home/ubuntu/.idapro/ida-config.json
