package gui

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"device-upgrade/internal/mqtt"
	"device-upgrade/internal/config"
)

type UpgradeWindow struct {
	window       fyne.Window
	mqttClient   *mqtt.Client
	macEntry     *widget.Entry
	urlEntry     *widget.Entry
	versionEntry *widget.Entry
	packageEntry *widget.Entry
	logText      *widget.Entry
	statusLabel  *widget.Label
	mqttStatus   *widget.Label
	upgradeStatus *widget.Label
	currentMac   string
}

func NewUpgradeWindow(cfg *config.Config) *UpgradeWindow {
	myApp := app.New()
	window := myApp.NewWindow("Device Upgrade Tool")
	
	// 创建日志组件
	logText := widget.NewMultiLineEntry()
	logText.Disable()
	logText.MultiLine = true
	logText.Wrapping = fyne.TextWrapWord
	logText.TextStyle = fyne.TextStyle{
		Bold:      true,
		Monospace: true,  // 使用等宽字体，更容易阅读日志
	}

	w := &UpgradeWindow{
		window:        window,
		macEntry:      widget.NewEntry(),
		urlEntry:      widget.NewEntry(),
		versionEntry:  widget.NewEntry(),
		packageEntry:  widget.NewEntry(),
		logText:       logText,
		statusLabel:   widget.NewLabel("MQTT Status: Not Connected"),
		mqttStatus:    widget.NewLabel("MQTT: Disconnected"),
		upgradeStatus: widget.NewLabel("Upgrade Status: Ready"),
		currentMac:    "",
	}

	// 创建MQTT客户端，此时日志组件已经准备好
	client, err := mqtt.NewClient(&cfg.MQTT, 
		// 日志回调
		func(text string) {
			// 确保在主线程中更新UI
			window.Canvas().Refresh(w.logText)
			w.appendLog(text)
			
			// 检查是否是设备响应消息
			if strings.Contains(text, "message_confirmation") {
				var response mqtt.UpgradeResponse
				if err := json.Unmarshal([]byte(text), &response); err == nil {
					// 检查设备地址是否匹配
					if response.MacAddress == w.currentMac {
						if response.Code == 200 && response.Status == "ok" {
							w.setUpgradeStatus("Command Sent Successfully")
							w.appendLog(fmt.Sprintf("Device %s confirmed upgrade command", w.currentMac))
						} else {
							w.setUpgradeStatus(fmt.Sprintf("Command Failed: %s", response.MessageInfo))
							w.appendLog(fmt.Sprintf("Device %s reported error: %s", w.currentMac, response.MessageInfo))
						}
					}
				}
			}
		},
		// 状态回调
		func(connected bool) {
			if connected {
				w.setMQTTStatus("Connected")
			} else {
				w.setMQTTStatus("Disconnected")
			}
		},
	)
	if err != nil {
		w.appendLog(fmt.Sprintf("Failed to create MQTT client: %v", err))
		w.setMQTTStatus("Disconnected")
	} else {
		w.mqttClient = client
		w.setMQTTStatus("Connected")
	}

	w.setupUI()
	return w
}

func (w *UpgradeWindow) setupUI() {
	// Status bar
	statusBar := container.NewHBox(
		w.statusLabel,
		layout.NewSpacer(),
	)

	// MAC address input
	w.macEntry.SetPlaceHolder("Enter device MAC address, e.g.: aa:bb:cc:dd:ee:ff")
	macForm := widget.NewForm(
		widget.NewFormItem("Device MAC Address", w.macEntry),
	)
	
	// Upgrade information input
	w.urlEntry.SetPlaceHolder("Enter firmware download URL")
	w.versionEntry.SetPlaceHolder("Enter new version, e.g.: V1.0.0")
	w.packageEntry.SetPlaceHolder("Enter package name, e.g.: firmware.img")
	upgradeForm := widget.NewForm(
		widget.NewFormItem("Firmware URL", w.urlEntry),
		widget.NewFormItem("Version", w.versionEntry),
		widget.NewFormItem("Package Name", w.packageEntry),
	)

	// Button area
	upgradeBtn := widget.NewButton("Send Upgrade Command", w.sendUpgradeCommand)
	clearBtn := widget.NewButton("Clear Log", func() {
		w.logText.SetText("")
	})
	buttons := container.NewHBox(
		upgradeBtn,
		layout.NewSpacer(),
		clearBtn,
	)

	// Create scrollable container for log area
	logContainer := container.NewScroll(w.logText)
	logContainer.SetMinSize(fyne.NewSize(600, 300))

	// Create container for input area
	inputArea := container.NewVBox(
		macForm,
		upgradeForm,
		buttons,
	)

	// Create container for log area
	logArea := container.NewVBox(
		widget.NewLabel("Log Output:"),
		logContainer,
	)

	// 状态显示区域 - 使用固定高度的容器
	statusArea := container.NewHBox(
		w.mqttStatus,
		layout.NewSpacer(),
		w.upgradeStatus,
	)
	statusAreaContainer := container.NewPadded(statusArea)

	// 主布局 - 使用VBox而不是BorderLayout
	content := container.NewVBox(
		statusBar,
		container.NewPadded(inputArea),
		widget.NewSeparator(),
		container.NewPadded(logArea),
		widget.NewSeparator(),
		statusAreaContainer,  // 添加状态区域到底部
	)

	// Set main window content and size
	w.window.SetContent(content)
	w.window.Resize(fyne.NewSize(800, 600))
}

// 修改状态设置函数，确保在主线程中更新UI
func (w *UpgradeWindow) setMQTTStatus(status string) {
	w.window.Canvas().Refresh(w.mqttStatus)
	w.mqttStatus.SetText(fmt.Sprintf("MQTT: %s", status))
}

func (w *UpgradeWindow) setUpgradeStatus(status string) {
	w.window.Canvas().Refresh(w.upgradeStatus)
	w.upgradeStatus.SetText(fmt.Sprintf("Upgrade Status: %s", status))
}

func (w *UpgradeWindow) sendUpgradeCommand() {
	mac := w.macEntry.Text
	url := w.urlEntry.Text
	version := w.versionEntry.Text
	packageName := w.packageEntry.Text

	if mac == "" || url == "" || version == "" || packageName == "" {
		w.appendLog("Error: MAC address, URL, version and package name cannot be empty")
		return
	}

	// 保存当前操作的设备MAC地址
	w.currentMac = mac
	w.setUpgradeStatus("Upgrading...")

	// Send upgrade command
	err := w.mqttClient.SendUpgradeCommand(mac, version, url, packageName)
	if err != nil {
		w.appendLog(fmt.Sprintf("Failed to send upgrade command: %v", err))
		w.setUpgradeStatus("Command Failed")
		return
	}

	w.appendLog(fmt.Sprintf("Sent upgrade command to device: %s", mac))
}

func (w *UpgradeWindow) appendLog(text string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logLine := fmt.Sprintf("[%s] %s\n", timestamp, text)
	currentText := w.logText.Text
	w.logText.SetText(currentText + logLine)
	
	// Scroll to latest log
	w.logText.CursorRow = len(w.logText.Text)
}

func (w *UpgradeWindow) Show() {
	w.window.ShowAndRun()
} 