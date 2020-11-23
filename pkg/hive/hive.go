package hive

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

const (
	tokenEndpoint = "https://beekeeper.hivehome.com/1.0/cognito"
	nodeEndpoint  = "https://api.prod.bgchprod.info/omnia/nodes/"
)

type Config struct {
	token    string
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type Hive struct {
	Config
}

type Nodes struct {
	Nodes []Node `json:"nodes,omitempty"`
}

type Node struct {
	ID            string    `json:"id,omitempty"`
	HREF          string    `json:"href,omitempty"`
	Name          string    `json:"name,omitempty"`
	ParentNodeID  string    `json:"parentNodeID,omitempty"`
	LastSeen      int64     `json:"lastSeen,omitempty"`
	CreatedOn     int64     `json:"createdOn,omitempty"`
	UserID        string    `json:"userID,omitempty"`
	OwnerID       string    `json:"ownerID,omitempty"`
	HomeID        string    `json:"homeID,omitempty"`
	UpgradeStatus string    `json:"upgradeStatus,omitempty"`
	Attributes    Attribute `json:"attributes,omitempty"`
}

type Attribute struct {
	Temperature           Report `json:"temperature,omitempty"`
	ActiveHeatCoolMode    Report `json:"activeHeatCoolMode,omitempty"`
	ScheduleLockDuration  Report `json:"scheduleLockDuration,omitempty"`
	TargetHeatTemperature Report `json:"targetHeatTemperature,omitempty"`
}

type Report struct {
	TargetValue        interface{} `json:"targetValue,omitempty"`
	ReportedValue      interface{} `json:"reportedValue,omitempty"`
	DisplayValue       interface{} `json:"displayValue,omitempty"`
	ReportReceivedTime int64       `json:"reportReceivedTime,omitempty"`
	ReportChangedTime  int64       `json:"reportChangedTime,omitempty"`
}

func New(c Config) *Hive {
	return &Hive{
		Config: c,
	}
}

// GenerateToken generates a token, using the username/password used when calling New
// and stores it in an unexported field in the Hive struct
func (h *Hive) GenerateToken() error {
	client := &http.Client{}

	b, err := json.Marshal(h.Config)
	if err != nil {
		return fmt.Errorf("error marshalling config: %w", err)
	}

	req, err := http.NewRequest("POST", tokenEndpoint+"/login", bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error requesting login token: %w", err)
	}

	defer res.Body.Close()

	var r map[string]string
	json.NewDecoder(res.Body).Decode(&r)

	t, ok := r["token"]
	if !ok {
		return errors.New("no token returned")
	}

	h.Config.token = t

	return nil
}

// GetTempForNode accepts a nodeID and gets the temperature for that node
// If multiple nodes are returned, it works from the zero'th index
func (h *Hive) GetTempForNode(nodeID string) (float64, error) {
	nodeInfo, err := h.getNodeInformation(nodeID)
	if err != nil {
		return 0.0, fmt.Errorf("error getting node information: %w", err)
	}

	if len(nodeInfo.Nodes) == 0 {
		return 0.0, errors.New("no node information returned")
	}

	f, ok := nodeInfo.Nodes[0].Attributes.Temperature.ReportedValue.(float64)
	if !ok {
		return 0.0, fmt.Errorf("could not assert reported (%v) value to float64", nodeInfo.Nodes[0].Attributes.Temperature.ReportedValue)
	}

	return f, nil
}

func (h *Hive) BoostHeating(nodeID string, targetDuration int32, targetTemperature int32) error {
	r := Nodes{
		Nodes: []Node{
			{
				Attributes: Attribute{
					ActiveHeatCoolMode: Report{
						TargetValue: "BOOST",
					},
					ScheduleLockDuration: Report{
						TargetValue: targetDuration,
					},
					TargetHeatTemperature: Report{
						TargetValue: targetTemperature,
					},
				},
			},
		},
	}

	b, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("error marshalling req: %w", err)
	}

	client := &http.Client{}

	req, err := http.NewRequest("PUT", nodeEndpoint+nodeID, bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/vnd.alertme.zoo-6.2+json")
	req.Header.Set("Accept", "application/vnd.alertme.zoo-6.2+json")
	req.Header.Set("X-AlertMe-Client", "swagger")
	req.Header.Set("X-Omnia-Access-Token", h.token)

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}

	defer res.Body.Close()

	return nil
}

// getNodeInformation takes a nodeID and returns the information for that node
func (h *Hive) getNodeInformation(nodeID string) (Nodes, error) {
	client := &http.Client{}

	var nodeInfo Nodes

	req, err := http.NewRequest("GET", nodeEndpoint+nodeID, nil)
	if err != nil {
		return nodeInfo, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/vnd.alertme.zoo-6.2+json")
	req.Header.Set("Accept", "application/vnd.alertme.zoo-6.2+json")
	req.Header.Set("X-AlertMe-Client", "swagger")
	req.Header.Set("X-Omnia-Access-Token", h.token)

	res, err := client.Do(req)
	if err != nil {
		return nodeInfo, fmt.Errorf("error requesting node information: %w", err)
	}

	defer res.Body.Close()

	json.NewDecoder(res.Body).Decode(&nodeInfo)

	return nodeInfo, nil
}
