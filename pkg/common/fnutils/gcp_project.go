package fnutils

const (
	GCP_STAGING_PROJECT = "beecash-staging"
	GCP_PROD_PROJECT    = "tokko-production"
)

func GetProject(env string) string{
	if env == "prod" {
		return GCP_PROD_PROJECT
	} else {
		return GCP_STAGING_PROJECT
	}
}
