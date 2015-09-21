# Install etcd
```
curl -L  https://github.com/coreos/etcd/releases/download/v2.2.0/etcd-v2.2.0-linux-amd64.tar.gz -o etcd-v2.2.0-linux-amd64.tar.gz
tar xzvf etcd-v2.2.0-linux-amd64.tar.gz
cd etcd-v2.2.0-linux-amd64
./etcd
```

# Start etcd cluster
- DEV: 172.16.1.104/15.209.122.238, Ubuntu 14.04, RAM 16GB, 8 VCPU, Disk 160GB
- PRD: 172.16.1.105/15.209.122.241, Ubuntu 14.04, RAM 16GB, 8 VCPU, Disk 160GB
```
etcd -name infra0 -initial-advertise-peer-urls http://172.16.1.104:2380 \
 -listen-peer-urls http://172.16.1.104:2380 \
 -listen-client-urls http://0.0.0.0:2379 \
 -advertise-client-urls http://0.0.0.0:2379 \
 -initial-cluster-token etcd-cluster-1 \
 -initial-cluster infra0=http://172.16.1.104:2380,infra1=http://172.16.1.105:2380 \
 -initial-cluster-state new

 etcd -name infra1 -initial-advertise-peer-urls http://172.16.1.105:2380 \
  -listen-peer-urls http://172.16.1.105:2380 \
  -listen-client-urls http://0.0.0.0:2379 \
  -advertise-client-urls http://0.0.0.0:2379 \
  -initial-cluster-token etcd-cluster-1 \
  -initial-cluster infra0=http://172.16.1.104:2380,infra1=http://172.16.1.105:2380 \
  -initial-cluster-state new
```

# Health
```
curl -L http://172.16.1.104:2379/health
curl -L http://172.16.1.105:2379/health
etcdctl cluster-health
```

# Local Testing

```
go install
service-scheduler -discoveryEtcdURL="http://172.17.42.1:2379" -discoveryServiceDir="/services"
```

```
curl http://127.0.0.1:2379/v2/keys/services/repl-1 -XPUT -d value="http://localhost:9001/api/1.0/replicate?filter=list&repos=busybox" -d ttl=60
curl http://172.16.1.105:2379/v2/keys/services/repl-1

curl http://localhost:2379/v2/keys/services/repl-1

curl http://127.0.0.1:2379/v2/keys/services/repl-2 -XPUT -d value="http://localhost:9000/api/1.0/replicate?filter=list&repos=centos" -d ttl=10

etcdctl update /services/repl '{"name":"repl", "url":"http://localhost:8000/api/1.0/replicate?filter=list&repos=centos,busybox"}'
etcdctl get /services/repl

```

# Test with container
cd /home/ubuntu/go-work/src/ghe/chun-yang-wang/service-scheduler
sudo docker build -t service-scheduler .

sudo docker run -p 8001:8000 --name sche --rm -e DISCOVERY_ETCD_URL="http://172.17.42.1:2379" -e DISCOVERY_SERVICE_DIR="/services" service-scheduler

# Buid Images
cdr
sudo docker build -t official-images-replication .
sudo docker tag official-images-replication docker.hos.hpecorp.net/chunyang/official-images-replication:hack-0.4
sudo docker tag -f official-images-replication docker.hos.hpecorp.net/chunyang/official-images-replication:hack-latest
sudo docker push docker.hos.hpecorp.net/chunyang/official-images-replication:hack-0.4
sudo docker push docker.hos.hpecorp.net/chunyang/official-images-replication:hack-latest

cds
sudo docker build -t service-scheduler .
sudo docker tag service-scheduler docker.hos.hpecorp.net/chunyang/service-scheduler:hack-0.2
sudo docker tag -f service-scheduler docker.hos.hpecorp.net/chunyang/service-scheduler:hack-latest
sudo docker push docker.hos.hpecorp.net/chunyang/service-scheduler:hack-0.2
sudo docker push docker.hos.hpecorp.net/chunyang/service-scheduler:hack-latest


# Demo Playbook
## Start etcd Cluster
```
etcd -name infra0 -initial-advertise-peer-urls http://172.16.1.104:2380 \
 -listen-peer-urls http://172.16.1.104:2380 \
 -listen-client-urls http://0.0.0.0:2379 \
 -advertise-client-urls http://0.0.0.0:2379 \
 -initial-cluster-token etcd-cluster-1 \
 -initial-cluster infra0=http://172.16.1.104:2380,infra1=http://172.16.1.105:2380 \
 -initial-cluster-state new &

 etcd -name infra1 -initial-advertise-peer-urls http://172.16.1.105:2380 \
  -listen-peer-urls http://172.16.1.105:2380 \
  -listen-client-urls http://0.0.0.0:2379 \
  -advertise-client-urls http://0.0.0.0:2379 \
  -initial-cluster-token etcd-cluster-1 \
  -initial-cluster infra0=http://172.16.1.104:2380,infra1=http://172.16.1.105:2380 \
  -initial-cluster-state new &
```

## Start containers: repl-1, repl-2, sche
### Option-1: docker-compose up

### Manually:
#### Start official-images-replication containers: repl-1, repl-2
```
sudo docker pull docker.hos.hpecorp.net/chunyang/official-images-replication:hack-latest

sudo docker run -v /var/run/docker.sock:/var/run/docker.sock -p 9001:8000 --name repl-1 --rm -e USERNAME="docker.registry@hp.com" -e PASSWORD="CHANGEME" -e SERVER_MODE=true -e DISCOVERY_ETCD_URLS="http://172.16.1.104:2379,http://172.16.1.105:2379" -e DISCOVERY_SERVICE_KEY="/services/repl-1" -e DISCOVERY_SERVICE_URL="http://172.16.1.104:9001/api/1.0/replicate?filter=list&repos=busybox" docker.hos.hpecorp.net/chunyang/official-images-replication:hack-latest

sudo docker run -v /var/run/docker.sock:/var/run/docker.sock -p 9002:8000 --name repl-2 --rm -e USERNAME="docker.registry@hp.com" -e PASSWORD="CHANGEME" -e SERVER_MODE=true -e DISCOVERY_ETCD_URLS="http://172.16.1.104:2379,http://172.16.1.105:2379" -e DISCOVERY_SERVICE_KEY="/services/repl-2" -e DISCOVERY_SERVICE_URL="http://172.16.1.104:9002/api/1.0/replicate?filter=list&repos=centos" docker.hos.hpecorp.net/chunyang/official-images-replication:hack-latest

curl http://172.17.42.1:2379/v2/keys/services/repl-1
curl http://172.17.42.1:2379/v2/keys/services/repl-2
curl http://172.16.1.104:2379/v2/keys/services | grep ttl
curl http://172.16.1.105:2379/v2/keys/services | grep ttl

curl http://172.17.42.1:2379/v2/keys/services | grep ttl
```
#### Start scheduler container : sche
```
sudo docker pull docker.hos.hpecorp.net/chunyang/service-scheduler:hack-latest
sudo docker run -p 8001:8000 --name sche --rm -e DISCOVERY_ETCD_URLS="http://172.16.1.104:2379,http://172.16.1.105:2379" -e DISCOVERY_SERVICE_DIR="/services" docker.hos.hpecorp.net/chunyang/service-scheduler:hack-latest
```