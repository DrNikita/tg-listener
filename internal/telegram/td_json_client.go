package telegram

// #cgo CFLAGS: -IC:/Users/helll/td/tdlib/include
// #cgo LDFLAGS: -LC:/Users/helll/td/tdlib/bin -ltdjson
// #include <stdlib.h>
// #include "td/telegram/td_json_client.h"
import "C"
import (
	"fmt"
	"strings"
)

const (
	WAIT_TIMEOUT = 10.0
	is_closed    = 0 // should be set to 1, when updateAuthorizationState with authorizationStateClosed is received
)

func MonitorChannel(channelUsername string) {
	client := C.td_json_client_create()
	defer C.td_json_client_destroy(client)

	C.td_json_client_send(client, C.CString(`{
		"@type": "setTdlibParameters",
		"parameters": {
			"database_directory": "tdlib",
			"use_file_database": true,
			"use_chat_info_database": true,
			"use_message_database": true,
			"use_secret_chats": false,
			"api_id": <Your_API_ID>,
			"api_hash": "<Your_API_Hash>",
			"system_language_code": "en",
			"device_model": "Desktop",
			"system_version": "Unknown",
			"application_version": "1.0",
			"enable_storage_optimizer": true,
			"ignore_file_names": false
		}
	}`))


	// Поиск канала
	C.td_json_client_send(client, C.CString(fmt.Sprintf(`{
      "@type": "searchPublicChat",
      "username": "%s"
    }`, channelUsername)))

	// Мониторинг обновлений
	for {
		result := C.td_json_client_receive(client, WAIT_TIMEOUT)
		if result != nil {
			response := C.GoString(result)
			fmt.Println("Update:", response)

			// Если это новый канал найден
			if strings.Contains(response, `"@type":"chat"`) {
				chatID := extractChatID(response)
				fmt.Printf("Chat ID: %d\n", chatID)

				// Получение истории сообщений
				C.td_json_client_send(client, C.CString(fmt.Sprintf(`{
                  "@type": "getChatHistory",
                  "chat_id": %d,
                  "from_message_id": 0,
                  "offset": 0,
                  "limit": 10,
                  "only_local": false
                }`, chatID)))
			}

			// Если это новое сообщение
			if strings.Contains(response, `"@type":"updateNewMessage"`) {
				fmt.Println("New message:", response)
			}
		}
	}
}

// Вспомогательная функция для извлечения chat_id
func extractChatID(response string) int64 {
	// Логика для извлечения ID чата из JSON-ответа
	fmt.Println(response)
	return 123456789 // Замените на реальную логику
}
