# Architecture

The following sections discuss various aspects of Solvas Fabric’s architecture.

In this article:

- [Application structure](https://app.notion.com/p/Architecture-b87f90540e8d4f1faf03b8cf32b61b9c?pvs=21)
- [Service discovery (consul-registry + kube2consul/swarm2consul)](https://app.notion.com/p/Architecture-b87f90540e8d4f1faf03b8cf32b61b9c?pvs=21)
- [Event sourcing (fabric-wassup)](https://app.notion.com/p/Architecture-b87f90540e8d4f1faf03b8cf32b61b9c?pvs=21)
- [HTTP Gateway (fabric-gateway)](https://app.notion.com/p/Architecture-b87f90540e8d4f1faf03b8cf32b61b9c?pvs=21)
- [UI root (fabric-ui-root)](https://app.notion.com/p/Architecture-b87f90540e8d4f1faf03b8cf32b61b9c?pvs=21)
- [gRPC load balancing (fabric-lb)](https://app.notion.com/p/Architecture-b87f90540e8d4f1faf03b8cf32b61b9c?pvs=21)
- [gRPC APIs](https://app.notion.com/p/Architecture-b87f90540e8d4f1faf03b8cf32b61b9c?pvs=21)
- [HTTP/REST APIs](https://app.notion.com/p/Architecture-b87f90540e8d4f1faf03b8cf32b61b9c?pvs=21)
- [Authentication and Authorization (fabric-auth)](https://app.notion.com/p/Architecture-b87f90540e8d4f1faf03b8cf32b61b9c?pvs=21)
- [Summary](https://app.notion.com/p/Architecture-b87f90540e8d4f1faf03b8cf32b61b9c?pvs=21)

## Application structure

A Solvas Fabric application is packaged as a collection of [Docker](https://www.docker.com/what-docker) images and is deployed as a [Kubernetes/Helm chart](https://helm.sh/docs/developing_charts/) or a [Docker Swarm service stack](https://docs.docker.com/get-started/part5/) in a Kubernetes or Docker Swarm cluster. Services communicate with each other through an overlay network created by the container orchestration system running the cluster.

Depending on requirements and implementation details, there are one or more instances of each Solvas Fabric service running simultaneously in the cluster.

The following diagram shows all major parts of a Solvas Fabric application including the main system services:

![http___solvasdocs_sltc_com__images_architecture.png](http___solvasdocs_sltc_com__images_architecture.png)

- **fabric-gateway** - HTTP gateway application front-end.
- **consul-registry** - service registration.
- **kube2consul / swarm2consul** - synchronization between the cluster services and Consul.
- **fabric-lb** - gRPC load balancing.
- **fabric-auth** - OpenID Connect authentication and authorization service.
- **fabric-ui-root** - dynamic user interface composition engine.
- **fabric-wassup** - asynchronous event sourcing side-car service.
- **fabric-pubsub** - asynchronous publish-subscribe messaging and task scheduling service.
- **fabric-log** - monitoring and log aggregation service.
- **fabric-metrics** - metrics collection and alert service.

## Service discovery (consul-registry + kube2consul/swarm2consul)

Services discover each other by name. Since each service running in the cluster can have an arbitrary IP provided by the orchestrator’s virtual IP infrastructure, static IP addresses cannot be used as service endpoints.

Both Kubernetes and Docker Swarm come with service discovery built in, however, that feature has several flaws making it inadequate for the needs of a distributed Solvas Fabric application:

- Docker Swarm discovery offers partial endpoint registration which includes only IP address but no TCP port. The lack of port registration makes this feature deficient for the needs of a system where each service endpoint’s port is configurable, and the client of a service cannot assume which port the service is listening on.
- Docker Swarm can only register services that run within the Docker stack. It offers no way to register external services such as database servers which might run outside the Docker stack. Solvas Fabric developers need to run services outside of the Docker stack while developing/debugging them on their developer machines, yet they need those services to be discoverable by other services running in the stack.
- Kubernetes offers robust service discovery, however, tying Solvas Fabric into it would be a lock-in which would prevent us from being able to deploy Solvas Fabric workloads to Docker Swarm.

In order to avoid the above limitations, Solvas Fabric ships with and uses [HashiCorp Consul](https://www.consul.io/) as a dedicated service registry which works equally well in both Kubernetes and Docker Swarm.

![http___solvasdocs_sltc_com__images_swarm2consul.png](http___solvasdocs_sltc_com__images_swarm2consul.png)

### Solvas Fabric service discovery for Kubernetes

A special Consul Helm chart is included as a sub-chart in every Solvas Fabric Helm application.

It contains an instance of Consul, and an instance of [kube2consul](http://solvasgit.sltc.com/solvas-fabric/kube2consul) - a monitoring daemon which synchronizes the state of Kubernetes services and the Consul instance running in the cluster.

`kube2consul` performs this synchronization based on service metadata included in the application’s Helm chart.

### Solvas Fabric service discovery for Docker Swarm

A special service image is included in every Solvas Fabric Docker Swarm application stack: `docker.ftbin.io/fabric/consul-kit`.

It contains an instance of Consul, and an instance of [swarm2consul](http://solvasgit.sltc.com/solvas-fabric/swarm2consul) - a monitoring daemon which synchronizes the state of the Docker swarm services and the Consul instance running in the cluster.

`swarm2consul` performs this synchronization based on service metadata included in the application stack definition.

### Consul service endpoints

A service endpoint registration includes the following items:

- Service name
- IP address
- TCP Port
- Health-checks

### DNS

Consul server exposes a DNS interface at port `8600` where service endpoints can be queried by service name.

Solvas Fabric’s Core .NET libraries include interface `IServiceResolver` with a concrete Consul implementation, which performs a [DNS SRV record query](https://en.wikipedia.org/wiki/SRV_record) to obtain a service endpoint by name.

For example, the following code obtains the IP endpoint of service `rabbitmq`:

`1var serviceEndPoint = await serviceResolver.ResolveServiceEndPointAsync("rabbitmq").ConfigureAwait(false);`

Note that this is a simple .NET wrapper around a standard [DNS SRV record query](https://en.wikipedia.org/wiki/SRV_record). Such a DNS query can be performed by any general-purpose language in a cross-platform manner.

### Health-checks

A service can register one or more custom [Consul health-checks](https://www.consul.io/intro/getting-started/checks.html). Those health-checks can perform any number of tasks, such as attempts to communicate with the service over TCP or HTTP protocols. Each health-check is executed at an interval defined during the health-check registration.

If a service health-check is in `critical` state, DNS queries for the service SRV record return empty responses.

When a service is added to the registry, Consul automatically registers the system **Consul Serf** health-check which monitors the connectivity between the service’s Consul agent and the Consul cluster.

### External services

Image `consul-kit` includes a startup script which looks up a folder for `*.json` definitions of external services. Docker configurations can be pushed to `*.json` files in that location to register named external service endpoints which then can be accessed from services running within the cluster.

## Event sourcing (fabric-wassup)

Solvas Fabric services often need to maintain a shared view of a certain domain of objects without actually sharing data structures.

For example, a master-data management service may own (manage and organize) lookup data, but other services may need to have the same lookup data stored locally for the purposes of being able to achieve reasonable performance. Sharing the lookup data structures between the master data management service and the third-party services at the storage level is not an option as this would create a low-level service coupling.

The solution to this is called **event sourcing**. The data-owning service keeps a record of all the events that affect the state of the data it owns. A third-party service can request those events in the form of an event stream. It can then replay those events locally, storing them in its own local storage in whatever form makes sense in order to achieve its internal functional and performance goals.

![http___solvasdocs_sltc_com__images_fabric_wassup.png](http___solvasdocs_sltc_com__images_fabric_wassup.png)

[fabric-wassup](http://solvasdocs.sltc.com/docs/solvas-docs/en/latest/fabric/services/fabric-wassup/doc/index.html) is a side-car service that can be attached to any data owner service interested in emitting events. The data owner service is responsible for storing its events in a SQL Server table, while **fabric-wassup** is responsible for sharing those events in the form of event stream with the rest of the world.

**fabric-wassup** can serve hundreds of third-party services with simultaneous long-running event streams and does that with a very low performance impact. It monitors changes to the source event table and automatically updates the event streams as changes occur. The monitoring is based on SQL Server’s Query Notification infrastructure in order to avoid constant polling and maintain low profile during idle time.

### Transactionally correct eventual consistency

**fabric-wassup** helps implement *transactionally correct* eventual consistency. This kind of consistency decouples the data owner service from the third-party services yet maintains local transactional consistency within the bounded context of each service.

Transactionally correct eventual consistency is not possible when the event sourcing infrastructure is based on message brokers (such as RabbitMQ), as those cannot participate in local transactions.

**fabric-wassup** implements transactionally correct eventual consistency by serving as a resumable stream bridge between the local transaction of the data owning service and the local transaction of the receiving service.

**fabric-wassup** does not store any data. It only streams committed data. A receiving service can keep track of all the events it replays on its end in its own local transaction. It can then request to resume the event stream at any point in time based on the progress of that replay process. In case the replay process fails, the progress of the failed event will not be committed and when the receiving service restarts, it can pick up the stream from the last successfully replayed event.

### Event Scope

Each event belongs to a scope. When a receiving service requests an event stream from **fabric-wassup**, it can specify a list of scopes which it is interested in. This allows for a data owner service to serve a number of different services separated by categories.

### Integration with Entity Framework Core

Solvas Fabric Core includes class `WassupEventCollector` which integrates with **Microsoft Entity Framework Core** and helps turn change-tracked Entity framework entries into Wassup events. It supports various event storage strategies including Json storage as well as custom event storage through interface `IWassupEventStorageProvider`.

## HTTP Gateway (fabric-gateway)

External access to a Solvas Fabric application is performed though [fabric-gateway](http://solvasdocs.sltc.com/docs/solvas-docs/en/latest/fabric/services/fabric-gateway/doc/index.html) - a special Solvas Fabric service whose only purpose is to expose application functionality and assets though a single HTTP entry point.

![http___solvasdocs_sltc_com__images_fabric_gateway.png](http___solvasdocs_sltc_com__images_fabric_gateway.png)

If an application service needs to expose an HTTP endpoint through **fabric-gateway**, all it needs to do is register itself with Consul using a special tag: `gateway_route:<route>`.

**fabric-gateway** monitors the Consul registry for changes and automatically discovers services which register with that tag. For each such service, **fabric-gateway** creates a route which proxies all HTTP requests to the internal HTTP endpoint of that service.

That way each service can be developed in isolation with no knowledge of the bigger application’s routing schema. Later, it can be hot plugged dynamically into the composite application and appear at a specific route location of the application’s main HTTP entry point.

To find out more about **fabric-gateway** and how it works, see [this fabric-gateway usage article](http://solvasdocs.sltc.com/docs/solvas-docs/en/latest/fabric/services/fabric-gateway/doc/usage.html).

## UI root (fabric-ui-root)

**fabric-gateway** provides dynamic HTTP routing, but it does not offer any HTML or user interface services.

That is the concern of another Solvas Fabric service - [fabric-ui-root](http://solvasdocs.sltc.com/docs/solvas-docs/en/latest/fabric/services/fabric-ui-root/doc/index.html).

**fabric-ui-root** registers itself with **fabric-gateway** at the root route `/` and provides generic bootstrapping code for composite JavaScript single-page applications (SPA).

Just like **fabric-gateway** provides plug-in architecture for HTTP routing, **fabric-ui-root** provides plug-in architecture for SPA applications.

When loaded in a browser, **fabric-ui-root’**s bootstrapping code enumerates all services registered with **fabric-gateway** and tagged with Consul tag `fabric_ui`. Then, using [RequireJS](http://requirejs.org/), **fabric-ui-root** dynamically loads those services’ own bootstrapping code which it expects to be served by each service at URL `/js/main.js`.

Note

**fabric-ui-root** doesn’t really care what’s in those `/js/main.js` assets - it simply expects them to be there for all services registered with **fabric-gateway** and tagged with Consul as a `fabric_ui` service.

This detached approach keeps **fabric-ui-root** independent from any specific JavaScript UI framework, which, in turn, allows it to be used for bootstrapping applications based on any SPA framework, including, but not limited to:

- [AngularJS](https://angular.io/)
- [ExtJS](https://www.sencha.com/products/extjs/)
- [ReactJS](https://reactjs.net/)
- [Vue.js](https://vuejs.org/)
- [Meteor](https://www.meteor.com/)
- [Ember](https://www.emberjs.com/)
- [Aurelia.io](http://aurelia.io/)

To learn more about **fabric-ui-root**, see [fabric-ui-root usage documentation](http://solvasdocs.sltc.com/docs/solvas-docs/en/latest/fabric/services/fabric-ui-root/doc/index.html).

## gRPC load balancing (fabric-lb)

Both Kubernetes and Docker Swarm provide a routing mesh with built-in load balancing for services deployed with multiple replicas.

Unfortunately, that load-balancing does not work for gRPC services, because of the way gRPC clients and servers communicate. A gRPC client opens a single TCP connection to the server and keeps it open until the client terminates. All gRPC calls to the server are multiplexed over that single TCP connection.

When using a classic TCP load balancer like the one built into Kubernetes and Docker Swarm, a gRPC client will establish a connection with only one of the gRPC server replicas and will keep it open “forever”. This behavior circumvents the orchestrator’s load balancer algorithms which leads to situations where a gRPC client always communicates with a single server instance.

With version 1.13.10, NGINX introduced load balancing which [specifically targets the gRPC protocol](https://www.nginx.com/blog/nginx-1-13-10-grpc/).

**fabric-lb** uses that new feature to implement proper gRPC load balancing in Solvas Fabric stacks.

![http___solvasdocs_sltc_com__images_fabric_lb.png](http___solvasdocs_sltc_com__images_fabric_lb.png)

To learn more about **fabric-lb**, see [fabric-lb documentation](http://solvasdocs.sltc.com/docs/solvas-docs/en/latest/fabric/services/fabric-lb/doc/index.html).

## gRPC APIs

Solvas Fabric services communicate with each other over language-neutral, platform-neutral [gRPC](http://www.grpc.io/) APIs. The gRPC protocol offers many benefits over HTTP/REST API, including:

- 25+ times better performance due to the use of binary protocol buffers
- Bi-directional full-duplex streaming
- Integrated authentication
- Formalized error codes

A gRPC API service publishes its gRPC interface definition (IDL) as part of its documentation. For example, see service [fabric-pubsub's gRPC API definition](http://solvasdocs.sltc.com/docs/solvas-docs/en/latest/fabric/services/fabric-pubsub/doc/api-ref/pubsub.proto.html).

Client services communicate with server services through client code generated from such gRPC IDLs.

Client code can be generated for any of these languages:

- C++
- C#
- Node.js
- Java
- Python
- Go
- Ruby
- PHP
- Objective-C
- Android
- Java

In addition to their gRPC IDL, Solvas Fabric’s .NET services publish pre-built gRPC client libraries as NuGet packages. For example, see [fabric-pubsub's documentation](http://solvasdocs.sltc.com/docs/solvas-docs/en/latest/fabric/services/fabric-pubsub/doc/index.html) which can be accessed through its gRPC .NET client assembly `SolvasFabric.PubSub.GrpcApi`.

Solvas Fabric offers [PanGalactic GargleBlaster scaffolding template project](http://solvasgit.sltc.com/solvas-fabric/pangalactic-gargleblaster) as a way to jump-start the development of .NET gRPC API services.

## HTTP/REST APIs

gRPC is reserved for inter-service communication. Solvas Fabric services communicate with front-end components through automatically generated HTTP/REST APIs deployed as side-car services.

An HTTP/REST side-car service is produced by the [http-api-builder](http://solvasgit.sltc.com/solvas-fabric/http-api-builder-docker) tool chain which is invoked as part of the gRPC service build.

The REST interface is generated off of [HTTP API annotations](https://cloud.google.com/service-management/reference/rpc/google.api#http) embedded in the gRPC service `proto` file. The resulting HTTP REST/API proxy server is packaged as a standalone Docker image which can optionally be deployed alongside the gRPC service to enable access from front-end components.

![http___solvasdocs_sltc_com__images_http_api.png](http___solvasdocs_sltc_com__images_http_api.png)

<aside>
💡 Solvas Fabric offers [Tequila Sunrise scaffolding template project](http://solvasgit.sltc.com/solvas-fabric/tequila-sunrise) as a way to implement MVC-based front-end services which serve front-end components in the form of JavaScript/CSS/HTML.

As a rule of thumb, use Tequila Sunrise only of you plan to explicitly serve front-end components. For all other services use one of the gRPC service templates. If you need HTTP access to those services, simply deploy their corresponding HTTP side-car service.

</aside>

## Authentication and Authorization (fabric-auth)

Solvas Fabric supports both authentication and authorization using [Keycloak](http://www.keycloak.org/). Keycloak provides OpenID Connect support for user authentication and authorization including UMA-compliant management of permissions. Keycloak supports JSON Web Tokens (**JWTs**) including Requesting Party Tokens (**RPTs**) which contain effective permissions for the requesting party (i.e. client or end user) based on policies and permissions configured using Keycloak. Access tokens obtained by a client can be exchanged for an **RPT** either by the client or by the service that is being called. Solvas Fabric provides gRPC interceptors which can exchange access tokens for **RPTs** and evaluate permissions based on attributes applied to gRPC API implementation classes. Solvas Fabric also provides client-side interceptors for adding client credentials to gRPC calls being made from one service to another.

### Authentication

The Solvas Fabric **fabric-auth** service supports OpenID Connect compliant authentication based on external identity providers, federated login with Kerberos or LDAP, or internal authentication using its own user store.

### Authorization Terms

In OpenID Connect terms, each Solvas Fabric service is typically a **Client**, **Resource Server**, and a **Resource**. This means that the **Client** (the service) can manage **Policies** and **Permissions** which can be granted to users to access the **Resource** (the service) as controlled by the **Resource Server** (the service). Permissions are granted as **Scopes** (i.e., the action being performed such as `read` or `write`) based on the evaluations of **Policies** which can be granted based on role membership, group membership, client identity, etc.

### Realms

**Realms** represent the highest-level container for all setup within Keycloak. By default, **fabric-auth** provides a **realm** named `master`.

### Users

**Users** is a term used by Keycloak to describe an identity which can be authenticated. This identity may be a human user of the application or a **Client** application calling another service. Unless otherwise specified, the term **user** in this document can apply to the identity of an end user or a **Client**.

### Clients

Within a realm, **Clients** provide the highest level of granularity for configuring Keycloak. In Solvas Fabric terms, a **Client** is an individual service. Solvas Fabric supports self-registering of services as a **Client**. As a **Client**, each service has credentials which can be used to authenticate to Keycloak to interact with other services or to act as a **Resource Server** to perform actions relevant to that specific service on its own behalf. Web front ends and other consumers of service APIs can request access tokens which can be exchanged for **RPTs** by the service for evaluation of effective permissions with respect to the **Client**.

### Resources

**Resources** represent something that needs to be protected. For Solvas Fabric services, it is assumed that the service itself would be the lone Resource to be protected. This single-resource convention eliminates the need for configuration of specific resources for each service. If multiple resources are defined for a given **Client**, then **Resources** can be grouped by **Resource Types**.

### Scopes

**Scopes** provide fine-grained breakdown of **Resources**. For Solvas Fabric services, **Scopes** are typically actions such as `read`, `write`, and `delete`. In general terms, Scopes can be specific pieces of data used in user-managed scenarios and are presented to the user for consent for sharing with a client (ex: e-mail address, phone number, name, etc.). Solvas Fabric provides mechanisms for evaluating permissions assigned to Scopes. Authorization can be applied at the **Resource** and **Resource Type** level.

### Roles

**Roles** are classifications that can be assigned to **users**. **Roles** can either be global or specific to a **Client**.

### Policies

**Policies** are rule sets that can be applied to a specific user. For example, a client may have a **policy** that is granted only if a user is assigned a specific role.

### Permissions

**Permissions** are the intersections of **users** and **Policies** and can be applied to **Resource Types**, **Resources**, or **Scopes**. For Solvas Fabric services, **policies** are typically applied to **Scopes** such that each service call can determine if the user has the permissions to perform the action (**Scope**) requested such as `read` or `write`. A **Permission** can apply multiple **Policies** and there are various decision strategies available to determine whether or not the user is granted the permission depending on the success or failure of the associated **policies**.

### Access Token

**Access Tokens** are **JWTs** issued to represent an authenticated identity which could be either and end user or **Client**.

### Protection API Token (PAT)

**Protection API Tokens** or **PATs** are a special OAuth2 access token with a scope defined as `uma_protection`. These tokens can be obtained by clients by providing client credentials consisting of a **Client** ID and secret.

### Requesting Party Token (RPT)

**Requesting Party Tokens** or **RPTs** are digitally signed **JWTs** based on an OAuth2 **access token**. **RPTs** include any effective permissions in a claim named `authorization`.

### Client Registration

Solvas Fabric services are able to register themselves as **Clients** with the **fabric-auth** service. This allows the service to have complete control over the **Resources**, **Scopes**, **Policies**, and **Permissions** used to control access to the service. Since **policies** can include client-specific **roles**, these are created automatically through the registration process.

Clients are able to access the Keycloak Resource and Protection APIs. This allows clients to create additional **resources** and **policies** on demand for fine-grande authorization (i.e. row-based security).

### Authorization Process

The typical flow for authorization of Fabric services is as follows:

![http___solvasdocs_sltc_com__images_fabric_auth.png](http___solvasdocs_sltc_com__images_fabric_auth.png)

The user performs an interactive login which results in the user interface obtaining an **Access Token** for the default front-end client `fabric-ui`.

- The front-end application requests data from a back-end service. This back-end service has access to credentials which allow it to call other services using a **PAT** for the default back-end client `fabric-backend`.
- The service can exchange the **Access Token** provided by the front end for an **RPT** and evaluate the permissions contained in the **RPT** claims.

### Summary

The article covered the most significant aspects of Solvas Fabric’s architecture.