package configuration

import "fmt"

const (
	OJTGUARDIAN_PATH    = "/etc/ojtguardian"
	MAIN_CONFIG         = OJTGUARDIAN_PATH + "/main.yml"
	MASTER_SERVER_URL   = "http://10.142.176.1:8000"
	INCIDENT_CREATE_URL = MASTER_SERVER_URL + "/api/v1/incidents/add"
)

func GetIncidentStatusURL(id uint64) string {
	return fmt.Sprintf("%s/api/v1/incidents/%d/update-status", MASTER_SERVER_URL, id)
}
