# BrainLocal 技术设计

## 1. 概述

BrainLocal 是 ZenModel 框架中 Brain 接口的一个基于内存的实现。它提供了一个完全在内存中运行的 Brain 实例,适用于单机环境下的 Brain 操作,无需额外的存储或分布式系统支持。

## 2. 核心模块

### 2.1 BrainLocal 结构体

BrainLocal 结构体是 Brain 接口的实现,包含以下主要字段:

- id: Brain 的唯一标识符
- labels: Brain 的标签
- neurons: Neuron 的索引，存储所有 Neuron 的映射
- links: Link 的索引，存储所有 Link 的映射
- state: Brain 的当前状态
- mu: Brain 状态的读写锁
- cond: 用于判断 Brain 是否是预期状态，实现 Brain 的 Wait() 方法

### 2.2 Neuron 结构体

Neuron 代表计算单元,主要包含:

- id: Neuron 的唯一标识符
- processor: 处理逻辑
- inLinks: 入向连接
- outLinks: 出向连接
- triggerGroups: 触发组
- castGroups: 传播组

### 2.3 Link 结构体

Link 代表 Neuron 之间的连接,包含:

- id: Link 的唯一标识符
- spec: 连接规格(源和目标 Neuron)
- status: 连接状态

### 2.4 Brain Memory

BrainMemory 是 Brain 的上下文实现，使用 [Ristretto](https://github.com/dgraph-io/ristretto) 缓存库来实现高效的上下文管理:

- cache: Ristretto 缓存实例
- numCounters: 用于跟踪频率的键数量
- maxCost: 缓存的最大成本

### 2.5 Brain Maintainer

BrainMaintainer 负责管理 Brain 的运行状态, 通过 channel 管理各类事件来推动 Brain 的运行:

- bQueue: 用于处理 Brain 事件的通道
- stop: 用于停止 Brain 的通道
- NeuronRunner: 负责 Neuron 的并发执行

### 2.6 NeuronRunner

NeuronRunner 是 BrainMaintainer 内的一部分，专注于管理 Neuron 的并发执行:

- nQueue: Neuron 执行队列
- nQueueLen: 队列长度
- nWorkerNum: 工作线程数量

## 3. 主要流程

### 3.1 Brain 构建

1. 通过 BuildBrain 函数创建 BrainLocal 实例
2. 根据传入的 Blueprint 创建 Neuron 和 Link
3. 初始化配置(如日志、工作线程数等)

### 3.2 Brain 运行

1. 通过 Entry 或 TrigLinks 方法触发 Brain 运行
2. Brain 根据触发的 Link 激活相应的 Neuron
3. Neuron 执行处理逻辑,可能会读写 Memory
4. 根据 Neuron 的输出和 Link 的配置,继续激活下游 Neuron

## 4. 并发控制

- 使用互斥锁和条件变量保证 Brain 操作的线程安全
- 支持并发执行多个 Neuron
- 提供 Wait 方法等待 Brain 执行完成

## 5. 性能考虑

- 使用 Ristretto 缓存提高内存读写速度
- 支持配置工作线程数,平衡资源使用和并发度

## 6. 未来优化方向

- 增强监控和调试功能,便于问题排查
- 优化内存管理策略,提高大规模数据处理能力
