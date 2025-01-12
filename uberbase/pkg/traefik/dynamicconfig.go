package traefik

type TraefikServiceFailover struct {
	Service     string   `yaml:"service"`
	Fallback    string   `yaml:"fallback"`
	HealthCheck struct{} `yaml:"healthCheck"`
}

type TraefikServiceLoadBalancerServer struct {
	URL          string `yaml:"url"`
	Weight       int    `yaml:"weight"`
	PreservePath bool   `yaml:"preservePath"`
}

type TraefikServiceLoadBalancerStickyCookie struct {
	Name     string `yaml:"name"`
	Secure   bool   `yaml:"secure"`
	HTTPOnly bool   `yaml:"httpOnly"`
	SameSite string `yaml:"sameSite"`
	MaxAge   int    `yaml:"maxAge"`
	Path     string `yaml:"path"`
}

type TraefikServiceLoadBalancerSticky struct {
	Cookie TraefikServiceLoadBalancerStickyCookie `yaml:"cookie"`
}

type TraefikServiceLoadBalancerHealthCheck struct {
	Scheme          string            `yaml:"scheme"`
	Mode            string            `yaml:"mode"`
	Path            string            `yaml:"path"`
	Method          string            `yaml:"method"`
	Status          int               `yaml:"status"`
	Port            int               `yaml:"port"`
	Interval        string            `yaml:"interval"`
	Timeout         string            `yaml:"timeout"`
	Hostname        string            `yaml:"hostname"`
	FollowRedirects bool              `yaml:"followRedirects"`
	Headers         map[string]string `yaml:"headers"`
}

type TraefikServiceLoadBalancerResponseForwarding struct {
	FlushInterval string `yaml:"flushInterval"`
}

type TraefikServiceLoadBalancer struct {
	Sticky             TraefikServiceLoadBalancerSticky             `yaml:"sticky"`
	Servers            []TraefikServiceLoadBalancerServer           `yaml:"servers"`
	HealthCheck        TraefikServiceLoadBalancerHealthCheck        `yaml:"healthCheck"`
	PassHostHeader     bool                                         `yaml:"passHostHeader"`
	ResponseForwarding TraefikServiceLoadBalancerResponseForwarding `yaml:"responseForwarding"`
	ServersTransport   string                                       `yaml:"serversTransport"`
}

type TraefikServiceMirror struct {
	Name    string `yaml:"name"`
	Percent int    `yaml:"percent"`
}

type TraefikServiceMirroring struct {
	Service     string                 `yaml:"service"`
	MirrorBody  bool                   `yaml:"mirrorBody"`
	MaxBodySize int                    `yaml:"maxBodySize"`
	Mirrors     []TraefikServiceMirror `yaml:"mirrors"`
	HealthCheck struct{}               `yaml:"healthCheck"`
}

type TraefikServiceWeightedService struct {
	Name   string `yaml:"name"`
	Weight int    `yaml:"weight"`
}

type TraefikServiceWeightedStickyCookie struct {
	Name     string `yaml:"name"`
	Secure   bool   `yaml:"secure"`
	HTTPOnly bool   `yaml:"httpOnly"`
	SameSite string `yaml:"sameSite"`
	MaxAge   int    `yaml:"maxAge"`
	Path     string `yaml:"path"`
}

type TraefikServiceWeightedSticky struct {
	Cookie TraefikServiceWeightedStickyCookie `yaml:"cookie"`
}

type TraefikServiceWeighted struct {
	Services    []TraefikServiceWeightedService `yaml:"services"`
	Sticky      TraefikServiceWeightedSticky    `yaml:"sticky"`
	HealthCheck struct{}                        `yaml:"healthCheck"`
}

type TraefikService struct {
	Failover     *TraefikServiceFailover     `yaml:"failover,omitempty"`
	LoadBalancer *TraefikServiceLoadBalancer `yaml:"loadBalancer,omitempty"`
	Mirroring    *TraefikServiceMirroring    `yaml:"mirroring,omitempty"`
	Weighted     *TraefikServiceWeighted     `yaml:"weighted,omitempty"`
}

type TraefikRouterTLSDomain struct {
	Main string   `yaml:"main"`
	Sans []string `yaml:"sans"`
}

type TraefikRouterTLS struct {
	Options      string                   `yaml:"options"`
	CertResolver string                   `yaml:"certResolver"`
	Domains      []TraefikRouterTLSDomain `yaml:"domains"`
}

type TraefikRouterObservability struct {
	AccessLogs bool `yaml:"accessLogs"`
	Tracing    bool `yaml:"tracing"`
	Metrics    bool `yaml:"metrics"`
}

type TraefikRouter struct {
	EntryPoints   []string                   `yaml:"entryPoints"`
	Middlewares   []string                   `yaml:"middlewares"`
	Service       string                     `yaml:"service"`
	Rule          string                     `yaml:"rule"`
	RuleSyntax    string                     `yaml:"ruleSyntax"`
	Priority      int                        `yaml:"priority"`
	TLS           TraefikRouterTLS           `yaml:"tls"`
	Observability TraefikRouterObservability `yaml:"observability"`
}

type TraefikHTTPConfiguration struct {
	Routers     map[string]TraefikRouter  `yaml:"routers"`
	Services    map[string]TraefikService `yaml:"services"`
	Middlewares map[string]struct {
		AddPrefix *struct {
			Prefix string `yaml:"prefix"`
		} `yaml:"addPrefix,omitempty"`
		BasicAuth *struct {
			Users        []string `yaml:"users"`
			UsersFile    string   `yaml:"usersFile"`
			Realm        string   `yaml:"realm"`
			RemoveHeader bool     `yaml:"removeHeader"`
			HeaderField  string   `yaml:"headerField"`
		} `yaml:"basicAuth,omitempty"`
		Buffering *struct {
			MaxRequestBodyBytes  int    `yaml:"maxRequestBodyBytes"`
			MemRequestBodyBytes  int    `yaml:"memRequestBodyBytes"`
			MaxResponseBodyBytes int    `yaml:"maxResponseBodyBytes"`
			MemResponseBodyBytes int    `yaml:"memResponseBodyBytes"`
			RetryExpression      string `yaml:"retryExpression"`
		} `yaml:"buffering,omitempty"`
		Chain *struct {
			Middlewares []string `yaml:"middlewares"`
		} `yaml:"chain,omitempty"`
		CircuitBreaker *struct {
			Expression       string `yaml:"expression"`
			CheckPeriod      string `yaml:"checkPeriod"`
			FallbackDuration string `yaml:"fallbackDuration"`
			RecoveryDuration string `yaml:"recoveryDuration"`
			ResponseCode     int    `yaml:"responseCode"`
		} `yaml:"circuitBreaker,omitempty"`
		Compress *struct {
			ExcludedContentTypes []string `yaml:"excludedContentTypes"`
			IncludedContentTypes []string `yaml:"includedContentTypes"`
			MinResponseBodyBytes int      `yaml:"minResponseBodyBytes"`
			Encodings            []string `yaml:"encodings"`
			DefaultEncoding      string   `yaml:"defaultEncoding"`
		} `yaml:"compress,omitempty"`
		ContentType *struct {
			AutoDetect bool `yaml:"autoDetect"`
		} `yaml:"contentType,omitempty"`
		DigestAuth *struct {
			Users        []string `yaml:"users"`
			UsersFile    string   `yaml:"usersFile"`
			RemoveHeader bool     `yaml:"removeHeader"`
			Realm        string   `yaml:"realm"`
			HeaderField  string   `yaml:"headerField"`
		} `yaml:"digestAuth,omitempty"`
		Errors *struct {
			Status  []string `yaml:"status"`
			Service string   `yaml:"service"`
			Query   string   `yaml:"query"`
		} `yaml:"errors,omitempty"`
		ForwardAuth *struct {
			Address string `yaml:"address"`
			TLS     struct {
				Ca                 string `yaml:"ca"`
				Cert               string `yaml:"cert"`
				Key                string `yaml:"key"`
				InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
				CaOptional         bool   `yaml:"caOptional"`
			} `yaml:"tls"`
			TrustForwardHeader       bool     `yaml:"trustForwardHeader"`
			AuthResponseHeaders      []string `yaml:"authResponseHeaders"`
			AuthResponseHeadersRegex string   `yaml:"authResponseHeadersRegex"`
			AuthRequestHeaders       []string `yaml:"authRequestHeaders"`
			AddAuthCookiesToResponse []string `yaml:"addAuthCookiesToResponse"`
			HeaderField              string   `yaml:"headerField"`
			ForwardBody              bool     `yaml:"forwardBody"`
			MaxBodySize              int      `yaml:"maxBodySize"`
			PreserveLocationHeader   bool     `yaml:"preserveLocationHeader"`
		} `yaml:"forwardAuth,omitempty"`
		GrpcWeb *struct {
			AllowOrigins []string `yaml:"allowOrigins"`
		} `yaml:"grpcWeb,omitempty"`
		Headers *struct {
			CustomRequestHeaders              map[string]string `yaml:"customRequestHeaders"`
			CustomResponseHeaders             map[string]string `yaml:"customResponseHeaders"`
			AccessControlAllowCredentials     bool              `yaml:"accessControlAllowCredentials"`
			AccessControlAllowHeaders         []string          `yaml:"accessControlAllowHeaders"`
			AccessControlAllowMethods         []string          `yaml:"accessControlAllowMethods"`
			AccessControlAllowOriginList      []string          `yaml:"accessControlAllowOriginList"`
			AccessControlAllowOriginListRegex []string          `yaml:"accessControlAllowOriginListRegex"`
			AccessControlExposeHeaders        []string          `yaml:"accessControlExposeHeaders"`
			AccessControlMaxAge               int               `yaml:"accessControlMaxAge"`
			AddVaryHeader                     bool              `yaml:"addVaryHeader"`
			AllowedHosts                      []string          `yaml:"allowedHosts"`
			HostsProxyHeaders                 []string          `yaml:"hostsProxyHeaders"`
			SslProxyHeaders                   map[string]string `yaml:"sslProxyHeaders"`
			StsSeconds                        int               `yaml:"stsSeconds"`
			StsIncludeSubdomains              bool              `yaml:"stsIncludeSubdomains"`
			StsPreload                        bool              `yaml:"stsPreload"`
			ForceSTSHeader                    bool              `yaml:"forceSTSHeader"`
			FrameDeny                         bool              `yaml:"frameDeny"`
			CustomFrameOptionsValue           string            `yaml:"customFrameOptionsValue"`
			ContentTypeNosniff                bool              `yaml:"contentTypeNosniff"`
			BrowserXSSFilter                  bool              `yaml:"browserXssFilter"`
			CustomBrowserXSSValue             string            `yaml:"customBrowserXSSValue"`
			ContentSecurityPolicy             string            `yaml:"contentSecurityPolicy"`
			ContentSecurityPolicyReportOnly   string            `yaml:"contentSecurityPolicyReportOnly"`
			PublicKey                         string            `yaml:"publicKey"`
			ReferrerPolicy                    string            `yaml:"referrerPolicy"`
			PermissionsPolicy                 string            `yaml:"permissionsPolicy"`
			IsDevelopment                     bool              `yaml:"isDevelopment"`
			FeaturePolicy                     string            `yaml:"featurePolicy"`
			SslRedirect                       bool              `yaml:"sslRedirect"`
			SslTemporaryRedirect              bool              `yaml:"sslTemporaryRedirect"`
			SslHost                           string            `yaml:"sslHost"`
			SslForceHost                      bool              `yaml:"sslForceHost"`
		} `yaml:"headers,omitempty"`
		IPAllowList *struct {
			SourceRange []string `yaml:"sourceRange"`
			IPStrategy  struct {
				Depth       int      `yaml:"depth"`
				ExcludedIPs []string `yaml:"excludedIPs"`
				Ipv6Subnet  int      `yaml:"ipv6Subnet"`
			} `yaml:"ipStrategy"`
			RejectStatusCode int `yaml:"rejectStatusCode"`
		} `yaml:"ipAllowList,omitempty"`
		IPWhiteList *struct {
			SourceRange []string `yaml:"sourceRange"`
			IPStrategy  struct {
				Depth       int      `yaml:"depth"`
				ExcludedIPs []string `yaml:"excludedIPs"`
				Ipv6Subnet  int      `yaml:"ipv6Subnet"`
			} `yaml:"ipStrategy"`
		} `yaml:"ipWhiteList,omitempty"`
		InFlightReq *struct {
			Amount          int `yaml:"amount"`
			SourceCriterion struct {
				IPStrategy struct {
					Depth       int      `yaml:"depth"`
					ExcludedIPs []string `yaml:"excludedIPs"`
					Ipv6Subnet  int      `yaml:"ipv6Subnet"`
				} `yaml:"ipStrategy"`
				RequestHeaderName string `yaml:"requestHeaderName"`
				RequestHost       bool   `yaml:"requestHost"`
			} `yaml:"sourceCriterion"`
		} `yaml:"inFlightReq,omitempty"`
		PassTLSClientCert *struct {
			Pem  bool `yaml:"pem"`
			Info struct {
				NotAfter     bool `yaml:"notAfter"`
				NotBefore    bool `yaml:"notBefore"`
				Sans         bool `yaml:"sans"`
				SerialNumber bool `yaml:"serialNumber"`
				Subject      struct {
					Country            bool `yaml:"country"`
					Province           bool `yaml:"province"`
					Locality           bool `yaml:"locality"`
					Organization       bool `yaml:"organization"`
					OrganizationalUnit bool `yaml:"organizationalUnit"`
					CommonName         bool `yaml:"commonName"`
					SerialNumber       bool `yaml:"serialNumber"`
					DomainComponent    bool `yaml:"domainComponent"`
				} `yaml:"subject"`
				Issuer struct {
					Country         bool `yaml:"country"`
					Province        bool `yaml:"province"`
					Locality        bool `yaml:"locality"`
					Organization    bool `yaml:"organization"`
					CommonName      bool `yaml:"commonName"`
					SerialNumber    bool `yaml:"serialNumber"`
					DomainComponent bool `yaml:"domainComponent"`
				} `yaml:"issuer"`
			} `yaml:"info"`
		} `yaml:"passTLSClientCert,omitempty"`
		Plugin    map[string]map[string]string `yaml:"plugin,omitempty"`
		RateLimit *struct {
			Average         int    `yaml:"average"`
			Period          string `yaml:"period"`
			Burst           int    `yaml:"burst"`
			SourceCriterion struct {
				IPStrategy struct {
					Depth       int      `yaml:"depth"`
					ExcludedIPs []string `yaml:"excludedIPs"`
					Ipv6Subnet  int      `yaml:"ipv6Subnet"`
				} `yaml:"ipStrategy"`
				RequestHeaderName string `yaml:"requestHeaderName"`
				RequestHost       bool   `yaml:"requestHost"`
			} `yaml:"sourceCriterion"`
		} `yaml:"rateLimit,omitempty"`
		RedirectRegex *struct {
			Regex       string `yaml:"regex"`
			Replacement string `yaml:"replacement"`
			Permanent   bool   `yaml:"permanent"`
		} `yaml:"redirectRegex,omitempty"`
		RedirectScheme *struct {
			Scheme    string `yaml:"scheme"`
			Port      string `yaml:"port"`
			Permanent bool   `yaml:"permanent"`
		} `yaml:"redirectScheme,omitempty"`
		ReplacePath *struct {
			Path string `yaml:"path"`
		} `yaml:"replacePath,omitempty"`
		ReplacePathRegex *struct {
			Regex       string `yaml:"regex"`
			Replacement string `yaml:"replacement"`
		} `yaml:"replacePathRegex,omitempty"`
		Retry *struct {
			Attempts        int    `yaml:"attempts"`
			InitialInterval string `yaml:"initialInterval"`
		} `yaml:"retry,omitempty"`
		StripPrefix *struct {
			Prefixes   []string `yaml:"prefixes"`
			ForceSlash bool     `yaml:"forceSlash"`
		} `yaml:"stripPrefix,omitempty"`
		StripPrefixRegex *struct {
			Regex []string `yaml:"regex"`
		} `yaml:"stripPrefixRegex,omitempty"`
	} `yaml:"middlewares"`

	ServersTransports map[string]struct {
		ServerName         string   `yaml:"serverName"`
		InsecureSkipVerify bool     `yaml:"insecureSkipVerify"`
		RootCAs            []string `yaml:"rootCAs"`
		Certificates       []struct {
			CertFile string `yaml:"certFile"`
			KeyFile  string `yaml:"keyFile"`
		} `yaml:"certificates"`
		MaxIdleConnsPerHost int `yaml:"maxIdleConnsPerHost"`
		ForwardingTimeouts  struct {
			DialTimeout           string `yaml:"dialTimeout"`
			ResponseHeaderTimeout string `yaml:"responseHeaderTimeout"`
			IdleConnTimeout       string `yaml:"idleConnTimeout"`
			ReadIdleTimeout       string `yaml:"readIdleTimeout"`
			PingTimeout           string `yaml:"pingTimeout"`
		} `yaml:"forwardingTimeouts"`
		DisableHTTP2 bool   `yaml:"disableHTTP2"`
		PeerCertURI  string `yaml:"peerCertURI"`
		Spiffe       struct {
			Ids         []string `yaml:"ids"`
			TrustDomain string   `yaml:"trustDomain"`
		} `yaml:"spiffe"`
	} `yaml:"serversTransports"`
}

type TraefikDynamicConfiguration struct {
	HTTP *TraefikHTTPConfiguration `yaml:"http"`
	TCP  struct {
		Routers map[string]struct {
			EntryPoints []string `yaml:"entryPoints"`
			Middlewares []string `yaml:"middlewares"`
			Service     string   `yaml:"service"`
			Rule        string   `yaml:"rule"`
			RuleSyntax  string   `yaml:"ruleSyntax"`
			Priority    int      `yaml:"priority"`
			TLS         struct {
				Passthrough  bool   `yaml:"passthrough"`
				Options      string `yaml:"options"`
				CertResolver string `yaml:"certResolver"`
				Domains      []struct {
					Main string   `yaml:"main"`
					Sans []string `yaml:"sans"`
				} `yaml:"domains"`
			} `yaml:"tls"`
		} `yaml:"routers"`

		Services map[string]struct {
			LoadBalancer *struct {
				ProxyProtocol struct {
					Version int `yaml:"version"`
				} `yaml:"proxyProtocol"`
				Servers []struct {
					Address string `yaml:"address"`
					TLS     bool   `yaml:"tls"`
				} `yaml:"servers"`
				ServersTransport string `yaml:"serversTransport"`
				TerminationDelay int    `yaml:"terminationDelay"`
			} `yaml:"loadBalancer,omitempty"`
			Weighted *struct {
				Services []struct {
					Name   string `yaml:"name"`
					Weight int    `yaml:"weight"`
				} `yaml:"services"`
			} `yaml:"weighted,omitempty"`
		} `yaml:"services"`

		Middlewares map[string]struct {
			IPAllowList *struct {
				SourceRange []string `yaml:"sourceRange"`
			} `yaml:"ipAllowList,omitempty"`
			IPWhiteList *struct {
				SourceRange []string `yaml:"sourceRange"`
			} `yaml:"ipWhiteList,omitempty"`
			InFlightConn *struct {
				Amount int `yaml:"amount"`
			} `yaml:"inFlightConn,omitempty"`
		} `yaml:"middlewares"`

		ServersTransports map[string]struct {
			DialKeepAlive    string `yaml:"dialKeepAlive"`
			DialTimeout      string `yaml:"dialTimeout"`
			TerminationDelay string `yaml:"terminationDelay"`
			TLS              struct {
				ServerName         string   `yaml:"serverName"`
				InsecureSkipVerify bool     `yaml:"insecureSkipVerify"`
				RootCAs            []string `yaml:"rootCAs"`
				Certificates       []struct {
					CertFile string `yaml:"certFile"`
					KeyFile  string `yaml:"keyFile"`
				} `yaml:"certificates"`
				PeerCertURI string `yaml:"peerCertURI"`
				Spiffe      struct {
					Ids         []string `yaml:"ids"`
					TrustDomain string   `yaml:"trustDomain"`
				} `yaml:"spiffe"`
			} `yaml:"tls"`
		} `yaml:"serversTransports"`
	} `yaml:"tcp"`

	UDP struct {
		Routers map[string]struct {
			EntryPoints []string `yaml:"entryPoints"`
			Service     string   `yaml:"service"`
		} `yaml:"routers"`

		Services map[string]struct {
			LoadBalancer *struct {
				Servers []struct {
					Address string `yaml:"address"`
				} `yaml:"servers"`
			} `yaml:"loadBalancer,omitempty"`
			Weighted *struct {
				Services []struct {
					Name   string `yaml:"name"`
					Weight int    `yaml:"weight"`
				} `yaml:"services"`
			} `yaml:"weighted,omitempty"`
		} `yaml:"services"`
	} `yaml:"udp"`

	TLS struct {
		Certificates []struct {
			CertFile string   `yaml:"certFile"`
			KeyFile  string   `yaml:"keyFile"`
			Stores   []string `yaml:"stores"`
		} `yaml:"certificates"`

		Options map[string]struct {
			MinVersion       string   `yaml:"minVersion"`
			MaxVersion       string   `yaml:"maxVersion"`
			CipherSuites     []string `yaml:"cipherSuites"`
			CurvePreferences []string `yaml:"curvePreferences"`
			ClientAuth       struct {
				CaFiles        []string `yaml:"caFiles"`
				ClientAuthType string   `yaml:"clientAuthType"`
			} `yaml:"clientAuth"`
			SniStrict                bool     `yaml:"sniStrict"`
			AlpnProtocols            []string `yaml:"alpnProtocols"`
			PreferServerCipherSuites bool     `yaml:"preferServerCipherSuites"`
		} `yaml:"options"`

		Stores map[string]struct {
			DefaultCertificate struct {
				CertFile string `yaml:"certFile"`
				KeyFile  string `yaml:"keyFile"`
			} `yaml:"defaultCertificate"`
			DefaultGeneratedCert struct {
				Resolver string `yaml:"resolver"`
				Domain   struct {
					Main string   `yaml:"main"`
					Sans []string `yaml:"sans"`
				} `yaml:"domain"`
			} `yaml:"defaultGeneratedCert"`
		} `yaml:"stores"`
	} `yaml:"tls"`
}

func (t *TraefikDynamicConfiguration) Copy() *TraefikDynamicConfiguration {
	copy := *t
	return &copy
}
