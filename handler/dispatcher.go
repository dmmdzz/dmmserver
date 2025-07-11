// internal/handler/dispatcher.go
package handler

import (
	"log"

	"github.com/gin-gonic/gin"
)

// HandlerFunc 定义了所有 msg_id 处理器的【新】标准函数签名
// 它不再直接操作发送，而是返回一个包含结果数据的map和一个error对象
type HandlerFunc func(c *gin.Context, msgData map[string]interface{}) (map[string]interface{}, error)

// handlerRegistry 是一个全局的 map，作为所有 msg_id 处理器的“注册表”
// key 是 msg_id (字符串), value 是对应的处理函数
var handlerRegistry = make(map[string]HandlerFunc)

// Register 用于向全局注册表中注册一个处理器
// 每个处理器文件(如30008.go)都会在init()中调用此函数来“自我注册”
func Register(msgID string, handler HandlerFunc) {
	if _, exists := handlerRegistry[msgID]; exists {
		log.Printf("Warning: Handler for msg_id %s is being overwritten.", msgID)
	}
	handlerRegistry[msgID] = handler
	log.Printf("Handler registered for msg_id: %s", msgID)
}

// GetHandler 根据 msg_id 从注册表中查找并返回处理器
func GetHandler(msgID string) (HandlerFunc, bool) {
	handler, found := handlerRegistry[msgID]
	log.Printf("GetHandler: Looking up handler for msg_id '%s'. Found: %t", msgID, found) // 添加日志
	return handler, found
}