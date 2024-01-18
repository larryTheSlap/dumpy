![Dumpy Logo](dumpy.png )

# Dumpy - Kubernetes Network Traffic Capture Plugin

Dumpy is an advanced kubectl plugin designed for Kubernetes administrators, providing seamless network traffic capture from various resources. It excels in isolating captures to specific pod containers, ensuring security and accurate analysis. Dumpy dynamically creates dedicated sniffers that will run tcpdump for each target pod.


- [Features](#features)
- [Quick Start](#quick-start)
  - [Installation](#installation)
  - [Capture Network Traffic](#capture-network-traffic)
  - [Capture Details](#get-capture-details)
  - [Export Captures](#export-captures)
  - [More Dumpy Operations](#more-dumpy-operations)
- [Contribution](#contribution)
- [License](#license)

## Features

- **Dynamic Sniffer Creation:** Optimizes resource utilization by creating a dedicated sniffer for each pod, guaranteeing accurate and unobtrusive analysis.

- **Flexible Filtering:** Apply custom TCPDump filters for fine-grained control over captured data.

- **Persistent Volume Claim (PVC) Support:** Dumpy supports PVCs to store network captures, crucial for production environments. This feature ensures worker nodes storage disks are not impacted during captures.

- **Process Namespace Isolation:** Ensures security by leveraging PID namespaces, running `tcpdump` exclusively within the targeted container's PID namespace.

- **Deployment Flexibility:** Deploy Dumpy sniffers in a separate namespace, offering adaptability to various security policies and restrictions in production environments. This ensures compatibility without compromising the effectiveness of network traffic capture.

- **Export Capabilities:** Export captured data in PCAP format for further analysis.

- **Concurrent Captures**: Ability to run mutliple dumpy captures concurrently.  

## Quick Start

### Installation

Dumpy exclusively supports Kubernetes clusters using the containerd runtime. To install Dumpy, download the right [release](https://https://github.com/larryTheSlap/dumpy/releases/tag/v0.1.0) for your OS  , unzip it then move the `kubectl-dumpy` binary where kubectl is located.
- Linux install :
```bash
curl -L -O https://github.com/larryTheSlap/dumpy/releases/download/v0.1.0/dumpy_Linux_x86_64.tar.gz
tar xf dumpy_Linux_x86_64.tar.gz 
chmod +x kubectl-dumpy && sudo mv kubectl-dumpy /usr/bin/kubectl-dumpy
```
### Capture Network Traffic

Deploy sniffers to capture traffic from target pods :
```bash
kubectl dumpy capture <pod|deployment|replicaset|daemonset|statefulset> <resourceName> \
  -n <captureNamespace>      \ # Namespace where dumpy sniffers will be deployed           (default: current namespace) 
  -t <targetNamespace>       \ # Target resource namespace                                 (default: captureNamespace)
  -f <tcpdumpFilters>        \ # Tcpdump filters for capture                               (default: "-i any")
  -c <containerName>         \ # Specific target container for multi-container pods        (default: Main container)
  -v <pvcName>               \ # PVC name that sniffers mount to store tcpdump captures, RWX PVC for mutli-pod
  -i <dumpyImage>            \ # Dumpy docker image for private clusters                   (default: larrytheslap/dumpy:latest) 
  -s <imagePullSecret>       \ # Image pull secret name for private clusters to pull dumpy image
  --name <captureName>       \ # Set specific capture name, if not set dumpy generates it  (default: dumpy-<ID>) 

```
Example: 
```bash
# Deployment nginx-deploy in foo-ns with 3 replicas
$ kubectl get pod -n foo-ns
NAME                            READY   STATUS    RESTARTS   AGE
nginx-deploy-846d6f46b7-6v8jz   1/1     Running   0          11s
nginx-deploy-846d6f46b7-hz5s2   1/1     Running   0          11s
nginx-deploy-846d6f46b7-lss7q   1/1     Running   0          11s

# capture http traffic from nginx-deploy
$ kubectl dumpy capture deploy nginx-deploy -n foo-ns -f "-i any port 80"
Getting target resource info..
Dumpy init

Capture name: dumpy-49366665

  PodName: nginx-deploy-846d6f46b7-lss7q
  ContainerName: nginx
  ContainerID: 9f98acdf651372b1a16c7cfbb346716915d355943e70347d730ba46017c85384
  NodeName: kind-worker3


  PodName: nginx-deploy-846d6f46b7-hz5s2
  ContainerName: nginx
  ContainerID: 02c473d2ffae82a9cd7754733eef7a5deb71a09903118eb2820d81a1311d0869
  NodeName: kind-worker


  PodName: nginx-deploy-846d6f46b7-6v8jz
  ContainerName: nginx
  ContainerID: bd023032b481cd257e633aeda9c440afb13b205adf3445a5efad2bfc4fdd7e43
  NodeName: kind-worker2

sniffer-dumpy-49366665-1122 started sniffing
sniffer-dumpy-49366665-2382 started sniffing
sniffer-dumpy-49366665-7759 started sniffing
All dumpy sniffers are Ready.
```
Flag usage :
```bash
# capture all traffic from foo pod in current namespace
  kubectl dumpy capture pod foo
# capture all traffic from foo pod in foo-ns with specific capture name
  kubectl dumpy capture pod foo -t foo-ns --name <captureName>
# capture traffic from foo pod using tcpdump filters
  kubectl dumpy capture pod foo -f "-i any host 10.0.0.1 and port 80"
# capture traffic from foo pod specific container foo-cont
  kubectl dumpy capture pod foo -c foo-cont
# capture traffic from deployment foo-deploy in foo-ns namespace with sniffers in bar-ns
  kubectl dumpy capture deploy foo-deploy -t foo-ns -n bar-ns
# set dumpy image from private repository using docker pullSecret
  kubectl dumpy capture deploy foo-deploy -i <repository>/<path>/dumpy:latest -s <secretName>
# set pvc volume [RWX for multiple sniffers] to store tcpdump captures
  kubectl dumpy capture daemonset foo-ds -v <pvcName>
```

### Capture Details
When deploying multiple captures or to get details about them, use dumpy command `get` with no arguments to show minified details in table format about captures running in the specified namespace:
```bash
$ kubectl dumpy get -n foo-ns
NAME            NAMESPACE  TARGET                   TARGETNAMESPACE  TCPDUMPFILTERS                      SNIFFERS
----            ---------  ------                   ---------------  --------------                      --------
dumpy-51994723  foo-ns     pod/p-test-2             foo-ns           -i any host 10.0.0.13 and port 443  1/1
dumpy-80508655  foo-ns     deployment/bar-deploy    foo-ns           -i any port 443                     3/3
mycap           foo-ns     pod/p-test-1             foo-ns           -i any                              1/1
dumpy-49366665  foo-ns     deployment/nginx-deploy  foo-ns           -i any port 80                      3/3 
```
It is also possible to get more details about a specific capture by adding the capture name :
```bash
$ kubectl dumpy get -n foo-ns dumpy-80508655
Getting capture details..

name: dumpy-80508655
namespace: foo-ns
tcpdumpfilters: -i any port 443
image: larrytheslap/dumpy:latest
targetSpec:
    name: bar-deploy
    namespace: foo-ns
    type: deployment
    container: nginx
    items:
        bar-deploy-58974b698b-fv9kb  <-----  sniffer-dumpy-80508655-2245 [Running]
        bar-deploy-58974b698b-vxxx6  <-----  sniffer-dumpy-80508655-3634 [Running]
        bar-deploy-58974b698b-g44np  <-----  sniffer-dumpy-80508655-7969 [Running]

pvc:
pullsecret:
```
### Export Captures
Extract tcpdump .pcap files directly from capture sniffers using `export` command :
```bash
$ kubectl dumpy export <captureName> <targetDir> [-n captureNamespace]
```
Example: 
```bash
$ kubectl dumpy export dumpy-49366665 /tmp/dumps -n foo-ns
Downloading capture dumps from sniffers:
  nginx-deploy-846d6f46b7-lss7q ---> path /tmp/dumps/dumpy-49366665-nginx-deploy-846d6f46b7-lss7q.pcap
  nginx-deploy-846d6f46b7-6v8jz ---> path /tmp/dumps/dumpy-49366665-nginx-deploy-846d6f46b7-6v8jz.pcap
  nginx-deploy-846d6f46b7-hz5s2 ---> path /tmp/dumps/dumpy-49366665-nginx-deploy-846d6f46b7-hz5s2.pcap
```
### More Dumpy Operations
- `delete` command to remove capture and related sniffers
```bash
kubectl dumpy delete <captureName> [-n captureNamespace]
```
- `restart` command to redeploy specified capture sniffers with ability to use new tcpdump filters
```bash
kubectl dumpy restart <captureName> [-n captureNamespace] [-f tcpdump filters]
```
- `stop` command to terminate tcpdump process on sniffers
```bash
kubectl dumpy stop <captureName> [-n captureNamespace]
```
**Notes:** 
- Dumpy captures only exists as long as the sniffers do.
- Docker image is publicly available at [larrytheslap/dumpy]((https://hub.docker.com/r/larrytheslap/dumpy))
- Sniffer pods will also log traffic to stdout, helpful to validate the capture setup :
```bash
$ kubectl logs -n foo-ns sniffer-mycap-7181
#  ______  _   _ ___  _________ __   __
#  |  _  \| | | ||  \/  || ___ \\ \ / /
#  | | | || | | || .  . || |_/ / \ V /
#  | | | || | | || |\/| ||  __/   \ /
#  | |/ / | |_| || |  | || |      | |
#  |___/   \___/ \_|  |_/\_|      \_/
#
#
[INFO] #### using bin /usr/local/bin/crictl

[INFO] #### target container PID : 1688
[INFO] #### starting capture on target pod..
[INFO] #### Dumpy sniffer PID: 1999
tcpdump: data link type LINUX_SLL2
tcpdump: listening on any, link-type LINUX_SLL2 (Linux cooked v2), snapshot length 262144 bytes
reading from file -, link-type LINUX_SLL2 (Linux cooked v2), snapshot length 262144
23:56:25.637654 eth0  Out IP6 fe80::4448:2fff:fe86:18a0 > ff02::2: ICMP6, router solicitation, length 16
00:09:47.923265 eth0  B   ARP, Request who-has 10.244.2.5 tell 10.244.2.1, length 28
00:09:47.923271 eth0  Out ARP, Reply 10.244.2.5 is-at 46:48:2f:86:18:a0 (oui Unknown), length 28
00:09:47.923273 eth0  In  IP 10.244.1.5.55676 > 10.244.2.5.80: Flags [S], seq 1393873422, win 64240, options [mss 1460,sackOK,TS val 3667806333 ecr 0,nop,wscale 7], length 0
00:09:47.923282 eth0  Out IP 10.244.2.5.80 > 10.244.1.5.55676: Flags [S.], seq 1297939382, ack 1393873423, win 65160, options [mss 1460,sackOK,TS val 118234409 ecr 3667806333,nop,wscale 7], length 0
00:09:47.923316 eth0  In  IP 10.244.1.5.55676 > 10.244.2.5.80: Flags [.], ack 1, win 502, options [nop,nop,TS val 3667806333 ecr 118234409], length 0
00:09:47.923566 eth0  In  IP 10.244.1.5.55676 > 10.244.2.5.80: Flags [P.], seq 1:75, ack 1, win 502, options [nop,nop,TS val 3667806333 ecr 118234409], length 74: HTTP: GET / HTTP/1.1
00:09:47.923585 eth0  Out IP 10.244.2.5.80 > 10.244.1.5.55676: Flags [.], ack 75, win 509, options [nop,nop,TS val 118234409 ecr 3667806333], length 0
00:09:47.923690 eth0  Out IP 10.244.2.5.80 > 10.244.1.5.55676: Flags [P.], seq 1:239, ack 75, win 509, options [nop,nop,TS val 118234410 ecr 3667806333], length 238: HTTP: HTTP/1.1 200 OK
00:09:47.923724 eth0  In  IP 10.244.1.5.55676 > 10.244.2.5.80: Flags [.], ack 239, win 501, options [nop,nop,TS val 3667806334 ecr 118234410], length 0
```

## Contribution

Dumpy is open-source, and we welcome contributions. [Open issues](https://github.com/larryTheSlap/dumpy/issues) or submit pull requests to enhance functionality or fix bugs.

## License

Dumpy is licensed under the [Apache License 2.0](LICENSE). Feel free to use, modify, and distribute the code according to the terms of the Apache 2.0 License.

Happy Sniffing!
