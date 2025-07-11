// internal/handler/all.go
package handler

// 这个文件是实现模块化自注册的关键。
// 它本身不需要包含任何导入语句，因为所有 `*.go` 文件都在同一个 `handler` 包内。
// 只要主程序导入了 `handler` 包（即使是匿名导入 `_`），
// Go 语言就会自动执行该包下所有文件中的 `init()` 函数。
//
// 它的存在确保了 `handler` 作为一个整体包的概念，
// 让我们在 `bootstrap` 中可以清晰地写 `_ "dmmserver/internal/handler"`。