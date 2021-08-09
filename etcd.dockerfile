FROM quay.io/coreos/etcd AS etcd

EXPOSE 2379
# EXPOSE 2380

ENTRYPOINT [ "/usr/local/bin/etcd", "--listen-client-urls=http://0.0.0.0:2379", "--advertise-client-urls=http://0.0.0.0:2379" ]