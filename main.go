package main

import "uptime-go/cmd"

func main() {
	cmd.Execute()

	// now := time.Now()
	// id := net.NotifyIncident(&models.Incident{
	// 	Monitor:     models.Monitor{CreatedAt: now, CertificateExpiredDate: &now},
	// 	Description: "The certificate is nearing expiration",
	// 	Type:        models.UnexpectedStatusCode,
	// }, incident.INFO)

	// net.UpdateIncidentStatus(&models.Incident{
	// 	IncidentID: 844055578484547123,
	// }, incident.Resolved)
}
