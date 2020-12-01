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

The matching engine uses a registration based API to map a request
to request metadata used for the further matching process.

By default the query parameters of the request are extracted and used
as metadata. If a handler for the *Discovery API* is registered,
it gets the default metadata and the request to provide final metadata. 


```golang
        type MetaDataMapper interface {
	       Map(logger logger.LogContext, values MetaData, req *http.Request) (MetaData, error)
        }
```

There are two predefined implementations for a MetaDataMapper:
- `NewDefaultMetaDataMapper` uses a spiff template to map the metadata.
  The actual metadata is available as stub with the single field
  `metadata` containing the metadata fields.
  The mapping result is taken from a field `output`, if present, otherwise the
  root values is used.
- `NewURLMetaDataMapper` send the metadata as json document with a POST request
  and expects the mapped data again as json content. The mime type 
  `application/json` is used.
  
## The Matching Engine

The matching engine matches http requests according to their resource
path and query parameters and determines content to be sent.

The mapping of requests to served resources is controlled by three kinds of
elements:

- *Matcher* elements describe matches based on request metadata. Every matcher
   refers to a *Profile* for describing available resources
- *Profile* elements provide a list of supported resource paths and maps
   those paths to *Resource* elements
- *Resource* elements describe content, either as inline content or based on
   URLs. Resources may be futhermore processed incorporating the metadata before
   sevring their content.

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
  
This way the information bubbling down to the resource can be modified along the 
matching path. The matcher may provide info evaluated and/or mapped by the profile
settings and finally reaches the resource that might again adapt the values.
A common profile might be used by different matchers providing different inbound
values to control the processing of the finally matched resource(s) according
to the rules established by the profile.

Finally the resource is used as template:
- if it is yaml or json, it is used again as *spiff template* by using the
  metadata as stub for the merge process.
- if it is a text document, the go templating engine is used for processing

As a special case for yaml or json documents, the processed metadata is directly
used as response, if no further content is defined for the resource.
This way a resource object might directly provide a yaml template for
the final content.

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

Here are some examples for boot profile matcher definitions.

##### Generic Matcher

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

If neither a selector nor a matcher field is configured a profile matcher matches
always. This can be used to add common resource or, with a very low weight, to
add default resource (content) for resource requests.

##### Selector based Matcher

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

The easiest was to do this, is to use the selector field. It supports the
basic logic of a kubernetes label selector to match metadata fields.

The reuest metadata might be enriched by mappers and may contain more complex
deep structures. The metadata is basically a deep JSON document.

This data structure has been adapted to allow selectors to match nested fields,
also. The name of a nested fiels might be a string composed by a sequence
of field names seperated by a slash (`/`), for example `foo/bar`.

##### Complex Matchers

For more complex matching rules a profile matcher might additionally provide
those rules in form of a *spiff* template in the spec field `matcher`. It must
provide a field `match` (basically of type bool) for the match result. The rest
of the matcher may be any `spiff` logic. The spiff template automatically offers
the field `metadata` to access the actual metadata in the rules. As processing
stubs it gets the `values` field plus the request metadata in the
field `metadata`.

<details><summary>Or matching a dedicated type of requester by enriched metadata</summary>

```yaml
apiVersion: ipxe.mandelsoft.org/v1alpha1
kind: BootProfileMatcher
metadata:
  name: matcher1
  namespace: default
spec:
  matcher:
    match: (( .metadata.task == "node" -and contains(.metadata.__mac__, "00:00:00:00:00:00))
  profileName: node-profile
```

The results of the `selector` and the `matcher` (if present) must both indicate
a match to let the boot profile matcher match the metadata.

</details>

#### Profiles

A profile describes a set of resources that will be servered when matched by a matcher.
Here are some examples for profile definitions.

<details><summary>A profile may look like this when using go templates</summary>

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

A resource may alternatively be declared by a pattern
([regular expression](https://github.com/google/re2/wiki/Syntax))
(field `pattern`).
If a profile is searched for a resource, first the direct (path) matches are 
checked. If no direct path is found, all patterns in the given order are
checked for a match. A pattern must match the requested resource path
completely (with or without the leading slash (`/`)).

The first matching entry is used to resolve the resource
request. The processing values and the request metadata are enriched by the
match information of the resource. The field `resource-match` contains a list
of values of all matched text for the complete pattern and all sub expressions
in the order they are defined in the pattern.

For example a pattern `i(nfo)?` matches the resource `info` with the 
values `[ "info", "nfo" ]` in the field `resource-match`.

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

The URL may contain go template substitutions based on the
final processing values. The URL as well as potentially the content
of the generated URL are eventually processed with the processing values.

<details><summary>URL with substitution</summary>

```yaml
apiVersion: ipxe.mandelsoft.org/v1alpha1
kind: BootResource
metadata:
  name: kipxe-tar
  namespace: default
spec:
  mimeType: application/octet-stream
  mapping:
    org: mandelsoft
  URL: http://github.com/{{.org}}/kipxe/tarball/master
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

#### Metadata Mappers

The resource `MetaDataMapper` can be used to declare metadadata mappers executed
before the matching process as Kunernetes objects.

Two variants are supported:
- `spec.URL` if an URL is given a URL mapper is created 
- `spec.Mppping` if a mapping field is specified a Spiff mapper is created.

Additionally a weight can be set to control the processing order.
The built-in machine manager (if used) uses the weight `100`.


<details><summary>A simple spiff based Metadata Mapper</summary>

```yaml
apiVersion: ipxe.mandelsoft.org/v1alpha1
kind: MetaDataMapper
metadata:
  name: mapper
  namespace: default
spec:
  weight: 10
  mapping:
     output:
       <<<: (( .metadata ))
       mapped: (( defined(.metadata.uuid) ? "yes" :"no" ))
```

</details>

#### Mappings

The initial metadata of a request is derived from the query parameters
of the request. If a parameter is given more than once, the first value
is used as direct value. An additional list property with the name `__<name>__`
is added for every found parameter. It provides the complete list of values for
this parameter.

The requested resource name (path of the URL below the handlers root path) is
also added with the property `RESOURCE_PATH` .

This set of metadata is then mapped through the registrations for the 
[*Discovery API*](#the-discovery-api). The outcome of this mapping
is the final metadata used for the following matching process.
The Kubernetes flavor also supports metadata mappers by a dedicated resource.
It may describe a `mapping` field which is translated to a *spiff* based
mapping.

But it is also possible for the processing step of the identified resource content
at the end of the process just before delivering the content to map
metadata.

These *Processing Values* are passed through the matching chain and can be
modified and/or enriched in-between. The value structure for these 
*Processing Values* available for the processing initially is a map with
a single field `metadata`, which contains the request metadata.
 
Every matching element (matchers, profiles and resources) may define
a mapping for these processing values before they are finally used by the
resource content processing.

There are several possible ways for those *mappings* usable both mapping
scenarios described above:

- A set of *Values* are specified for the mapping element (In the Kubernetes
  objects this is the field `values`).
  Those values are used as defaults for enriching the processing values.
  It is not a deep merge, only the first level is merged. This can be used
  to enrich the processing values by defaults.
  
- A *Mapping* is given. In the API this is an implementation of the `Mapping`
  interface. It gets access to merging stubs containing the values, the initial
  processing values with the `metadat`field and the processing values resulting
  from the previous matching step.
  
  The Kubernetes objects support a field `mapping`, whose
  content is used as a *spiff* template for processing the processing data.
  It uses the values, the initial processing data (with the metadata field) and
  the previous processing data as merge stub (in this order)
  (see [the spiff doc](https://github.com/mandelsoft/spiff#structural-auto-merge)
  for merging)
  
  If a mapping is given, there is no additional value merging, the merging task
  is completely left to the mapping implementation.
  
  The default spiff based mapper (used by the Kubernetes adaption)
  supports several usage modes, all of them have implicit access to the request
  metadata via the field `metadata` and the previous chained intermediate
  processing values via the implicit field `current`.
  
  - *pure metadata modification*  
    if the template specifies the `metadata` field as non-temporary field,
    the result of the mapping is its content. This can be used, for example for
    mapping fields in metadata mappers.
  - *dedicated output field*  
    if the template specifies a field with the name `output`, the mapping result
    is the content of this field. This can be used to compose a completely new
    structure of the processing values.
  - *plain template*  
    if neither the metadata field is given, nor the output field, the content of
    the complete processing result of the template (non-temporary fields) is
    used. The implicit `metadata` field is a temporary field. This mode can be
    used to create an initial or modify a given set of processing values by
    accessing the original request metadata.

  Spiff mappings always use up to three stubs (if available) in the following order:
  - (optional) the `values` field describing default values
  - the metadata 
  - the intermediate processing values calculated by the last step (initially this
    is the request metadata)

#### The Resource Processing

After executing all mappings of the matching chain the final processing values
are used as values for processing of the resource data described by the resource
element.

Hereby the content is used as template:
- if it is yaml or json, it is used again as *spiff template* by using the
  processing values as stub for the merge process.
- if it is a text document, the go templating engine is used for processing
  with the processing values as data input

All other content types are not processed.

The task of the mappings in summary is to provide the necessary processing
context for using the described resource as template.

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

## Certificates

The ipxe server can run with http or https.
For https a TLS secret must be specified for the Kubernetes version of the
server. It can completely be maintained by the server.

The option `--certificate-mode` can be used to control this behaviour.
If set to `none`, nothing is done and no certificate is injected into the 
request metadata. If TLS is requested, nevertleless a secret or certificate
files must be given.

The server can run without TLS and maintain a TLS secret that can be used
by an ingress to let the ingress controller do the TLS termination.

If the mode is not equal to `none` (`managed`or `use`) the ca certificate is
injected into the request metadata using the field `CACERT` and can be used
to resolve resource requests (for example for a flat resource `cacert`) using
a default matcher and profile.



## Command Line Reference

```
kipxe serves iPXE requests based on Kubernetes rosources

Usage:
  kipxe [flags]

Flags:
      --bind-address-http string                         HTTP server bind address
      --cacertfile string                                kipxe server ca certificate file
      --cache-cleanup.pool.resync-period duration        Period for resynchronization for pool cache-cleanup
      --cache-cleanup.pool.size int                      Worker pool size for pool cache-cleanup
      --cache-dir string                                 enable URL caching in a dedicated directory
      --cache-ttl duration                               TTL for cache entries
      --cakeyfile string                                 kipxe server ca certificate key file
      --certfile string                                  kipxe server certificate file
      --certificate-mode string                          mode for cert management
      --config string                                    config file
  -c, --controllers string                               comma separated list of controllers to start (<name>,<group>,all) (default "all")
      --cpuprofile string                                set file for cpu profiling
      --default.pool.size int                            Worker pool size for pool default
      --disable-namespace-restriction                    disable access restriction for namespace local access only
      --grace-period duration                            inactivity grace period for detecting end of cleanup for shutdown
  -h, --help                                             help for kipxe
      --hostname stringArray                             hostname to use for kipxe registration
      --ipxe.cacertfile string                           kipxe server ca certificate file of controller ipxe
      --ipxe.cache-cleanup.pool.resync-period duration   Period for resynchronization for pool cache-cleanup of controller ipxe (default 1m0s)
      --ipxe.cache-cleanup.pool.size int                 Worker pool size for pool cache-cleanup of controller ipxe (default 1)
      --ipxe.cache-dir string                            enable URL caching in a dedicated directory of controller ipxe
      --ipxe.cache-ttl duration                          TTL for cache entries of controller ipxe (default 10m0s)
      --ipxe.cakeyfile string                            kipxe server ca certificate key file of controller ipxe
      --ipxe.certfile string                             kipxe server certificate file of controller ipxe
      --ipxe.certificate-mode string                     mode for cert management of controller ipxe (default "manage")
      --ipxe.default.pool.size int                       Worker pool size for pool default of controller ipxe (default 5)
      --ipxe.hostname stringArray                        hostname to use for kipxe registration of controller ipxe
      --ipxe.keyfile string                              kipxe server certificate key file of controller ipxe
      --ipxe.local-namespace-only                        server only resources in local namespace of controller ipxe
      --ipxe.pool.resync-period duration                 Period for resynchronization of controller ipxe
      --ipxe.pool.size int                               Worker pool size of controller ipxe
      --ipxe.pxe-port int                                pxe server port of controller ipxe (default 8081)
      --ipxe.secret string                               name of secret to maintain for kipxe server of controller ipxe
      --ipxe.service string                              name of service to use for kipxe server of controller ipxe
      --ipxe.trace-requests                              trace mapping of request data of controller ipxe
      --ipxe.use-tls                                     use https of controller ipxe
      --keyfile string                                   kipxe server certificate key file
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
      --secret string                                    name of secret to maintain for kipxe server
      --server-port-http int                             HTTP server port (serving /healthz, /metrics, ...)
      --service string                                   name of service to use for kipxe server
      --trace-requests                                   trace mapping of request data
      --use-tls                                          use https
      --version                                          version for kipxe

```

Supported controllers are:
 - `ipxe`: http server with matching engine configured by Kubernetes resources
 - `machines`: controller implementing the [*Discovery API*](#the-discovery-api) based
    on machine resources.
