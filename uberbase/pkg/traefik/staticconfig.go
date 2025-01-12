package traefik

type TraefikStaticConfiguration struct {
	Global struct {
		CheckNewVersion    bool `yaml:"checkNewVersion"`
		SendAnonymousUsage bool `yaml:"sendAnonymousUsage"`
	} `yaml:"global"`
	ServersTransport struct {
		InsecureSkipVerify  bool     `yaml:"insecureSkipVerify"`
		RootCAs             []string `yaml:"rootCAs"`
		MaxIdleConnsPerHost int      `yaml:"maxIdleConnsPerHost"`
		ForwardingTimeouts  struct {
			DialTimeout           string `yaml:"dialTimeout"`
			ResponseHeaderTimeout string `yaml:"responseHeaderTimeout"`
			IdleConnTimeout       string `yaml:"idleConnTimeout"`
		} `yaml:"forwardingTimeouts"`
		Spiffe struct {
			Ids         []string `yaml:"ids"`
			TrustDomain string   `yaml:"trustDomain"`
		} `yaml:"spiffe"`
	} `yaml:"serversTransport"`
	TCPServersTransport struct {
		DialKeepAlive    string `yaml:"dialKeepAlive"`
		DialTimeout      string `yaml:"dialTimeout"`
		TerminationDelay string `yaml:"terminationDelay"`
		TLS              struct {
			InsecureSkipVerify bool     `yaml:"insecureSkipVerify"`
			RootCAs            []string `yaml:"rootCAs"`
			Spiffe             struct {
				Ids         []string `yaml:"ids"`
				TrustDomain string   `yaml:"trustDomain"`
			} `yaml:"spiffe"`
		} `yaml:"tls"`
	} `yaml:"tcpServersTransport"`
	EntryPoints map[string]struct {
		Address         string `yaml:"address"`
		AllowACMEByPass bool   `yaml:"allowACMEByPass"`
		ReusePort       bool   `yaml:"reusePort"`
		AsDefault       bool   `yaml:"asDefault"`
		Transport       struct {
			LifeCycle struct {
				RequestAcceptGraceTimeout string `yaml:"requestAcceptGraceTimeout"`
				GraceTimeOut              string `yaml:"graceTimeOut"`
			} `yaml:"lifeCycle"`
			RespondingTimeouts struct {
				ReadTimeout  string `yaml:"readTimeout"`
				WriteTimeout string `yaml:"writeTimeout"`
				IdleTimeout  string `yaml:"idleTimeout"`
			} `yaml:"respondingTimeouts"`
			KeepAliveMaxTime     string `yaml:"keepAliveMaxTime"`
			KeepAliveMaxRequests int    `yaml:"keepAliveMaxRequests"`
		} `yaml:"transport"`
		ProxyProtocol struct {
			Insecure   bool     `yaml:"insecure"`
			TrustedIPs []string `yaml:"trustedIPs"`
		} `yaml:"proxyProtocol"`
		ForwardedHeaders struct {
			Insecure   bool     `yaml:"insecure"`
			TrustedIPs []string `yaml:"trustedIPs"`
			Connection []string `yaml:"connection"`
		} `yaml:"forwardedHeaders"`
		HTTP struct {
			Redirections struct {
				EntryPoint struct {
					To        string `yaml:"to"`
					Scheme    string `yaml:"scheme"`
					Permanent bool   `yaml:"permanent"`
					Priority  int    `yaml:"priority"`
				} `yaml:"entryPoint"`
			} `yaml:"redirections"`
			Middlewares []string `yaml:"middlewares"`
			TLS         struct {
				Options      string `yaml:"options"`
				CertResolver string `yaml:"certResolver"`
				Domains      []struct {
					Main string   `yaml:"main"`
					Sans []string `yaml:"sans"`
				} `yaml:"domains"`
			} `yaml:"tls"`
			EncodeQuerySemicolons bool `yaml:"encodeQuerySemicolons"`
			MaxHeaderBytes        int  `yaml:"maxHeaderBytes"`
		} `yaml:"http"`
		HTTP2 struct {
			MaxConcurrentStreams int `yaml:"maxConcurrentStreams"`
		} `yaml:"http2"`
		HTTP3 struct {
			AdvertisedPort int `yaml:"advertisedPort"`
		} `yaml:"http3"`
		UDP struct {
			Timeout string `yaml:"timeout"`
		} `yaml:"udp"`
		Observability struct {
			AccessLogs bool `yaml:"accessLogs"`
			Tracing    bool `yaml:"tracing"`
			Metrics    bool `yaml:"metrics"`
		} `yaml:"observability"`
	} `yaml:"entryPoints"`
	Providers struct {
		ProvidersThrottleDuration string `yaml:"providersThrottleDuration"`
		Docker                    struct {
			ExposedByDefault   bool   `yaml:"exposedByDefault"`
			Constraints        string `yaml:"constraints"`
			AllowEmptyServices bool   `yaml:"allowEmptyServices"`
			Network            string `yaml:"network"`
			UseBindPortIP      bool   `yaml:"useBindPortIP"`
			Watch              bool   `yaml:"watch"`
			DefaultRule        string `yaml:"defaultRule"`
			Username           string `yaml:"username"`
			Password           string `yaml:"password"`
			Endpoint           string `yaml:"endpoint"`
			TLS                struct {
				Ca                 string `yaml:"ca"`
				Cert               string `yaml:"cert"`
				Key                string `yaml:"key"`
				InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
			} `yaml:"tls"`
			HTTPClientTimeout string `yaml:"httpClientTimeout"`
		} `yaml:"docker"`
		Swarm struct {
			ExposedByDefault   bool   `yaml:"exposedByDefault"`
			Constraints        string `yaml:"constraints"`
			AllowEmptyServices bool   `yaml:"allowEmptyServices"`
			Network            string `yaml:"network"`
			UseBindPortIP      bool   `yaml:"useBindPortIP"`
			Watch              bool   `yaml:"watch"`
			DefaultRule        string `yaml:"defaultRule"`
			Username           string `yaml:"username"`
			Password           string `yaml:"password"`
			Endpoint           string `yaml:"endpoint"`
			TLS                struct {
				Ca                 string `yaml:"ca"`
				Cert               string `yaml:"cert"`
				Key                string `yaml:"key"`
				InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
			} `yaml:"tls"`
			HTTPClientTimeout string `yaml:"httpClientTimeout"`
			RefreshSeconds    string `yaml:"refreshSeconds"`
		} `yaml:"swarm"`
		File struct {
			Directory                 string `yaml:"directory"`
			Watch                     bool   `yaml:"watch"`
			Filename                  string `yaml:"filename"`
			DebugLogGeneratedTemplate bool   `yaml:"debugLogGeneratedTemplate"`
		} `yaml:"file"`
		KubernetesIngress struct {
			Endpoint         string   `yaml:"endpoint"`
			Token            string   `yaml:"token"`
			CertAuthFilePath string   `yaml:"certAuthFilePath"`
			Namespaces       []string `yaml:"namespaces"`
			LabelSelector    string   `yaml:"labelSelector"`
			IngressClass     string   `yaml:"ingressClass"`
			IngressEndpoint  struct {
				IP               string `yaml:"ip"`
				Hostname         string `yaml:"hostname"`
				PublishedService string `yaml:"publishedService"`
			} `yaml:"ingressEndpoint"`
			ThrottleDuration             string `yaml:"throttleDuration"`
			AllowEmptyServices           bool   `yaml:"allowEmptyServices"`
			AllowExternalNameServices    bool   `yaml:"allowExternalNameServices"`
			DisableIngressClassLookup    bool   `yaml:"disableIngressClassLookup"`
			DisableClusterScopeResources bool   `yaml:"disableClusterScopeResources"`
			NativeLBByDefault            bool   `yaml:"nativeLBByDefault"`
		} `yaml:"kubernetesIngress"`
		KubernetesCRD struct {
			Endpoint                     string   `yaml:"endpoint"`
			Token                        string   `yaml:"token"`
			CertAuthFilePath             string   `yaml:"certAuthFilePath"`
			Namespaces                   []string `yaml:"namespaces"`
			AllowCrossNamespace          bool     `yaml:"allowCrossNamespace"`
			AllowExternalNameServices    bool     `yaml:"allowExternalNameServices"`
			LabelSelector                string   `yaml:"labelSelector"`
			IngressClass                 string   `yaml:"ingressClass"`
			ThrottleDuration             string   `yaml:"throttleDuration"`
			AllowEmptyServices           bool     `yaml:"allowEmptyServices"`
			NativeLBByDefault            bool     `yaml:"nativeLBByDefault"`
			DisableClusterScopeResources bool     `yaml:"disableClusterScopeResources"`
		} `yaml:"kubernetesCRD"`
		KubernetesGateway struct {
			Endpoint            string   `yaml:"endpoint"`
			Token               string   `yaml:"token"`
			CertAuthFilePath    string   `yaml:"certAuthFilePath"`
			Namespaces          []string `yaml:"namespaces"`
			LabelSelector       string   `yaml:"labelSelector"`
			ThrottleDuration    string   `yaml:"throttleDuration"`
			ExperimentalChannel bool     `yaml:"experimentalChannel"`
			StatusAddress       struct {
				IP       string `yaml:"ip"`
				Hostname string `yaml:"hostname"`
				Service  struct {
					Name      string `yaml:"name"`
					Namespace string `yaml:"namespace"`
				} `yaml:"service"`
			} `yaml:"statusAddress"`
			NativeLBByDefault bool `yaml:"nativeLBByDefault"`
		} `yaml:"kubernetesGateway"`
		Rest struct {
			Insecure bool `yaml:"insecure"`
		} `yaml:"rest"`
		ConsulCatalog struct {
			Constraints string `yaml:"constraints"`
			Endpoint    struct {
				Address    string `yaml:"address"`
				Scheme     string `yaml:"scheme"`
				Datacenter string `yaml:"datacenter"`
				Token      string `yaml:"token"`
				TLS        struct {
					Ca                 string `yaml:"ca"`
					Cert               string `yaml:"cert"`
					Key                string `yaml:"key"`
					InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
				} `yaml:"tls"`
				HTTPAuth struct {
					Username string `yaml:"username"`
					Password string `yaml:"password"`
				} `yaml:"httpAuth"`
				EndpointWaitTime string `yaml:"endpointWaitTime"`
			} `yaml:"endpoint"`
			Prefix            string   `yaml:"prefix"`
			RefreshInterval   string   `yaml:"refreshInterval"`
			RequireConsistent bool     `yaml:"requireConsistent"`
			Stale             bool     `yaml:"stale"`
			Cache             bool     `yaml:"cache"`
			ExposedByDefault  bool     `yaml:"exposedByDefault"`
			DefaultRule       string   `yaml:"defaultRule"`
			ConnectAware      bool     `yaml:"connectAware"`
			ConnectByDefault  bool     `yaml:"connectByDefault"`
			ServiceName       string   `yaml:"serviceName"`
			Watch             bool     `yaml:"watch"`
			StrictChecks      []string `yaml:"strictChecks"`
			Namespaces        []string `yaml:"namespaces"`
		} `yaml:"consulCatalog"`
		Nomad struct {
			DefaultRule string `yaml:"defaultRule"`
			Constraints string `yaml:"constraints"`
			Endpoint    struct {
				Address string `yaml:"address"`
				Region  string `yaml:"region"`
				Token   string `yaml:"token"`
				TLS     struct {
					Ca                 string `yaml:"ca"`
					Cert               string `yaml:"cert"`
					Key                string `yaml:"key"`
					InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
				} `yaml:"tls"`
				EndpointWaitTime string `yaml:"endpointWaitTime"`
			} `yaml:"endpoint"`
			Prefix             string   `yaml:"prefix"`
			Stale              bool     `yaml:"stale"`
			ExposedByDefault   bool     `yaml:"exposedByDefault"`
			RefreshInterval    string   `yaml:"refreshInterval"`
			AllowEmptyServices bool     `yaml:"allowEmptyServices"`
			Watch              bool     `yaml:"watch"`
			ThrottleDuration   string   `yaml:"throttleDuration"`
			Namespaces         []string `yaml:"namespaces"`
		} `yaml:"nomad"`
		Ecs struct {
			Constraints          string   `yaml:"constraints"`
			ExposedByDefault     bool     `yaml:"exposedByDefault"`
			RefreshSeconds       int      `yaml:"refreshSeconds"`
			DefaultRule          string   `yaml:"defaultRule"`
			Clusters             []string `yaml:"clusters"`
			AutoDiscoverClusters bool     `yaml:"autoDiscoverClusters"`
			HealthyTasksOnly     bool     `yaml:"healthyTasksOnly"`
			EcsAnywhere          bool     `yaml:"ecsAnywhere"`
			Region               string   `yaml:"region"`
			AccessKeyID          string   `yaml:"accessKeyID"`
			SecretAccessKey      string   `yaml:"secretAccessKey"`
		} `yaml:"ecs"`
		Consul struct {
			RootKey   string   `yaml:"rootKey"`
			Endpoints []string `yaml:"endpoints"`
			Token     string   `yaml:"token"`
			TLS       struct {
				Ca                 string `yaml:"ca"`
				Cert               string `yaml:"cert"`
				Key                string `yaml:"key"`
				InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
			} `yaml:"tls"`
			Namespaces []string `yaml:"namespaces"`
		} `yaml:"consul"`
		Etcd struct {
			RootKey   string   `yaml:"rootKey"`
			Endpoints []string `yaml:"endpoints"`
			TLS       struct {
				Ca                 string `yaml:"ca"`
				Cert               string `yaml:"cert"`
				Key                string `yaml:"key"`
				InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
			} `yaml:"tls"`
			Username string `yaml:"username"`
			Password string `yaml:"password"`
		} `yaml:"etcd"`
		ZooKeeper struct {
			RootKey   string   `yaml:"rootKey"`
			Endpoints []string `yaml:"endpoints"`
			Username  string   `yaml:"username"`
			Password  string   `yaml:"password"`
		} `yaml:"zooKeeper"`
		Redis struct {
			RootKey   string   `yaml:"rootKey"`
			Endpoints []string `yaml:"endpoints"`
			TLS       struct {
				Ca                 string `yaml:"ca"`
				Cert               string `yaml:"cert"`
				Key                string `yaml:"key"`
				InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
			} `yaml:"tls"`
			Username string `yaml:"username"`
			Password string `yaml:"password"`
			Db       int    `yaml:"db"`
			Sentinel struct {
				MasterName              string `yaml:"masterName"`
				Username                string `yaml:"username"`
				Password                string `yaml:"password"`
				LatencyStrategy         bool   `yaml:"latencyStrategy"`
				RandomStrategy          bool   `yaml:"randomStrategy"`
				ReplicaStrategy         bool   `yaml:"replicaStrategy"`
				UseDisconnectedReplicas bool   `yaml:"useDisconnectedReplicas"`
			} `yaml:"sentinel"`
		} `yaml:"redis"`
		HTTP struct {
			Endpoint     string            `yaml:"endpoint"`
			PollInterval string            `yaml:"pollInterval"`
			PollTimeout  string            `yaml:"pollTimeout"`
			Headers      map[string]string `yaml:"headers"`
			TLS          struct {
				Ca                 string `yaml:"ca"`
				Cert               string `yaml:"cert"`
				Key                string `yaml:"key"`
				InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
			} `yaml:"tls"`
		} `yaml:"http"`
		Plugin map[string]map[string]string `yaml:"plugin"`
	} `yaml:"providers"`
	API struct {
		BasePath           string `yaml:"basePath"`
		Insecure           bool   `yaml:"insecure"`
		Dashboard          bool   `yaml:"dashboard"`
		Debug              bool   `yaml:"debug"`
		DisableDashboardAd bool   `yaml:"disableDashboardAd"`
	} `yaml:"api"`
	Metrics struct {
		AddInternals bool `yaml:"addInternals"`
		Prometheus   struct {
			Buckets              []int             `yaml:"buckets"`
			AddEntryPointsLabels bool              `yaml:"addEntryPointsLabels"`
			AddRoutersLabels     bool              `yaml:"addRoutersLabels"`
			AddServicesLabels    bool              `yaml:"addServicesLabels"`
			EntryPoint           string            `yaml:"entryPoint"`
			ManualRouting        bool              `yaml:"manualRouting"`
			HeaderLabels         map[string]string `yaml:"headerLabels"`
		} `yaml:"prometheus"`
		Datadog struct {
			Address              string `yaml:"address"`
			PushInterval         string `yaml:"pushInterval"`
			AddEntryPointsLabels bool   `yaml:"addEntryPointsLabels"`
			AddRoutersLabels     bool   `yaml:"addRoutersLabels"`
			AddServicesLabels    bool   `yaml:"addServicesLabels"`
			Prefix               string `yaml:"prefix"`
		} `yaml:"datadog"`
		StatsD struct {
			Address              string `yaml:"address"`
			PushInterval         string `yaml:"pushInterval"`
			AddEntryPointsLabels bool   `yaml:"addEntryPointsLabels"`
			AddRoutersLabels     bool   `yaml:"addRoutersLabels"`
			AddServicesLabels    bool   `yaml:"addServicesLabels"`
			Prefix               string `yaml:"prefix"`
		} `yaml:"statsD"`
		InfluxDB2 struct {
			Address              string            `yaml:"address"`
			Token                string            `yaml:"token"`
			PushInterval         string            `yaml:"pushInterval"`
			Org                  string            `yaml:"org"`
			Bucket               string            `yaml:"bucket"`
			AddEntryPointsLabels bool              `yaml:"addEntryPointsLabels"`
			AddRoutersLabels     bool              `yaml:"addRoutersLabels"`
			AddServicesLabels    bool              `yaml:"addServicesLabels"`
			AdditionalLabels     map[string]string `yaml:"additionalLabels"`
		} `yaml:"influxDB2"`
		Otlp struct {
			Grpc struct {
				Endpoint string `yaml:"endpoint"`
				Insecure bool   `yaml:"insecure"`
				TLS      struct {
					Ca                 string `yaml:"ca"`
					Cert               string `yaml:"cert"`
					Key                string `yaml:"key"`
					InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
				} `yaml:"tls"`
				Headers map[string]string `yaml:"headers"`
			} `yaml:"grpc"`
			HTTP struct {
				Endpoint string `yaml:"endpoint"`
				TLS      struct {
					Ca                 string `yaml:"ca"`
					Cert               string `yaml:"cert"`
					Key                string `yaml:"key"`
					InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
				} `yaml:"tls"`
				Headers map[string]string `yaml:"headers"`
			} `yaml:"http"`
			AddEntryPointsLabels bool   `yaml:"addEntryPointsLabels"`
			AddRoutersLabels     bool   `yaml:"addRoutersLabels"`
			AddServicesLabels    bool   `yaml:"addServicesLabels"`
			ExplicitBoundaries   []int  `yaml:"explicitBoundaries"`
			PushInterval         string `yaml:"pushInterval"`
			ServiceName          string `yaml:"serviceName"`
		} `yaml:"otlp"`
	} `yaml:"metrics"`
	Ping struct {
		EntryPoint            string `yaml:"entryPoint"`
		ManualRouting         bool   `yaml:"manualRouting"`
		TerminatingStatusCode int    `yaml:"terminatingStatusCode"`
	} `yaml:"ping"`
	Log struct {
		Level      string `yaml:"level"`
		Format     string `yaml:"format"`
		NoColor    bool   `yaml:"noColor"`
		FilePath   string `yaml:"filePath"`
		MaxSize    int    `yaml:"maxSize"`
		MaxAge     int    `yaml:"maxAge"`
		MaxBackups int    `yaml:"maxBackups"`
		Compress   bool   `yaml:"compress"`
		Otlp       struct {
			ServiceName        string            `yaml:"serviceName"`
			ResourceAttributes map[string]string `yaml:"resourceAttributes"`
			Grpc               struct {
				Endpoint string `yaml:"endpoint"`
				Insecure bool   `yaml:"insecure"`
				TLS      struct {
					Ca                 string `yaml:"ca"`
					Cert               string `yaml:"cert"`
					Key                string `yaml:"key"`
					InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
				} `yaml:"tls"`
				Headers map[string]string `yaml:"headers"`
			} `yaml:"grpc"`
			HTTP struct {
				Endpoint string `yaml:"endpoint"`
				TLS      struct {
					Ca                 string `yaml:"ca"`
					Cert               string `yaml:"cert"`
					Key                string `yaml:"key"`
					InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
				} `yaml:"tls"`
				Headers map[string]string `yaml:"headers"`
			} `yaml:"http"`
		} `yaml:"otlp"`
	} `yaml:"log"`
	AccessLog struct {
		FilePath string `yaml:"filePath"`
		Format   string `yaml:"format"`
		Filters  struct {
			StatusCodes   []string `yaml:"statusCodes"`
			RetryAttempts bool     `yaml:"retryAttempts"`
			MinDuration   string   `yaml:"minDuration"`
		} `yaml:"filters"`
		Fields struct {
			DefaultMode string            `yaml:"defaultMode"`
			Names       map[string]string `yaml:"names"`
			Headers     struct {
				DefaultMode string            `yaml:"defaultMode"`
				Names       map[string]string `yaml:"names"`
			} `yaml:"headers"`
		} `yaml:"fields"`
		BufferingSize int  `yaml:"bufferingSize"`
		AddInternals  bool `yaml:"addInternals"`
		Otlp          struct {
			ServiceName        string            `yaml:"serviceName"`
			ResourceAttributes map[string]string `yaml:"resourceAttributes"`
			Grpc               struct {
				Endpoint string `yaml:"endpoint"`
				Insecure bool   `yaml:"insecure"`
				TLS      struct {
					Ca                 string `yaml:"ca"`
					Cert               string `yaml:"cert"`
					Key                string `yaml:"key"`
					InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
				} `yaml:"tls"`
				Headers map[string]string `yaml:"headers"`
			} `yaml:"grpc"`
			HTTP struct {
				Endpoint string `yaml:"endpoint"`
				TLS      struct {
					Ca                 string `yaml:"ca"`
					Cert               string `yaml:"cert"`
					Key                string `yaml:"key"`
					InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
				} `yaml:"tls"`
				Headers map[string]string `yaml:"headers"`
			} `yaml:"http"`
		} `yaml:"otlp"`
	} `yaml:"accessLog"`
	Tracing struct {
		ServiceName             string            `yaml:"serviceName"`
		ResourceAttributes      map[string]string `yaml:"resourceAttributes"`
		CapturedRequestHeaders  []string          `yaml:"capturedRequestHeaders"`
		CapturedResponseHeaders []string          `yaml:"capturedResponseHeaders"`
		SafeQueryParams         []string          `yaml:"safeQueryParams"`
		SampleRate              int               `yaml:"sampleRate"`
		AddInternals            bool              `yaml:"addInternals"`
		Otlp                    struct {
			Grpc struct {
				Endpoint string `yaml:"endpoint"`
				Insecure bool   `yaml:"insecure"`
				TLS      struct {
					Ca                 string `yaml:"ca"`
					Cert               string `yaml:"cert"`
					Key                string `yaml:"key"`
					InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
				} `yaml:"tls"`
				Headers map[string]string `yaml:"headers"`
			} `yaml:"grpc"`
			HTTP struct {
				Endpoint string `yaml:"endpoint"`
				TLS      struct {
					Ca                 string `yaml:"ca"`
					Cert               string `yaml:"cert"`
					Key                string `yaml:"key"`
					InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
				} `yaml:"tls"`
				Headers map[string]string `yaml:"headers"`
			} `yaml:"http"`
		} `yaml:"otlp"`
		GlobalAttributes map[string]string `yaml:"globalAttributes"`
	} `yaml:"tracing"`
	HostResolver struct {
		CnameFlattening bool   `yaml:"cnameFlattening"`
		ResolvConfig    string `yaml:"resolvConfig"`
		ResolvDepth     int    `yaml:"resolvDepth"`
	} `yaml:"hostResolver"`
	CertificatesResolvers map[string]struct {
		Acme struct {
			Email          string `yaml:"email"`
			CaServer       string `yaml:"caServer"`
			PreferredChain string `yaml:"preferredChain"`
			Storage        string `yaml:"storage"`
			KeyType        string `yaml:"keyType"`
			Eab            struct {
				Kid         string `yaml:"kid"`
				HmacEncoded string `yaml:"hmacEncoded"`
			} `yaml:"eab"`
			CertificatesDuration int      `yaml:"certificatesDuration"`
			CaCertificates       []string `yaml:"caCertificates"`
			CaSystemCertPool     bool     `yaml:"caSystemCertPool"`
			CaServerName         string   `yaml:"caServerName"`
			DNSChallenge         struct {
				Provider    string   `yaml:"provider"`
				Resolvers   []string `yaml:"resolvers"`
				Propagation struct {
					DisableChecks     bool   `yaml:"disableChecks"`
					DisableANSChecks  bool   `yaml:"disableANSChecks"`
					RequireAllRNS     bool   `yaml:"requireAllRNS"`
					DelayBeforeChecks string `yaml:"delayBeforeChecks"`
				} `yaml:"propagation"`
				DelayBeforeCheck        string `yaml:"delayBeforeCheck"`
				DisablePropagationCheck bool   `yaml:"disablePropagationCheck"`
			} `yaml:"dnsChallenge"`
			HTTPChallenge struct {
				EntryPoint string `yaml:"entryPoint"`
			} `yaml:"httpChallenge"`
			TLSChallenge struct {
			} `yaml:"tlsChallenge"`
		} `yaml:"acme"`
		Tailscale struct{} `yaml:"tailscale"`
	} `yaml:"certificatesResolvers"`
	Experimental struct {
		Plugins []struct {
			ModuleName string `yaml:"moduleName"`
			Version    string `yaml:"version"`
			Settings   struct {
				Envs   []string `yaml:"envs"`
				Mounts []string `yaml:"mounts"`
			} `yaml:"settings"`
		} `yaml:"plugins"`
		LocalPlugins []struct {
			ModuleName string `yaml:"moduleName"`
			Settings   struct {
				Envs   []string `yaml:"envs"`
				Mounts []string `yaml:"mounts"`
			} `yaml:"settings"`
		} `yaml:"localPlugins"`
		AbortOnPluginFailure bool `yaml:"abortOnPluginFailure"`
		FastProxy            struct {
			Debug bool `yaml:"debug"`
		} `yaml:"fastProxy"`
		Otlplogs          bool `yaml:"otlplogs"`
		KubernetesGateway bool `yaml:"kubernetesGateway"`
	} `yaml:"experimental"`
	Core struct {
		DefaultRuleSyntax string `yaml:"defaultRuleSyntax"`
	} `yaml:"core"`
	Spiffe struct {
		WorkloadAPIAddr string `yaml:"workloadAPIAddr"`
	} `yaml:"spiffe"`
}
