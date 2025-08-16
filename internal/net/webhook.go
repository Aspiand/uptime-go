package net

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
	"uptime-go/internal/configuration"
	"uptime-go/internal/incident"
	"uptime-go/internal/models"
)

type incidentCreateResponse struct {
	Message string
	Data    struct {
		ID uint64 `json:"incident_id"`
	} `json:"data"`
}

type incidentUpdateStatusResponse struct{}

func NotifyIncident(incident *models.Incident, severity incident.Severity) uint64 {
	// TODO: improve thos function

	if incident.Monitor.CreatedAt.IsZero() {
		// TODO: handle
		return 0
	}

	reader := configuration.NewConfigReader()
	reader.ReadConfig(configuration.MAIN_CONFIG)
	token := reader.GetServerToken()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	payload, err := json.Marshal(map[string]any{
		"server_ip": GetIPAddress(),
		"module":    "fim",
		"severity":  string(severity),
		"message":   incident.Description,
		"event":     "uptime_" + incident.Type,
		"tags":      []string{"uptime", "monitoring"},
		"attributes": map[string]any{
			"expired_date": incident.Monitor.CertificateExpiredDate,
		},
	})

	// fmt.Println(string(payload))

	if err != nil {
		log.Print("failed to send incident")
		log.Printf("incident id: %s", incident.ID)
		log.Printf("incident master id: %d", incident.IncidentID)
	}

	request, err := http.NewRequest(
		"POST", configuration.INCIDNET_CREATE_URL, bytes.NewBuffer(payload),
	)
	if err != nil {
		log.Printf("Error creating request for %s: %v", configuration.INCIDNET_CREATE_URL, err)
		return 0
	}

	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		// TODO: handle
		return 0
	}
	defer response.Body.Close()

	if response.StatusCode != 201 {
		// TODO: handle
	}

	// Decode body
	var result incidentCreateResponse
	body, _ := io.ReadAll(response.Body)

	if err := json.Unmarshal(body, &result); err != nil {
		// TODO: handle
		fmt.Println(err)
		return 0
	}

	return result.Data.ID
}

func UpdateIncidentStatus(incident *models.Incident, status incident.Status) {
	reader := configuration.NewConfigReader()
	reader.ReadConfig(configuration.MAIN_CONFIG)
	token := reader.GetServerToken()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	payload, err := json.Marshal(map[string]any{
		"status": status,
	})

	if err != nil {
		log.Print("[webhook] failed to update incident status!")
		log.Printf("[webhook] incident id: %s", incident.ID)
		log.Printf("[webhook] incident master id: %d", incident.IncidentID)
	}

	request, err := http.NewRequest(
		"POST", configuration.GetIncidentStatusURL(incident.IncidentID),
		bytes.NewBuffer(payload),
	)
	if err != nil {
		log.Printf("[webhook] Error creating request for %s: %v", configuration.INCIDNET_CREATE_URL, err)
		return
	}

	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		// TODO: handle
		return
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		// TODO: handle
	}

	// Decode body
	var result incidentUpdateStatusResponse
	body, _ := io.ReadAll(response.Body)

	if err := json.Unmarshal(body, &result); err != nil {
		// TODO: handle
		fmt.Println(err)
	}

	fmt.Println(result)
}
