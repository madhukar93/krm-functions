package workloads

var workloadscrd = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: (devel)
  creationTimestamp: null
  name: functionconfigs.krm
spec:
  group: krm
  names:
    kind: FunctionConfig
    listKind: FunctionConfigList
    plural: functionconfigs
    singular: functionconfig
  scope: Namespaced
  versions:
  - name: workloads
    schema:
      openAPIV3Schema:
        properties:
          apiVersion:
            description: 'APIVersion defines the versioned schema of this representation
              of an object. Servers should convert recognized schemas to the latest
              internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
            type: string
          kind:
            description: 'Kind is a string value representing the REST resource this
              object represents. Servers may infer this from the endpoint the client
              submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
            type: string
          metadata:
            type: object
          spec:
            properties:
              app:
                type: string
              containers:
                items:
                  properties:
                    args:
                      description: 'Arguments to the entrypoint. The container image''s
                        CMD is used if this is not provided. Variable references $(VAR_NAME)
                        are expanded using the container''s environment. If a variable
                        cannot be resolved, the reference in the input string will
                        be unchanged. Double $$ are reduced to a single $, which allows
                        for escaping the $(VAR_NAME) syntax: i.e. "$$(VAR_NAME)" will
                        produce the string literal "$(VAR_NAME)". Escaped references
                        will never be expanded, regardless of whether the variable
                        exists or not. Cannot be updated. More info: https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell'
                      items:
                        type: string
                      type: array
                    command:
                      description: 'Entrypoint array. Not executed within a shell.
                        The container image''s ENTRYPOINT is used if this is not provided.
                        Variable references $(VAR_NAME) are expanded using the container''s
                        environment. If a variable cannot be resolved, the reference
                        in the input string will be unchanged. Double $$ are reduced
                        to a single $, which allows for escaping the $(VAR_NAME) syntax:
                        i.e. "$$(VAR_NAME)" will produce the string literal "$(VAR_NAME)".
                        Escaped references will never be expanded, regardless of whether
                        the variable exists or not. Cannot be updated. More info:
                        https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell'
                      items:
                        type: string
                      type: array
                    configs:
                      description: this is docker compose-ish if these fields are
                        populated, they augment the container if they are not populated,
                        the container is used as is if they are populated, the container
                        is used as a base and the fields are applied on top if they
                        are populated, the container is used as a base and the fields
                        are applied on top
                      items:
                        type: string
                      type: array
                    env:
                      description: List of environment variables to set in the container.
                        Cannot be updated.
                      items:
                        description: EnvVar represents an environment variable present
                          in a Container.
                        properties:
                          name:
                            description: Name of the environment variable. Must be
                              a C_IDENTIFIER.
                            type: string
                          value:
                            description: 'Variable references $(VAR_NAME) are expanded
                              using the previously defined environment variables in
                              the container and any service environment variables.
                              If a variable cannot be resolved, the reference in the
                              input string will be unchanged. Double $$ are reduced
                              to a single $, which allows for escaping the $(VAR_NAME)
                              syntax: i.e. "$$(VAR_NAME)" will produce the string
                              literal "$(VAR_NAME)". Escaped references will never
                              be expanded, regardless of whether the variable exists
                              or not. Defaults to "".'
                            type: string
                          valueFrom:
                            description: Source for the environment variable's value.
                              Cannot be used if value is not empty.
                            properties:
                              configMapKeyRef:
                                description: Selects a key of a ConfigMap.
                                properties:
                                  key:
                                    description: The key to select.
                                    type: string
                                  name:
                                    description: 'Name of the referent. More info:
                                      https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
                                      TODO: Add other useful fields. apiVersion, kind,
                                      uid?'
                                    type: string
                                  optional:
                                    description: Specify whether the ConfigMap or
                                      its key must be defined
                                    type: boolean
                                required:
                                - key
                                type: object
                                x-kubernetes-map-type: atomic
                              fieldRef:
                                description: 'Selects a field of the pod: supports
                                  metadata.name, metadata.namespace, spec.nodeName,
                                  spec.serviceAccountName, status.hostIP, status.podIP,
                                  status.podIPs.'
                                properties:
                                  apiVersion:
                                    description: Version of the schema the FieldPath
                                      is written in terms of, defaults to "v1".
                                    type: string
                                  fieldPath:
                                    description: Path of the field to select in the
                                      specified API version.
                                    type: string
                                required:
                                - fieldPath
                                type: object
                                x-kubernetes-map-type: atomic
                              resourceFieldRef:
                                description: 'Selects a resource of the container:
                                  only resources limits and requests (limits.cpu,
                                  limits.memory, limits.ephemeral-storage, requests.cpu,
                                  requests.memory and requests.ephemeral-storage)
                                  are currently supported.'
                                properties:
                                  containerName:
                                    description: 'Container name: required for volumes,
                                      optional for env vars'
                                    type: string
                                  divisor:
                                    anyOf:
                                    - type: integer
                                    - type: string
                                    description: Specifies the output format of the
                                      exposed resources, defaults to "1"
                                    pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                    x-kubernetes-int-or-string: true
                                  resource:
                                    description: 'Required: resource to select'
                                    type: string
                                required:
                                - resource
                                type: object
                                x-kubernetes-map-type: atomic
                              secretKeyRef:
                                description: Selects a key of a secret in the pod's
                                  namespace
                                properties:
                                  key:
                                    description: The key of the secret to select from.  Must
                                      be a valid secret key.
                                    type: string
                                  name:
                                    description: 'Name of the referent. More info:
                                      https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
                                      TODO: Add other useful fields. apiVersion, kind,
                                      uid?'
                                    type: string
                                  optional:
                                    description: Specify whether the Secret or its
                                      key must be defined
                                    type: boolean
                                required:
                                - key
                                type: object
                                x-kubernetes-map-type: atomic
                            type: object
                        required:
                        - name
                        type: object
                      type: array
                    envFrom:
                      description: List of sources to populate environment variables
                        in the container. The keys defined within a source must be
                        a C_IDENTIFIER. All invalid keys will be reported as an event
                        when the container is starting. When a key exists in multiple
                        sources, the value associated with the last source will take
                        precedence. Values defined by an Env with a duplicate key
                        will take precedence. Cannot be updated.
                      items:
                        description: EnvFromSource represents the source of a set
                          of ConfigMaps
                        properties:
                          configMapRef:
                            description: The ConfigMap to select from
                            properties:
                              name:
                                description: 'Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
                                  TODO: Add other useful fields. apiVersion, kind,
                                  uid?'
                                type: string
                              optional:
                                description: Specify whether the ConfigMap must be
                                  defined
                                type: boolean
                            type: object
                            x-kubernetes-map-type: atomic
                          prefix:
                            description: An optional identifier to prepend to each
                              key in the ConfigMap. Must be a C_IDENTIFIER.
                            type: string
                          secretRef:
                            description: The Secret to select from
                            properties:
                              name:
                                description: 'Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
                                  TODO: Add other useful fields. apiVersion, kind,
                                  uid?'
                                type: string
                              optional:
                                description: Specify whether the Secret must be defined
                                type: boolean
                            type: object
                            x-kubernetes-map-type: atomic
                        type: object
                      type: array
                    grpc:
                      properties:
                        port:
                          format: int32
                          type: integer
                      required:
                      - port
                      type: object
                    http:
                      properties:
                        port:
                          format: int32
                          type: integer
                      required:
                      - port
                      type: object
                    image:
                      description: 'Container image name. More info: https://kubernetes.io/docs/concepts/containers/images
                        This field is optional to allow higher level config management
                        to default or override container images in workload controllers
                        like Deployments and StatefulSets.'
                      type: string
                    imagePullPolicy:
                      description: 'Image pull policy. One of Always, Never, IfNotPresent.
                        Defaults to Always if :latest tag is specified, or IfNotPresent
                        otherwise. Cannot be updated. More info: https://kubernetes.io/docs/concepts/containers/images#updating-images'
                      type: string
                    lifecycle:
                      description: Actions that the management system should take
                        in response to container lifecycle events. Cannot be updated.
                      properties:
                        postStart:
                          description: 'PostStart is called immediately after a container
                            is created. If the handler fails, the container is terminated
                            and restarted according to its restart policy. Other management
                            of the container blocks until the hook completes. More
                            info: https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks/#container-hooks'
                          properties:
                            exec:
                              description: Exec specifies the action to take.
                              properties:
                                command:
                                  description: Command is the command line to execute
                                    inside the container, the working directory for
                                    the command  is root ('/') in the container's
                                    filesystem. The command is simply exec'd, it is
                                    not run inside a shell, so traditional shell instructions
                                    ('|', etc) won't work. To use a shell, you need
                                    to explicitly call out to that shell. Exit status
                                    of 0 is treated as live/healthy and non-zero is
                                    unhealthy.
                                  items:
                                    type: string
                                  type: array
                              type: object
                            httpGet:
                              description: HTTPGet specifies the http request to perform.
                              properties:
                                host:
                                  description: Host name to connect to, defaults to
                                    the pod IP. You probably want to set "Host" in
                                    httpHeaders instead.
                                  type: string
                                httpHeaders:
                                  description: Custom headers to set in the request.
                                    HTTP allows repeated headers.
                                  items:
                                    description: HTTPHeader describes a custom header
                                      to be used in HTTP probes
                                    properties:
                                      name:
                                        description: The header field name
                                        type: string
                                      value:
                                        description: The header field value
                                        type: string
                                    required:
                                    - name
                                    - value
                                    type: object
                                  type: array
                                path:
                                  description: Path to access on the HTTP server.
                                  type: string
                                port:
                                  anyOf:
                                  - type: integer
                                  - type: string
                                  description: Name or number of the port to access
                                    on the container. Number must be in the range
                                    1 to 65535. Name must be an IANA_SVC_NAME.
                                  x-kubernetes-int-or-string: true
                                scheme:
                                  description: Scheme to use for connecting to the
                                    host. Defaults to HTTP.
                                  type: string
                              required:
                              - port
                              type: object
                            tcpSocket:
                              description: Deprecated. TCPSocket is NOT supported
                                as a LifecycleHandler and kept for the backward compatibility.
                                There are no validation of this field and lifecycle
                                hooks will fail in runtime when tcp handler is specified.
                              properties:
                                host:
                                  description: 'Optional: Host name to connect to,
                                    defaults to the pod IP.'
                                  type: string
                                port:
                                  anyOf:
                                  - type: integer
                                  - type: string
                                  description: Number or name of the port to access
                                    on the container. Number must be in the range
                                    1 to 65535. Name must be an IANA_SVC_NAME.
                                  x-kubernetes-int-or-string: true
                              required:
                              - port
                              type: object
                          type: object
                        preStop:
                          description: 'PreStop is called immediately before a container
                            is terminated due to an API request or management event
                            such as liveness/startup probe failure, preemption, resource
                            contention, etc. The handler is not called if the container
                            crashes or exits. The Pod''s termination grace period
                            countdown begins before the PreStop hook is executed.
                            Regardless of the outcome of the handler, the container
                            will eventually terminate within the Pod''s termination
                            grace period (unless delayed by finalizers). Other management
                            of the container blocks until the hook completes or until
                            the termination grace period is reached. More info: https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks/#container-hooks'
                          properties:
                            exec:
                              description: Exec specifies the action to take.
                              properties:
                                command:
                                  description: Command is the command line to execute
                                    inside the container, the working directory for
                                    the command  is root ('/') in the container's
                                    filesystem. The command is simply exec'd, it is
                                    not run inside a shell, so traditional shell instructions
                                    ('|', etc) won't work. To use a shell, you need
                                    to explicitly call out to that shell. Exit status
                                    of 0 is treated as live/healthy and non-zero is
                                    unhealthy.
                                  items:
                                    type: string
                                  type: array
                              type: object
                            httpGet:
                              description: HTTPGet specifies the http request to perform.
                              properties:
                                host:
                                  description: Host name to connect to, defaults to
                                    the pod IP. You probably want to set "Host" in
                                    httpHeaders instead.
                                  type: string
                                httpHeaders:
                                  description: Custom headers to set in the request.
                                    HTTP allows repeated headers.
                                  items:
                                    description: HTTPHeader describes a custom header
                                      to be used in HTTP probes
                                    properties:
                                      name:
                                        description: The header field name
                                        type: string
                                      value:
                                        description: The header field value
                                        type: string
                                    required:
                                    - name
                                    - value
                                    type: object
                                  type: array
                                path:
                                  description: Path to access on the HTTP server.
                                  type: string
                                port:
                                  anyOf:
                                  - type: integer
                                  - type: string
                                  description: Name or number of the port to access
                                    on the container. Number must be in the range
                                    1 to 65535. Name must be an IANA_SVC_NAME.
                                  x-kubernetes-int-or-string: true
                                scheme:
                                  description: Scheme to use for connecting to the
                                    host. Defaults to HTTP.
                                  type: string
                              required:
                              - port
                              type: object
                            tcpSocket:
                              description: Deprecated. TCPSocket is NOT supported
                                as a LifecycleHandler and kept for the backward compatibility.
                                There are no validation of this field and lifecycle
                                hooks will fail in runtime when tcp handler is specified.
                              properties:
                                host:
                                  description: 'Optional: Host name to connect to,
                                    defaults to the pod IP.'
                                  type: string
                                port:
                                  anyOf:
                                  - type: integer
                                  - type: string
                                  description: Number or name of the port to access
                                    on the container. Number must be in the range
                                    1 to 65535. Name must be an IANA_SVC_NAME.
                                  x-kubernetes-int-or-string: true
                              required:
                              - port
                              type: object
                          type: object
                      type: object
                    livenessProbe:
                      description: 'Periodic probe of container liveness. Container
                        will be restarted if the probe fails. Cannot be updated. More
                        info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                      properties:
                        exec:
                          description: Exec specifies the action to take.
                          properties:
                            command:
                              description: Command is the command line to execute
                                inside the container, the working directory for the
                                command  is root ('/') in the container's filesystem.
                                The command is simply exec'd, it is not run inside
                                a shell, so traditional shell instructions ('|', etc)
                                won't work. To use a shell, you need to explicitly
                                call out to that shell. Exit status of 0 is treated
                                as live/healthy and non-zero is unhealthy.
                              items:
                                type: string
                              type: array
                          type: object
                        failureThreshold:
                          description: Minimum consecutive failures for the probe
                            to be considered failed after having succeeded. Defaults
                            to 3. Minimum value is 1.
                          format: int32
                          type: integer
                        grpc:
                          description: GRPC specifies an action involving a GRPC port.
                            This is a beta field and requires enabling GRPCContainerProbe
                            feature gate.
                          properties:
                            port:
                              description: Port number of the gRPC service. Number
                                must be in the range 1 to 65535.
                              format: int32
                              type: integer
                            service:
                              description: "Service is the name of the service to
                                place in the gRPC HealthCheckRequest (see https://github.com/grpc/grpc/blob/master/doc/health-checking.md).
                                \n If this is not specified, the default behavior
                                is defined by gRPC."
                              type: string
                          required:
                          - port
                          type: object
                        httpGet:
                          description: HTTPGet specifies the http request to perform.
                          properties:
                            host:
                              description: Host name to connect to, defaults to the
                                pod IP. You probably want to set "Host" in httpHeaders
                                instead.
                              type: string
                            httpHeaders:
                              description: Custom headers to set in the request. HTTP
                                allows repeated headers.
                              items:
                                description: HTTPHeader describes a custom header
                                  to be used in HTTP probes
                                properties:
                                  name:
                                    description: The header field name
                                    type: string
                                  value:
                                    description: The header field value
                                    type: string
                                required:
                                - name
                                - value
                                type: object
                              type: array
                            path:
                              description: Path to access on the HTTP server.
                              type: string
                            port:
                              anyOf:
                              - type: integer
                              - type: string
                              description: Name or number of the port to access on
                                the container. Number must be in the range 1 to 65535.
                                Name must be an IANA_SVC_NAME.
                              x-kubernetes-int-or-string: true
                            scheme:
                              description: Scheme to use for connecting to the host.
                                Defaults to HTTP.
                              type: string
                          required:
                          - port
                          type: object
                        initialDelaySeconds:
                          description: 'Number of seconds after the container has
                            started before liveness probes are initiated. More info:
                            https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                          format: int32
                          type: integer
                        periodSeconds:
                          description: How often (in seconds) to perform the probe.
                            Default to 10 seconds. Minimum value is 1.
                          format: int32
                          type: integer
                        successThreshold:
                          description: Minimum consecutive successes for the probe
                            to be considered successful after having failed. Defaults
                            to 1. Must be 1 for liveness and startup. Minimum value
                            is 1.
                          format: int32
                          type: integer
                        tcpSocket:
                          description: TCPSocket specifies an action involving a TCP
                            port.
                          properties:
                            host:
                              description: 'Optional: Host name to connect to, defaults
                                to the pod IP.'
                              type: string
                            port:
                              anyOf:
                              - type: integer
                              - type: string
                              description: Number or name of the port to access on
                                the container. Number must be in the range 1 to 65535.
                                Name must be an IANA_SVC_NAME.
                              x-kubernetes-int-or-string: true
                          required:
                          - port
                          type: object
                        terminationGracePeriodSeconds:
                          description: Optional duration in seconds the pod needs
                            to terminate gracefully upon probe failure. The grace
                            period is the duration in seconds after the processes
                            running in the pod are sent a termination signal and the
                            time when the processes are forcibly halted with a kill
                            signal. Set this value longer than the expected cleanup
                            time for your process. If this value is nil, the pod's
                            terminationGracePeriodSeconds will be used. Otherwise,
                            this value overrides the value provided by the pod spec.
                            Value must be non-negative integer. The value zero indicates
                            stop immediately via the kill signal (no opportunity to
                            shut down). This is a beta field and requires enabling
                            ProbeTerminationGracePeriod feature gate. Minimum value
                            is 1. spec.terminationGracePeriodSeconds is used if unset.
                          format: int64
                          type: integer
                        timeoutSeconds:
                          description: 'Number of seconds after which the probe times
                            out. Defaults to 1 second. Minimum value is 1. More info:
                            https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                          format: int32
                          type: integer
                      type: object
                    name:
                      description: Name of the container specified as a DNS_LABEL.
                        Each container in a pod must have a unique name (DNS_LABEL).
                        Cannot be updated.
                      type: string
                    ports:
                      description: List of ports to expose from the container. Not
                        specifying a port here DOES NOT prevent that port from being
                        exposed. Any port which is listening on the default "0.0.0.0"
                        address inside a container will be accessible from the network.
                        Modifying this array with strategic merge patch may corrupt
                        the data. For more information See https://github.com/kubernetes/kubernetes/issues/108255.
                        Cannot be updated.
                      items:
                        description: ContainerPort represents a network port in a
                          single container.
                        properties:
                          containerPort:
                            description: Number of port to expose on the pod's IP
                              address. This must be a valid port number, 0 < x < 65536.
                            format: int32
                            type: integer
                          hostIP:
                            description: What host IP to bind the external port to.
                            type: string
                          hostPort:
                            description: Number of port to expose on the host. If
                              specified, this must be a valid port number, 0 < x <
                              65536. If HostNetwork is specified, this must match
                              ContainerPort. Most containers do not need this.
                            format: int32
                            type: integer
                          name:
                            description: If specified, this must be an IANA_SVC_NAME
                              and unique within the pod. Each named port in a pod
                              must have a unique name. Name for the port that can
                              be referred to by services.
                            type: string
                          protocol:
                            default: TCP
                            description: Protocol for port. Must be UDP, TCP, or SCTP.
                              Defaults to "TCP".
                            type: string
                        required:
                        - containerPort
                        type: object
                      type: array
                      x-kubernetes-list-map-keys:
                      - containerPort
                      - protocol
                      x-kubernetes-list-type: map
                    readinessProbe:
                      description: 'Periodic probe of container service readiness.
                        Container will be removed from service endpoints if the probe
                        fails. Cannot be updated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                      properties:
                        exec:
                          description: Exec specifies the action to take.
                          properties:
                            command:
                              description: Command is the command line to execute
                                inside the container, the working directory for the
                                command  is root ('/') in the container's filesystem.
                                The command is simply exec'd, it is not run inside
                                a shell, so traditional shell instructions ('|', etc)
                                won't work. To use a shell, you need to explicitly
                                call out to that shell. Exit status of 0 is treated
                                as live/healthy and non-zero is unhealthy.
                              items:
                                type: string
                              type: array
                          type: object
                        failureThreshold:
                          description: Minimum consecutive failures for the probe
                            to be considered failed after having succeeded. Defaults
                            to 3. Minimum value is 1.
                          format: int32
                          type: integer
                        grpc:
                          description: GRPC specifies an action involving a GRPC port.
                            This is a beta field and requires enabling GRPCContainerProbe
                            feature gate.
                          properties:
                            port:
                              description: Port number of the gRPC service. Number
                                must be in the range 1 to 65535.
                              format: int32
                              type: integer
                            service:
                              description: "Service is the name of the service to
                                place in the gRPC HealthCheckRequest (see https://github.com/grpc/grpc/blob/master/doc/health-checking.md).
                                \n If this is not specified, the default behavior
                                is defined by gRPC."
                              type: string
                          required:
                          - port
                          type: object
                        httpGet:
                          description: HTTPGet specifies the http request to perform.
                          properties:
                            host:
                              description: Host name to connect to, defaults to the
                                pod IP. You probably want to set "Host" in httpHeaders
                                instead.
                              type: string
                            httpHeaders:
                              description: Custom headers to set in the request. HTTP
                                allows repeated headers.
                              items:
                                description: HTTPHeader describes a custom header
                                  to be used in HTTP probes
                                properties:
                                  name:
                                    description: The header field name
                                    type: string
                                  value:
                                    description: The header field value
                                    type: string
                                required:
                                - name
                                - value
                                type: object
                              type: array
                            path:
                              description: Path to access on the HTTP server.
                              type: string
                            port:
                              anyOf:
                              - type: integer
                              - type: string
                              description: Name or number of the port to access on
                                the container. Number must be in the range 1 to 65535.
                                Name must be an IANA_SVC_NAME.
                              x-kubernetes-int-or-string: true
                            scheme:
                              description: Scheme to use for connecting to the host.
                                Defaults to HTTP.
                              type: string
                          required:
                          - port
                          type: object
                        initialDelaySeconds:
                          description: 'Number of seconds after the container has
                            started before liveness probes are initiated. More info:
                            https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                          format: int32
                          type: integer
                        periodSeconds:
                          description: How often (in seconds) to perform the probe.
                            Default to 10 seconds. Minimum value is 1.
                          format: int32
                          type: integer
                        successThreshold:
                          description: Minimum consecutive successes for the probe
                            to be considered successful after having failed. Defaults
                            to 1. Must be 1 for liveness and startup. Minimum value
                            is 1.
                          format: int32
                          type: integer
                        tcpSocket:
                          description: TCPSocket specifies an action involving a TCP
                            port.
                          properties:
                            host:
                              description: 'Optional: Host name to connect to, defaults
                                to the pod IP.'
                              type: string
                            port:
                              anyOf:
                              - type: integer
                              - type: string
                              description: Number or name of the port to access on
                                the container. Number must be in the range 1 to 65535.
                                Name must be an IANA_SVC_NAME.
                              x-kubernetes-int-or-string: true
                          required:
                          - port
                          type: object
                        terminationGracePeriodSeconds:
                          description: Optional duration in seconds the pod needs
                            to terminate gracefully upon probe failure. The grace
                            period is the duration in seconds after the processes
                            running in the pod are sent a termination signal and the
                            time when the processes are forcibly halted with a kill
                            signal. Set this value longer than the expected cleanup
                            time for your process. If this value is nil, the pod's
                            terminationGracePeriodSeconds will be used. Otherwise,
                            this value overrides the value provided by the pod spec.
                            Value must be non-negative integer. The value zero indicates
                            stop immediately via the kill signal (no opportunity to
                            shut down). This is a beta field and requires enabling
                            ProbeTerminationGracePeriod feature gate. Minimum value
                            is 1. spec.terminationGracePeriodSeconds is used if unset.
                          format: int64
                          type: integer
                        timeoutSeconds:
                          description: 'Number of seconds after which the probe times
                            out. Defaults to 1 second. Minimum value is 1. More info:
                            https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                          format: int32
                          type: integer
                      type: object
                    resources:
                      description: 'Compute Resources required by this container.
                        Cannot be updated. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                      properties:
                        limits:
                          additionalProperties:
                            anyOf:
                            - type: integer
                            - type: string
                            pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                            x-kubernetes-int-or-string: true
                          description: 'Limits describes the maximum amount of compute
                            resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                          type: object
                        requests:
                          additionalProperties:
                            anyOf:
                            - type: integer
                            - type: string
                            pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                            x-kubernetes-int-or-string: true
                          description: 'Requests describes the minimum amount of compute
                            resources required. If Requests is omitted for a container,
                            it defaults to Limits if that is explicitly specified,
                            otherwise to an implementation-defined value. More info:
                            https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                          type: object
                      type: object
                    secrets:
                      items:
                        type: string
                      type: array
                    securityContext:
                      description: 'SecurityContext defines the security options the
                        container should be run with. If set, the fields of SecurityContext
                        override the equivalent fields of PodSecurityContext. More
                        info: https://kubernetes.io/docs/tasks/configure-pod-container/security-context/'
                      properties:
                        allowPrivilegeEscalation:
                          description: 'AllowPrivilegeEscalation controls whether
                            a process can gain more privileges than its parent process.
                            This bool directly controls if the no_new_privs flag will
                            be set on the container process. AllowPrivilegeEscalation
                            is true always when the container is: 1) run as Privileged
                            2) has CAP_SYS_ADMIN Note that this field cannot be set
                            when spec.os.name is windows.'
                          type: boolean
                        capabilities:
                          description: The capabilities to add/drop when running containers.
                            Defaults to the default set of capabilities granted by
                            the container runtime. Note that this field cannot be
                            set when spec.os.name is windows.
                          properties:
                            add:
                              description: Added capabilities
                              items:
                                description: Capability represent POSIX capabilities
                                  type
                                type: string
                              type: array
                            drop:
                              description: Removed capabilities
                              items:
                                description: Capability represent POSIX capabilities
                                  type
                                type: string
                              type: array
                          type: object
                        privileged:
                          description: Run container in privileged mode. Processes
                            in privileged containers are essentially equivalent to
                            root on the host. Defaults to false. Note that this field
                            cannot be set when spec.os.name is windows.
                          type: boolean
                        procMount:
                          description: procMount denotes the type of proc mount to
                            use for the containers. The default is DefaultProcMount
                            which uses the container runtime defaults for readonly
                            paths and masked paths. This requires the ProcMountType
                            feature flag to be enabled. Note that this field cannot
                            be set when spec.os.name is windows.
                          type: string
                        readOnlyRootFilesystem:
                          description: Whether this container has a read-only root
                            filesystem. Default is false. Note that this field cannot
                            be set when spec.os.name is windows.
                          type: boolean
                        runAsGroup:
                          description: The GID to run the entrypoint of the container
                            process. Uses runtime default if unset. May also be set
                            in PodSecurityContext.  If set in both SecurityContext
                            and PodSecurityContext, the value specified in SecurityContext
                            takes precedence. Note that this field cannot be set when
                            spec.os.name is windows.
                          format: int64
                          type: integer
                        runAsNonRoot:
                          description: Indicates that the container must run as a
                            non-root user. If true, the Kubelet will validate the
                            image at runtime to ensure that it does not run as UID
                            0 (root) and fail to start the container if it does. If
                            unset or false, no such validation will be performed.
                            May also be set in PodSecurityContext.  If set in both
                            SecurityContext and PodSecurityContext, the value specified
                            in SecurityContext takes precedence.
                          type: boolean
                        runAsUser:
                          description: The UID to run the entrypoint of the container
                            process. Defaults to user specified in image metadata
                            if unspecified. May also be set in PodSecurityContext.  If
                            set in both SecurityContext and PodSecurityContext, the
                            value specified in SecurityContext takes precedence. Note
                            that this field cannot be set when spec.os.name is windows.
                          format: int64
                          type: integer
                        seLinuxOptions:
                          description: The SELinux context to be applied to the container.
                            If unspecified, the container runtime will allocate a
                            random SELinux context for each container.  May also be
                            set in PodSecurityContext.  If set in both SecurityContext
                            and PodSecurityContext, the value specified in SecurityContext
                            takes precedence. Note that this field cannot be set when
                            spec.os.name is windows.
                          properties:
                            level:
                              description: Level is SELinux level label that applies
                                to the container.
                              type: string
                            role:
                              description: Role is a SELinux role label that applies
                                to the container.
                              type: string
                            type:
                              description: Type is a SELinux type label that applies
                                to the container.
                              type: string
                            user:
                              description: User is a SELinux user label that applies
                                to the container.
                              type: string
                          type: object
                        seccompProfile:
                          description: The seccomp options to use by this container.
                            If seccomp options are provided at both the pod & container
                            level, the container options override the pod options.
                            Note that this field cannot be set when spec.os.name is
                            windows.
                          properties:
                            localhostProfile:
                              description: localhostProfile indicates a profile defined
                                in a file on the node should be used. The profile
                                must be preconfigured on the node to work. Must be
                                a descending path, relative to the kubelet's configured
                                seccomp profile location. Must only be set if type
                                is "Localhost".
                              type: string
                            type:
                              description: "type indicates which kind of seccomp profile
                                will be applied. Valid options are: \n Localhost -
                                a profile defined in a file on the node should be
                                used. RuntimeDefault - the container runtime default
                                profile should be used. Unconfined - no profile should
                                be applied."
                              type: string
                          required:
                          - type
                          type: object
                        windowsOptions:
                          description: The Windows specific settings applied to all
                            containers. If unspecified, the options from the PodSecurityContext
                            will be used. If set in both SecurityContext and PodSecurityContext,
                            the value specified in SecurityContext takes precedence.
                            Note that this field cannot be set when spec.os.name is
                            linux.
                          properties:
                            gmsaCredentialSpec:
                              description: GMSACredentialSpec is where the GMSA admission
                                webhook (https://github.com/kubernetes-sigs/windows-gmsa)
                                inlines the contents of the GMSA credential spec named
                                by the GMSACredentialSpecName field.
                              type: string
                            gmsaCredentialSpecName:
                              description: GMSACredentialSpecName is the name of the
                                GMSA credential spec to use.
                              type: string
                            hostProcess:
                              description: HostProcess determines if a container should
                                be run as a 'Host Process' container. This field is
                                alpha-level and will only be honored by components
                                that enable the WindowsHostProcessContainers feature
                                flag. Setting this field without the feature flag
                                will result in errors when validating the Pod. All
                                of a Pod's containers must have the same effective
                                HostProcess value (it is not allowed to have a mix
                                of HostProcess containers and non-HostProcess containers).  In
                                addition, if HostProcess is true then HostNetwork
                                must also be set to true.
                              type: boolean
                            runAsUserName:
                              description: The UserName in Windows to run the entrypoint
                                of the container process. Defaults to the user specified
                                in image metadata if unspecified. May also be set
                                in PodSecurityContext. If set in both SecurityContext
                                and PodSecurityContext, the value specified in SecurityContext
                                takes precedence.
                              type: string
                          type: object
                      type: object
                    startupProbe:
                      description: 'StartupProbe indicates that the Pod has successfully
                        initialized. If specified, no other probes are executed until
                        this completes successfully. If this probe fails, the Pod
                        will be restarted, just as if the livenessProbe failed. This
                        can be used to provide different probe parameters at the beginning
                        of a Pod''s lifecycle, when it might take a long time to load
                        data or warm a cache, than during steady-state operation.
                        This cannot be updated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                      properties:
                        exec:
                          description: Exec specifies the action to take.
                          properties:
                            command:
                              description: Command is the command line to execute
                                inside the container, the working directory for the
                                command  is root ('/') in the container's filesystem.
                                The command is simply exec'd, it is not run inside
                                a shell, so traditional shell instructions ('|', etc)
                                won't work. To use a shell, you need to explicitly
                                call out to that shell. Exit status of 0 is treated
                                as live/healthy and non-zero is unhealthy.
                              items:
                                type: string
                              type: array
                          type: object
                        failureThreshold:
                          description: Minimum consecutive failures for the probe
                            to be considered failed after having succeeded. Defaults
                            to 3. Minimum value is 1.
                          format: int32
                          type: integer
                        grpc:
                          description: GRPC specifies an action involving a GRPC port.
                            This is a beta field and requires enabling GRPCContainerProbe
                            feature gate.
                          properties:
                            port:
                              description: Port number of the gRPC service. Number
                                must be in the range 1 to 65535.
                              format: int32
                              type: integer
                            service:
                              description: "Service is the name of the service to
                                place in the gRPC HealthCheckRequest (see https://github.com/grpc/grpc/blob/master/doc/health-checking.md).
                                \n If this is not specified, the default behavior
                                is defined by gRPC."
                              type: string
                          required:
                          - port
                          type: object
                        httpGet:
                          description: HTTPGet specifies the http request to perform.
                          properties:
                            host:
                              description: Host name to connect to, defaults to the
                                pod IP. You probably want to set "Host" in httpHeaders
                                instead.
                              type: string
                            httpHeaders:
                              description: Custom headers to set in the request. HTTP
                                allows repeated headers.
                              items:
                                description: HTTPHeader describes a custom header
                                  to be used in HTTP probes
                                properties:
                                  name:
                                    description: The header field name
                                    type: string
                                  value:
                                    description: The header field value
                                    type: string
                                required:
                                - name
                                - value
                                type: object
                              type: array
                            path:
                              description: Path to access on the HTTP server.
                              type: string
                            port:
                              anyOf:
                              - type: integer
                              - type: string
                              description: Name or number of the port to access on
                                the container. Number must be in the range 1 to 65535.
                                Name must be an IANA_SVC_NAME.
                              x-kubernetes-int-or-string: true
                            scheme:
                              description: Scheme to use for connecting to the host.
                                Defaults to HTTP.
                              type: string
                          required:
                          - port
                          type: object
                        initialDelaySeconds:
                          description: 'Number of seconds after the container has
                            started before liveness probes are initiated. More info:
                            https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                          format: int32
                          type: integer
                        periodSeconds:
                          description: How often (in seconds) to perform the probe.
                            Default to 10 seconds. Minimum value is 1.
                          format: int32
                          type: integer
                        successThreshold:
                          description: Minimum consecutive successes for the probe
                            to be considered successful after having failed. Defaults
                            to 1. Must be 1 for liveness and startup. Minimum value
                            is 1.
                          format: int32
                          type: integer
                        tcpSocket:
                          description: TCPSocket specifies an action involving a TCP
                            port.
                          properties:
                            host:
                              description: 'Optional: Host name to connect to, defaults
                                to the pod IP.'
                              type: string
                            port:
                              anyOf:
                              - type: integer
                              - type: string
                              description: Number or name of the port to access on
                                the container. Number must be in the range 1 to 65535.
                                Name must be an IANA_SVC_NAME.
                              x-kubernetes-int-or-string: true
                          required:
                          - port
                          type: object
                        terminationGracePeriodSeconds:
                          description: Optional duration in seconds the pod needs
                            to terminate gracefully upon probe failure. The grace
                            period is the duration in seconds after the processes
                            running in the pod are sent a termination signal and the
                            time when the processes are forcibly halted with a kill
                            signal. Set this value longer than the expected cleanup
                            time for your process. If this value is nil, the pod's
                            terminationGracePeriodSeconds will be used. Otherwise,
                            this value overrides the value provided by the pod spec.
                            Value must be non-negative integer. The value zero indicates
                            stop immediately via the kill signal (no opportunity to
                            shut down). This is a beta field and requires enabling
                            ProbeTerminationGracePeriod feature gate. Minimum value
                            is 1. spec.terminationGracePeriodSeconds is used if unset.
                          format: int64
                          type: integer
                        timeoutSeconds:
                          description: 'Number of seconds after which the probe times
                            out. Defaults to 1 second. Minimum value is 1. More info:
                            https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                          format: int32
                          type: integer
                      type: object
                    stdin:
                      description: Whether this container should allocate a buffer
                        for stdin in the container runtime. If this is not set, reads
                        from stdin in the container will always result in EOF. Default
                        is false.
                      type: boolean
                    stdinOnce:
                      description: Whether the container runtime should close the
                        stdin channel after it has been opened by a single attach.
                        When stdin is true the stdin stream will remain open across
                        multiple attach sessions. If stdinOnce is set to true, stdin
                        is opened on container start, is empty until the first client
                        attaches to stdin, and then remains open and accepts data
                        until the client disconnects, at which time stdin is closed
                        and remains closed until the container is restarted. If this
                        flag is false, a container processes that reads from stdin
                        will never receive an EOF. Default is false
                      type: boolean
                    terminationMessagePath:
                      description: 'Optional: Path at which the file to which the
                        container''s termination message will be written is mounted
                        into the container''s filesystem. Message written is intended
                        to be brief final status, such as an assertion failure message.
                        Will be truncated by the node if greater than 4096 bytes.
                        The total message length across all containers will be limited
                        to 12kb. Defaults to /dev/termination-log. Cannot be updated.'
                      type: string
                    terminationMessagePolicy:
                      description: Indicate how the termination message should be
                        populated. File will use the contents of terminationMessagePath
                        to populate the container status message on both success and
                        failure. FallbackToLogsOnError will use the last chunk of
                        container log output if the termination message file is empty
                        and the container exited with an error. The log output is
                        limited to 2048 bytes or 80 lines, whichever is smaller. Defaults
                        to File. Cannot be updated.
                      type: string
                    tty:
                      description: Whether this container should allocate a TTY for
                        itself, also requires 'stdin' to be true. Default is false.
                      type: boolean
                    volumeDevices:
                      description: volumeDevices is the list of block devices to be
                        used by the container.
                      items:
                        description: volumeDevice describes a mapping of a raw block
                          device within a container.
                        properties:
                          devicePath:
                            description: devicePath is the path inside of the container
                              that the device will be mapped to.
                            type: string
                          name:
                            description: name must match the name of a persistentVolumeClaim
                              in the pod
                            type: string
                        required:
                        - devicePath
                        - name
                        type: object
                      type: array
                    volumeMounts:
                      description: Pod volumes to mount into the container's filesystem.
                        Cannot be updated.
                      items:
                        description: VolumeMount describes a mounting of a Volume
                          within a container.
                        properties:
                          mountPath:
                            description: Path within the container at which the volume
                              should be mounted.  Must not contain ':'.
                            type: string
                          mountPropagation:
                            description: mountPropagation determines how mounts are
                              propagated from the host to container and the other
                              way around. When not set, MountPropagationNone is used.
                              This field is beta in 1.10.
                            type: string
                          name:
                            description: This must match the Name of a Volume.
                            type: string
                          readOnly:
                            description: Mounted read-only if true, read-write otherwise
                              (false or unspecified). Defaults to false.
                            type: boolean
                          subPath:
                            description: Path within the volume from which the container's
                              volume should be mounted. Defaults to "" (volume's root).
                            type: string
                          subPathExpr:
                            description: Expanded path within the volume from which
                              the container's volume should be mounted. Behaves similarly
                              to SubPath but environment variable references $(VAR_NAME)
                              are expanded using the container's environment. Defaults
                              to "" (volume's root). SubPathExpr and SubPath are mutually
                              exclusive.
                            type: string
                        required:
                        - mountPath
                        - name
                        type: object
                      type: array
                    workingDir:
                      description: Container's working directory. If not specified,
                        the container runtime's default will be used, which might
                        be configured in the container image. Cannot be updated.
                      type: string
                  required:
                  - configs
                  - name
                  - secrets
                  type: object
                type: array
              env:
                type: string
              part-of:
                type: string
              reloader:
                type: boolean
              scaling:
                properties:
                  cpu:
                    properties:
                      target:
                        type: string
                    type: object
                  maxreplica:
                    format: int32
                    type: integer
                  memory:
                    properties:
                      target:
                        type: string
                    type: object
                  minreplica:
                    format: int32
                    type: integer
                  pubsubTopic:
                    properties:
                      name:
                        type: string
                      size:
                        type: string
                    type: object
                required:
                - maxreplica
                - minreplica
                type: object
              strategy:
                properties:
                  metrics:
                    properties:
                      datadog:
                        properties:
                          errorRPM:
                            type: string
                          operation:
                            type: string
                          p95latency:
                            type: string
                        required:
                        - operation
                        type: object
                    required:
                    - datadog
                    type: object
                required:
                - metrics
                type: object
            required:
            - app
            - part-of
            type: object
        required:
        - metadata
        - spec
        type: object
    served: true
    storage: true
`
