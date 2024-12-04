package mqtt

// MQTT upgrade command structures
type UpgradeCommand struct {
	ConfirmationTopic string      `json:"confirmation_topic"`
	MessageID        int         `json:"message_id"`
	MessageUUID      string      `json:"message_uuid"`
	RequestType      string      `json:"request_type"`
	Data             CommandData `json:"data"`
}

type CommandData struct {
	CmdType string      `json:"cmd_type"`
	Data    UpgradeData `json:"data"`
}

type UpgradeData struct {
	AppVersion    string `json:"app_version"`
	DownloadURL   string `json:"download_url"`
	CreatedAt     string `json:"created_at"`
	DeviceType    int    `json:"device_type"`
	Enabled       bool   `json:"enabled"`
	PackageName   string `json:"package_name"`
}

type UpgradeResponse struct {
	Code         int    `json:"code"`
	Data         string `json:"data"`
	MacAddress   string `json:"mac_address"`
	MessageID    int    `json:"message_id"`
	MessageInfo  string `json:"message_info"`
	MessageUUID  string `json:"message_uuid"`
	ProductKey   string `json:"product_key"`
	ResponseType string `json:"response_type"`
	Status       string `json:"status"`
	Time         int64  `json:"time"`
} 