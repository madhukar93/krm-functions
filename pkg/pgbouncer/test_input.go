package pgbouncer

var allPresent = `
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: unified-customer-profile
spec:
  refreshInterval: 30s
  secretStoreRef:
    kind: ClusterSecretStore
    name: vault-backend-app
  target:
    name: tokko-api-postgres-creds
  data:
    - secretKey: POSTGRESQL_HOST
      remoteRef:
        key: staging/ucp/postgres-host
        property: data
    - secretKey: POSTGRESQL_PORT
      remoteRef:
        key: staging/ucp/postgres-port
        property: data
    - secretKey: POSTGRESQL_USERNAME
      remoteRef:
        key: staging/ucp/postgres-user
        property: data
    - secretKey: POSTGRESQL_PASSWORD
      remoteRef:
        key: staging/ucp/postgres-password
        property: data
    - secretKey: POSTGRESQL_DATABASE
      remoteRef:
        key: staging/ucp/postgres-db
        property: data
    - secretKey: PGBOUNCER_USER
      remoteRef:
        key: staging/ucp/postgres-user
        property: data
    - secretKey: PGBOUNCER_PASS
      remoteRef:
        key: staging/ucp/postgres-password
        property: data
    - secretKey: PGBOUNCER_DATABASE
      remoteRef:
        key: staging/ucp/postgres-db
        property: data
    - secretKey: SENTRY_DSN
      remoteRef:
        key: staging/ucp/sentry-dsn
        property: data`

var missingFields = `
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: unified-customer-profile
spec:
  refreshInterval: 30s
  secretStoreRef:
    kind: ClusterSecretStore
    name: vault-backend-app
  target:
    name: tokko-api-postgres-creds
  data:
    - secretKey: POSTGRESQL_HOST
      remoteRef:
        key: staging/ucp/postgres-host
        property: data
    - secretKey: POSTGRESQL_PORT
      remoteRef:
        key: staging/ucp/postgres-port
        property: data
    - secretKey: POSTGRESQL_USERNAME
      remoteRef:
        key: staging/ucp/postgres-user
        property: data
    - secretKey: PGBOUNCER_USER
      remoteRef:
        key: staging/ucp/postgres-user
        property: data
    - secretKey: PGBOUNCER_PASS
      remoteRef:
        key: staging/ucp/postgres-password
        property: data
    - secretKey: PGBOUNCER_DATABASE
      remoteRef:
        key: staging/ucp/postgres-db
        property: data
    - secretKey: SENTRY_DSN
      remoteRef:
        key: staging/ucp/sentry-dsn
        property: data`

var missingSecretKey = `
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: unified-customer-profile
spec:
  refreshInterval: 30s
  secretStoreRef:
    kind: ClusterSecretStore
    name: vault-backend-app
  target:
    name: tokko-api-postgres-creds
  data:
    - secretKey: POSTGRESQL_HOST
      remoteRef:
        key: staging/ucp/postgres-host
        property: data
    - secretKey: POSTGRESQL_PORT
      remoteRef:
        key: staging/ucp/postgres-port
        property:
    - secretKey: POSTGRESQL_USERNAME
      remoteRef:
        key: staging/ucp/postgres-user
        property: data
    - secretKey: POSTGRESQL_PASSWORD
      remoteRef:
        key: staging/ucp/postgres-password
        property: data
    - secretKey:
      remoteRef:
        key: staging/ucp/postgres-db
        property: data
    - secretKey: PGBOUNCER_USER
      remoteRef:
        key: staging/ucp/postgres-user
        property: data
    - secretKey: PGBOUNCER_PASS
      remoteRef:
        key: staging/ucp/postgres-password
        property: data
    - secretKey: PGBOUNCER_DATABASE
      remoteRef:
        key: staging/ucp/postgres-db
        property: data
    - secretKey: SENTRY_DSN
      remoteRef:
        key: staging/ucp/sentry-dsn
        property: data`
