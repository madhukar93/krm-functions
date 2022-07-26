package injectroutes

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

type test struct {
	name        string
	input       string
	expected    string
	resultCount int
	errorMsg    string
}

func TestInjectRoutes(t *testing.T) {
	var tests = []test{
		{
			name:        "route injection",
			resultCount: 1,
			input: `
apiVersion: config.kubernetes.io/v1
kind: ResourceList
items:
- apiVersion: traefik.containo.us/v1alpha1
  kind: IngressRoute
  metadata:
    name: echo-server-insecure
  spec:
    entryPoints:
      - web
    routes:
    - match: Host('test.com')
      kind: Rule
      services:
        - name: echo-server
          port: 80
    - match: Host('test-secure.com') && Path('/vpn')
      kind: Rule
      middlewares:
        - name: vpn-only
          namespace: traefik
      services:
        - name: echo-server
          port: 80
    - match: Host('domain1.test.com') || Host('domain2.test.com') && Path('/somepath')
      kind: Rule
      middlewares:
        - name: middleware1
          namespace: default
      priority: 10
      services:
        - name: foo
          namespace: default
          passHostHeader: true
          port: 80
          responseForwarding:
            flushInterval: 1ms
          scheme: https
          serversTransport: transport
          sticky:
            cookie:
            httpOnly: true
            name: cookie
            sameSite: none
            secure: true
          strategy: RoundRobin
          weight: 10
    - match: Host('domain1.test.com') || Host('domain2.test.com') && Path('/somepath2')
      kind: Rule
      middlewares:
        - name: middleware1
          namespace: default
      priority: 10
      services:
        - name: foo
          namespace: default
          passHostHeader: true
          port: 80
          responseForwarding:
            flushInterval: 1ms
          scheme: https
          serversTransport: transport
          sticky:
            cookie:
            httpOnly: true
            name: cookie
            sameSite: none
            secure: true
          strategy: RoundRobin
          weight: 10
    tls:
      secretName: echo-server-insecure-cert

functionConfig:
  apiVersion: v1
  kind: SetRoutes
  metadata:
    name: setroutes-fn-config
  data:
    app: testapp
    hosts:
    - domain1.test.com
    - domain2.test.com
    routes:
      - match: Path('/test1')
        kind: Rule`,
		},
	}
	runTests(t, tests)
}

func runTests(t *testing.T, tests []test) {
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			baseDir, err := ioutil.TempDir("", "")
			if !assert.NoError(t, err, test.name) {
				t.FailNow()
			}
			defer os.RemoveAll(baseDir)

			r, err := ioutil.TempFile(baseDir, "fn-*.yaml")
			if !assert.NoError(t, err, test.name) {
				t.FailNow()
			}
			defer os.Remove(r.Name())
			err = ioutil.WriteFile(r.Name(), []byte(test.input), 0600)
			if !assert.NoError(t, err, test.name) {
				t.FailNow()
			}

			injector := &InjectRoutes{}
			inout := &kio.LocalPackageReadWriter{
				PackagePath: baseDir,
			}
			err = kio.Pipeline{
				Inputs:  []kio.Reader{inout},
				Filters: []kio.Filter{injector},
				Outputs: []kio.Writer{inout},
			}.Execute()

			if test.errorMsg != "" {
				if !assert.NotNil(t, err, test.name) {
					t.FailNow()
				}
				if !assert.Contains(t, err.Error(), test.errorMsg) {
					t.FailNow()
				}
			}

			if test.errorMsg == "" && !assert.NoError(t, err, test.name) {
				t.FailNow()
			}

			// get results
			results, err := injector.Results()
			if !assert.NoError(t, err, test.name, test.name) {
				t.FailNow()
			}
			if !assert.Equal(t, test.resultCount, len(results), test.name) {
				t.FailNow()
			}

		})
	}
}
