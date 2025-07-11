// internal/conf/response_format.go
package conf

// textResponseMsgIDs 存储所有应以纯文本格式响应的 msg_id
// 使用 map[int]bool 模拟一个集合(Set)，查找效率为 O(1)，非常高。
var textResponseMsgIDs = map[int]bool{
	30099: true, // 示例：msg_id 30099 返回纯文本
	50023: true, // 示例：msg_id 50023 返回纯文本
	// 后续所有需要返回纯文本的 msg_id 都添加到这里
}

// IsTextResponse 是一个公共函数，用于检查给定的 msg_id 是否需要以纯文本格式响应
// 默认情况下，如果一个 msg_id 不在此列表中，它将被视为需要 JSON 响应。
func IsTextResponse(msgID int) bool {
	_, found := textResponseMsgIDs[msgID]
	return found
}