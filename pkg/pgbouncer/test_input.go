package pgbouncer

var allPresent = `
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: pgbouncer-app
spec:
  secretStoreRef:
    kind: ClusterSecretStore
    name: vault-backend-app
  target:
    name: postgres-creds
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
        property: data`

var missingDatabase = `
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: pgbouncer-app
spec:
  secretStoreRef:
    kind: ClusterSecretStore
    name: vault-backend-app
  target:
    name: postgres-creds
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
        property: data`
