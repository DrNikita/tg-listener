package telegram

// #cgo CFLAGS: -IC:/td/tdlib/include
// #cgo LDFLAGS: -LC:/td/tdlib/bin -ltdjson
// #include <stdlib.h>
// #include "td/telegram/td_json_client.h"
import "C"
import (
	"encoding/json"
	"log/slog"
	"tg-listener/configs"
	"unsafe"
)

type telegramRepository struct {
	config *configs.TgConfig
	logger *slog.Logger
}

func NewTelegramRepository(config *configs.TgConfig, logger *slog.Logger) telegramRepository {
	return telegramRepository{
		config: config,
		logger: logger,
	}
}

const (
	WAIT_TIMEOUT = 10.0
	is_closed    = 0 // should be set to 1, when updateAuthorizationState with authorizationStateClosed is received
)

type TdlibParameters struct {
	Type       string      `json:"@type,omitempty"`
	Extra      string      `json:"@extra,omitempty"`
	ClientID   string      `json:"@client_id,omitempty"`
	Parameters interface{} `json:"parameters,omitempty"`
}

func (tr *telegramRepository) MonitorChannel(channelUsername string) {
	client := C.td_json_client_create()
	defer C.td_json_client_destroy(client)

	// Set TDLib parameters
	parameters := TdlibParameters{
		Type: "setTdlibParameters",
		Parameters: map[string]interface{}{
			"database_directory":      "tdlib",
			"files_directory":         "tdlib/files",
			"database_encryption_key": "",
			"use_file_database":       true,
			"use_chat_info_database":  true,
			"use_message_database":    true,
			"use_secret_chats":        true,
			"api_id":                  tr.config.ApiID,
			"api_hash":                tr.config.ApiHash,
			"system_language_code":    "en",
			"device_model":            "Desktop",
			"system_version":          "1.0",
			"application_version":     "1.0",
		},
	}
	jsonData, err := json.Marshal(parameters)
	if err != nil {
		tr.logger.Error("Ошибка сериализации JSON", "error", err)
		return
	}
	jsonStr := C.CString(string(jsonData))
	defer C.free(unsafe.Pointer(jsonStr))
	C.td_json_client_send(client, jsonStr)

	for {
		response := C.td_json_client_receive(client, 1.0)
		if response != nil {
			result := C.GoString(response)
			tr.logger.Debug("Получен ответ", "response", result)

			// Handle authorization states
			var update map[string]interface{}
			if err := json.Unmarshal([]byte(result), &update); err != nil {
				tr.logger.Error("Ошибка десериализации JSON", "error", err)
				continue
			}

			if update["@type"] == "updateAuthorizationState" {
				authState := update["authorization_state"].(map[string]interface{})
				switch authState["@type"] {
				case "authorizationStateWaitTdlibParameters":
					C.td_json_client_send(client, jsonStr)
				case "authorizationStateWaitEncryptionKey":
					encryptionKey := map[string]interface{}{
						"@type": "checkDatabaseEncryptionKey",
						"key":   "",
					}
					keyData, _ := json.Marshal(encryptionKey)
					keyStr := C.CString(string(keyData))
					defer C.free(unsafe.Pointer(keyStr))
					C.td_json_client_send(client, keyStr)
				case "authorizationStateWaitPhoneNumber":
					phoneNumber := map[string]interface{}{
						"@type":        "setAuthenticationPhoneNumber",
						"phone_number": tr.config.PhoneNumber,
					}
					phoneData, _ := json.Marshal(phoneNumber)
					phoneStr := C.CString(string(phoneData))
					defer C.free(unsafe.Pointer(phoneStr))
					C.td_json_client_send(client, phoneStr)
				case "authorizationStateWaitCode":
					// Here you should implement a way to get the code from the user
					code := map[string]interface{}{
						"@type": "checkAuthenticationCode",
						"code":  tr.config.AuthCode,
					}
					codeData, _ := json.Marshal(code)
					codeStr := C.CString(string(codeData))
					defer C.free(unsafe.Pointer(codeStr))
					C.td_json_client_send(client, codeStr)
				case "authorizationStateReady":
					tr.logger.Info("Авторизация прошла успешно")
					// Authorization is complete, you can now use the client
				}
			}
		}
	}
}
