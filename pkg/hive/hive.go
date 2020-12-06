package hive

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/openlyinc/pointy"

	"github.com/simondrake/home-stats/pkg/cognitosrp"
)

const (
	authEndpoint = "https://cognito-idp.eu-west-1.amazonaws.com"
	nodeEndpoint = "https://api.prod.bgchprod.info/omnia/nodes/"
)

type Config struct {
	token                    string
	Username                 string `json:"username,omitempty"`
	Password                 string `json:"password,omitempty"`
	SSOPoolID                string `json:"ssoPoolID,omitempty"`
	SSOPublicCognitoClientID string `json:"ssoPublicCognitoClientID,omitempty"`
}

type Hive struct {
	httpClient httpClient
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

// New takes a Config object and an optional httpClient
// and returns a pointer to a Hive object
func New(c Config, client httpClient) *Hive {
	if client == nil {
		client = &http.Client{}
	}

	return &Hive{
		httpClient: client,
		Config:     c,
	}
}

// httpClient implements the Do method, which is the exact
// API of the http.Client's DO function. This helps with testing.
type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// GenerateToken generates a token, using the username/password used when calling New
// and stores it in an unexported field in the Hive struct
func (h *Hive) GenerateToken() error {
	csrp, err := cognitosrp.NewCognitoSRP(h.Username, h.Password, h.SSOPoolID, h.SSOPublicCognitoClientID, nil)
	if err != nil {
		return fmt.Errorf("error getting new cognito srp: %w", err)
	}

	awsSession := session.Must(session.NewSession())

	svc := cognitoidentityprovider.New(awsSession, aws.NewConfig().WithEndpoint(authEndpoint).WithRegion("eu-west-1"))

	// initiate auth
	rsp, err := svc.InitiateAuth(&cognitoidentityprovider.InitiateAuthInput{
		AuthFlow:       pointy.String("USER_SRP_AUTH"),
		ClientId:       aws.String(csrp.GetClientId()),
		AuthParameters: csrp.GetAuthParams(),
	})
	if err != nil {
		return fmt.Errorf("error initiating auth: %w", err)
	}

	if rsp.ChallengeName == nil {
		return errors.New("empty challenge name")
	}

	if *rsp.ChallengeName != "PASSWORD_VERIFIER" {
		return errors.New("unhandled challenge returned")
	}

	challengeResponses, _ := csrp.PasswordVerifierChallenge(rsp.ChallengeParameters, time.Now())

	authResponse, err := svc.RespondToAuthChallenge(&cognitoidentityprovider.RespondToAuthChallengeInput{
		ChallengeName:      pointy.String("PASSWORD_VERIFIER"),
		ChallengeResponses: challengeResponses,
		ClientId:           aws.String(csrp.GetClientId()),
	})
	if err != nil {
		return fmt.Errorf("error responding to auth challenge: %w", err)
	}

	if authResponse.AuthenticationResult.IdToken == nil {
		return errors.New("empty id token")
	}

	h.token = *authResponse.AuthenticationResult.IdToken

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

	req, err := http.NewRequest("PUT", nodeEndpoint+nodeID, bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/vnd.alertme.zoo-6.2+json")
	req.Header.Set("Accept", "application/vnd.alertme.zoo-6.2+json")
	req.Header.Set("X-Omnia-Client", "ESP")
	req.Header.Set("Authorization", "Bearer "+h.token)

	res, err := h.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}

	defer res.Body.Close()

	return nil
}

// getNodeInformation takes a nodeID and returns the information for that node
func (h *Hive) getNodeInformation(nodeID string) (Nodes, error) {
	var nodeInfo Nodes

	endpoint := fmt.Sprintf("%s%s%s", nodeEndpoint, nodeID, "?fields=attributes.temperature")

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nodeInfo, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/vnd.alertme.zoo-6.2+json")
	req.Header.Set("Accept", "application/vnd.alertme.zoo-6.2+json")
	req.Header.Set("X-Omnia-Client", "ESP")
	req.Header.Set("Authorization", "Bearer "+h.token)

	res, err := h.httpClient.Do(req)
	if err != nil {
		return nodeInfo, fmt.Errorf("error requesting node information: %w", err)
	}

	defer res.Body.Close()

	json.NewDecoder(res.Body).Decode(&nodeInfo)

	return nodeInfo, nil
}
