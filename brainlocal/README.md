# BrainLocal

BrainLocal 是 ZenModel 中 Brain 接口的一种基于内存的实现。

## 简介

BrainLocal 提供了一个完全在内存中运行的 Brain 实例。它适用于单机环境下的 Brain 操作,无需额外的存储或分布式系统支持。

主要特点:

- 完全基于内存,运行速度快
- 适用于单机环境
- 实现了 Brain 接口的所有功能
- 支持异步运行和并发处理

## 使用方法

以下是一个简单的使用 BrainLocal 的示例:

```go
package main

import (
    "fmt"
    "github.com/zenmodel/zenmodel/brainlocal"
    "github.com/zenmodel/zenmodel"
)

func main() {
    bp := zenmodel.NewBlueprint()
    
    // 绘制蓝图 ...

    // 构建大脑
    brain := brainlocal.BuildBrain(bp)

    // 设置内存并触发所有入口链接
    _ = brain.EntryWithMemory(
        "messages", []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "What is the weather in Boston today?"}})

    // 阻塞进程直到大脑进入休眠状态
    brain.Wait()

    messages, _ := json.Marshal(brain.GetMemory("messages"))
    fmt.Printf("messages: %s\n", messages)
}
```