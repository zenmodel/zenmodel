# BrainLite 技术设计文档

## 1. 概述

BrainLite 是 ZenModel 框架中 Brain 接口的另一种轻量级实现。它在 BrainLocal 的基础上进行了一些修改,主要区别在于 BrainMemory 的实现。

## 2. 核心模块

BrainLite 的 Brain、Neuron 和 Link 结构与 BrainLocal 基本相同,此处不再赘述。主要区别在于 BrainMemory 的实现。

### 2.1 Brain Memory

BrainLite 的 BrainMemory 使用 SQLite 数据库实现:

- db: SQLite 数据库连接
- datasourceName: 数据库文件名，默认为 `${brain_id}.db`
- keepMemory: 是否在 Brain Shutdown 后保留数据库文件

这种实现方式相比 BrainLocal 的内存上下文实现,具有以下特点:

- 支持持久化存储,可以在 Brain 重启后恢复上下文
- 可以处理更大规模的数据,不受内存限制
- 是多语言 Processor 实现的基础，sqlite 实现的 BrainMemory 可以支持不同编程语言的 Processor 一起读写

### 2.2. Brain Maintainer

BrainMaintainer 是 BrainLite 的核心组件之一, 目前实现和 brainlocal 一致, 后续要重构来支持多编程语言的 brainContext 实现



## 3. 未来优化方向

- 支持多语言 Processor: 计划在未来版本中支持使用不同编程语言实现的 Processor,增强系统的灵活性和扩展性
