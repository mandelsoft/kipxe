# HTTP Server matching and serving iPXE Requests

This project provides an HTTP server answering to requests according
to matches, resources and mappings described as kubernetes resources.

It provides three different parts:
- a library for an HTTP server serving requests according to 
  configured query-matchers, mappings, resources and an optional
  *Discovery API* for metadata
- a Kubernetes controller offering such a server by feeding it with
  configuration taken from Kubernetes resources.
- a Kubernetes controller implementing the discovery API based on 
  a machine Kubernetes resource

This ecosystem is intended to be used to serve iPXE requests when 
booting machines based on predefined rules. But it can also be used
as a general matching engine to match requests to configurable resources.

## The Discovery API

The matching engive use a registration based api to map a request
to request metadata used for the further matching process.

By default the query parameters of the request are extracted and used
as metadata. If a handler for the *Discovery API* is registered,
it gets the default metadata and the request to provide final metadata. 


```golang
        type MetaDataMapper interface {
	       Map(logger logger.LogContext, values MetaData, req *http.Request) (MetaData, error)
        }
```

## The Matching Engine

The matching engine matches http requests according to their resource
path and query parameters and determines content to be sent.

The mapping of requests to served resources is controlled by three kinds of
elements:

- *Matcher* elements describe matches based on requset metadata. Every matcher
   refers to a *Profile* for describing available resources
- *Profile* elements provide a list of supported resource paths and maps
   those paths to *Resource* elements
- *Resource* elements describe content, either as inline content or based on
   URLs. Resources may be futhermore processed incorporating the metadata before
   seving their content.

This resolving of requests is a five step approach:
1) Map the request to some metadata by an optional [*Discovery API*](#the-discovery-api)
   implementation. Its task is to identify a described element and complete
   the provided metadata used for the resource matching process.
   By default (no implementation for the API attached) the query parameters
   are used as metadata for the following matching process
2) Determine all matchers matching the metadata derived from the request.
   This list is ordered by a weight provided by every matcher
3) Determine a resource used to serve the request by examining all
   matched profiles in the order of their weight (highest weight first)
4) Map the metadata according to the settings of the involved elements
   (matcher, profile, resource) 
5) Process the content of the found resource by using it as template
   if the resource is of type text, json or yaml 

For processing the metadata every element may provide own *values* or *mappings*. 
- Values are used to enrich the metadata.
- Mappings use a [*spiff template*](https://github.com/mandelsoft/spiff/blob/master/README.md) to
  map the metadata.

Finally the resource is used as template:
- if it is yaml or json, it is used again as *spiff template* by using the metadata as stub for the merge process.
- if it is a text document, the go templating engine is used for processing

As a special case for yaml or json documents, the processed metadata is directly used as response,
if no further content is defined for the resource

## The HTTP Server

The HTTP server uses the matching engine to resolve requests to content to be
served. It contains a cache for URL based content. Such requests are directly
served from the filesystem cache. (Binary content (`application/octet-stream`)
is never processed)

The cache supports a simple TTL for house keeping.

## The Kubernetes Backend

This project offers two kubernetes controllers:
- a controller to offer an HTTP server and matching engine which is
  configured by kubernetes resources.
- a controller implementing the [*Discovery API*](#the-discovery-api) of the
  matching engine based on a *Machine* resource.

### The HTTP server

The provided Kubernetes controller uses three dedicated kinds of Kubernetes
resources to describe *Matchers*, *Profiles* and *Resources*.

Every change here will automatically be used to reconfigure the matching engine
of the provided http server.

#### Matchers

Here are some examples for profile definitions.

<details><summary>A matcher may look like this when matching always</summary>

```yaml
apiVersion: ipxe.mandelsoft.org/v1alpha1
kind: BootProfileMatcher
metadata:
  name: info 
  namespace: default
spec:
  weight: 0
  profileName: info
```

</details>

<details><summary>Or matching a dedicated type of requester by enriched metadata</summary>

```yaml
apiVersion: ipxe.mandelsoft.org/v1alpha1
kind: BootProfileMatcher
metadata:
  name: matcher1
  namespace: default
spec:
  selector:
    matchLabels:
      machine-type: node
  profileName: node-profile

```

</details>


#### Profiles

Here are some examples for profile definitions.

<details><summary>A resource may look like this when using go templates</summary>

```yaml
apiVersion: ipxe.mandelsoft.org/v1alpha1
kind: BootProfile
metadata:
  name: info 
  namespace: default
spec:
  resources:
    - path: info
      documentName: info
```

</details>


#### Resources

here are some examples for resource definitions:

<details><summary>A resource may look like this when using go templates</summary>

```yaml
apiVersion: ipxe.mandelsoft.org/v1alpha1
kind: BootResource
metadata:
  name: script
  namespace: default
spec:
  mimeType: text/plain
  text: |+
    #!/bin/bash

    echo hallo echo
    echo {{ .metadata.uuid }}
```

</details>


<details><summary>A resource may look like this when using URL based content</summary>

```yaml
apiVersion: ipxe.mandelsoft.org/v1alpha1
kind: BootResource
metadata:
  name: kipxe-tar
  namespace: default
spec:
  mimeType: application/octet-stream
  URL: http://github.com/mandelsoft/kipxe/tarball/master
```

</details>


<details><summary>Or just a processed json document</summary>

```yaml
apiVersion: ipxe.mandelsoft.org/v1alpha1
kind: BootResource
metadata:
  name: info
  namespace: default
spec:
  mimeType: application/json
  mapping:
    metadata: (( &temporary ))
    info:
      state:      (( defined(metadata.macsbypurpose) ? "known" :"unknown" ))
      uuid:       (( metadata.uuid || "" ))
      attributes: (( metadata.attributes || {} ))
      macsbypurpose: (( metadata.macsbypurpose || ~~ ))
```

</details>

### The Machine Manager

The machine manager uses a Kubernetes *Machine* resource to configure
metadata for an identified requester. It is identified by the
provided query parameters that must either match the `UUID` or 
a provided `MAC` address.

To enrich the metadata of the request formally the UUID and all configured MAC
addresses will be used. Additionally an arbitrary set of values will be added
to controll the following matching process.

<details><summary>A machine resource additionally defining a machine type</summary>

```yaml
apiVersion: ipxe.mandelsoft.org/v1alpha1
kind: Machine
metadata:
  name: cluster1-node-1
  namespace: default
spec:
  uuid: abc-123
  macs:
    standard:
      - mac1
      - mac2
    infra:
      - mac3
  values:
    message: "this is a test"
    machine-type: node
```

</details>

## Command Line Reference

```
kipxe serves iPXE requests based on Kubernetes rosources

Usage:
  kipxe [flags]

Flags:
      --bind-address-http string                         HTTP server bind address
      --cache-cleanup.pool.resync-period duration        Period for resynchronization for pool cache-cleanup
      --cache-cleanup.pool.size int                      Worker pool size for pool cache-cleanup
      --cache-dir string                                 enable URL caching in a dedicated directory
      --cache-ttl duration                               TTL for cache entries
      --config string                                    config file
  -c, --controllers string                               comma separated list of controllers to start (<name>,<group>,all) (default "all")
      --cpuprofile string                                set file for cpu profiling
      --default.pool.size int                            Worker pool size for pool default
      --disable-namespace-restriction                    disable access restriction for namespace local access only
      --grace-period duration                            inactivity grace period for detecting end of cleanup for shutdown
  -h, --help                                             help for kipxe
      --ipxe.cache-cleanup.pool.resync-period duration   Period for resynchronization for pool cache-cleanup of controller ipxe (default 1m0s)
      --ipxe.cache-cleanup.pool.size int                 Worker pool size for pool cache-cleanup of controller ipxe (default 1)
      --ipxe.cache-dir string                            enable URL caching in a dedicated directory of controller ipxe
      --ipxe.cache-ttl duration                          TTL for cache entries of controller ipxe (default 10m0s)
      --ipxe.default.pool.size int                       Worker pool size for pool default of controller ipxe (default 5)
      --ipxe.local-namespace-only                        server only resources in local namespace of controller ipxe
      --ipxe.pool.resync-period duration                 Period for resynchronization of controller ipxe
      --ipxe.pool.size int                               Worker pool size of controller ipxe
      --ipxe.pxe-port int                                pxe server port of controller ipxe (default 8081)
      --kubeconfig string                                default cluster access
      --kubeconfig.disable-deploy-crds                   disable deployment of required crds for cluster default
      --kubeconfig.id string                             id for cluster default
      --lease-name string                                name for lease object
      --local-namespace-only                             server only resources in local namespace
  -D, --log-level string                                 logrus log level
      --machines.default.pool.size int                   Worker pool size for pool default of controller machines (default 5)
      --machines.local-namespace-only                    server only resources in local namespace of controller machines
      --machines.pool.size int                           Worker pool size of controller machines
      --maintainer string                                maintainer key for crds (defaulted by manager name)
      --name string                                      name used for controller manager
      --namespace string                                 namespace for lease (default "kube-system")
  -n, --namespace-local-access-only                      enable access restriction for namespace local access only (deprecated)
      --omit-lease                                       omit lease for development
      --plugin-file string                               directory containing go plugins
      --pool.resync-period duration                      Period for resynchronization
      --pool.size int                                    Worker pool size
      --pxe-port int                                     pxe server port
      --server-port-http int                             HTTP server port (serving /healthz, /metrics, ...)
      --version                                          version for kipxe
```

Supported controllers are:
 - `ipxe`: http server with matching engine configured by Kubernetes resources
 - `machines`: controller implementing the [*Discovery API*](#the-discovery-api) based
    on machine resources.