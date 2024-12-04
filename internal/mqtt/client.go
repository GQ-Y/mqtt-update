package mqtt

import (
	"encoding/json"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"device-upgrade/internal/config"
)

type Client struct {
	client      mqtt.Client
	config      *config.MQTTConfig
	logCallback func(string)
	statusCallback func(bool)
}

func NewClient(cfg *config.MQTTConfig, logCallback func(string), statusCallback func(bool)) (*Client, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.Broker)
	opts.SetClientID(cfg.ClientID)
	opts.SetUsername(cfg.Username)
	opts.SetPassword(cfg.Password)
	
	// 设置连接参数
	opts.SetKeepAlive(60 * time.Second)
	opts.SetPingTimeout(10 * time.Second)
	opts.SetConnectTimeout(10 * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(10 * time.Second)
	
	c := &Client{
		config:         cfg,
		logCallback:    logCallback,
		statusCallback: statusCallback,
	}
	
	// 设置连接回调
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		if logCallback != nil {
			logCallback("Connected to MQTT broker")
		}
		if statusCallback != nil {
			statusCallback(true)
		}
	})

	// 设置连接丢失回调
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		if logCallback != nil {
			logCallback(fmt.Sprintf("Connection lost: %v", err))
		}
		if statusCallback != nil {
			statusCallback(false)
		}
	})

	// 设置重连回调
	opts.SetReconnectingHandler(func(client mqtt.Client, opts *mqtt.ClientOptions) {
		if logCallback != nil {
			logCallback("Attempting to reconnect to MQTT broker...")
		}
	})

	if logCallback != nil {
		logCallback(fmt.Sprintf("Connecting to MQTT broker: %s with client ID: %s", cfg.Broker, cfg.ClientID))
	}

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("failed to connect to MQTT broker: %v", token.Error())
	}

	c.client = client

	// 连接成功后订阅回复主题
	if err := c.Subscribe(cfg.Topics.Upgrade, func(client mqtt.Client, msg mqtt.Message) {
		if logCallback != nil {
			logCallback(fmt.Sprintf("Received message on topic %s: %s", msg.Topic(), string(msg.Payload())))
		}
	}); err != nil {
		return nil, fmt.Errorf("failed to subscribe to response topic: %v", err)
	}

	if logCallback != nil {
		logCallback(fmt.Sprintf("Successfully subscribed to topic: %s", cfg.Topics.Upgrade))
	}

	return c, nil
}

func (c *Client) Subscribe(topic string, callback mqtt.MessageHandler) error {
	if c.logCallback != nil {
		c.logCallback(fmt.Sprintf("Subscribing to topic: %s", topic))
	}
	
	token := c.client.Subscribe(topic, 1, func(client mqtt.Client, msg mqtt.Message) {
		if callback != nil {
			callback(client, msg)
		}
		// 检查是否是升级响应消息
		var response UpgradeResponse
		if err := json.Unmarshal(msg.Payload(), &response); err == nil {
			if response.ResponseType == "message_confirmation" {
				if c.logCallback != nil {
					c.logCallback(fmt.Sprintf("Received upgrade response: code=%d, status=%s", 
						response.Code, response.Status))
				}
			}
		}
	})
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %v", topic, token.Error())
	}
	
	if c.logCallback != nil {
		c.logCallback(fmt.Sprintf("Successfully subscribed to topic: %s", topic))
	}
	return nil
}

func (c *Client) Publish(topic string, payload interface{}) error {
	token := c.client.Publish(topic, 1, false, payload)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to publish to topic %s: %v", topic, token.Error())
	}
	return nil
}

func (c *Client) Disconnect() {
	if c.logCallback != nil {
		c.logCallback("Disconnecting from MQTT broker")
	}
	c.client.Disconnect(250)
}

func (c *Client) SendUpgradeCommand(mac string, version string, url string, packageName string) error {
	// 构建并发送升级命令
	topic := fmt.Sprintf("/hiot/%s/request_setting", mac)
	
	if c.logCallback != nil {
		c.logCallback(fmt.Sprintf("Preparing upgrade command for device: %s", mac))
	}

	cmd := UpgradeCommand{
		ConfirmationTopic: c.config.Topics.Upgrade,
		MessageID:        0,
		MessageUUID:      fmt.Sprintf("%d", time.Now().Unix()),
		RequestType:      "device_cmd",
		Data: CommandData{
			CmdType: "upgrade_app",
			Data: UpgradeData{
				AppVersion:  version,
				DownloadURL: url,
				CreatedAt:   time.Now().Format("2006-01-02 15:04:05"),
				DeviceType:  2,
				Enabled:     true,
				PackageName: packageName,
			},
		},
	}

	if c.logCallback != nil {
		c.logCallback(fmt.Sprintf("Using confirmation topic: %s", c.config.Topics.Upgrade))
	}

	payload, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("failed to marshal upgrade command: %v", err)
	}

	if c.logCallback != nil {
		c.logCallback(fmt.Sprintf("Publishing upgrade command to topic: %s", topic))
		c.logCallback(fmt.Sprintf("Command payload: %s", string(payload)))
	}

	err = c.Publish(topic, payload)
	if err != nil {
		if c.logCallback != nil {
			c.logCallback(fmt.Sprintf("Failed to publish command: %v", err))
		}
		return err
	}

	if c.logCallback != nil {
		c.logCallback("Command published successfully")
		c.logCallback("Waiting for device response...")
	}

	return nil
} 