package main

import (
	"fmt"
	"time"

	"github.com/zenmodel/zenmodel"
)

func main() {
	//brainPrint := zenmodel.NewBrainPrint()
	//neuron1 := brainPrint.AddNeuron(fn1)
	//neuron2 := brainPrint.AddNeuron(fn2)
	//neuron3 := brainPrint.AddNeuron(fn3)
	//neuron4 := brainPrint.AddNeuron(fn4)
	//brainPrint.AddNeuronWithProcessor()
	//
	//entry1, err := brainPrint.AddEntryLink(neuron1)
	//_, err = brainPrint.AddLink(neuron1, neuron2)
	//_, err = brainPrint.AddLink(neuron1, neuron3)
	//link24, err := brainPrint.AddLink(neuron2, neuron4)
	//link34, err := brainPrint.AddLink(neuron3, neuron4)
	//err = brainPrint.AddTriggerGroup(neuron2, link24, link34)
	//
	//brain := brainPrint.Build(zenmodel.WithLocalMaintainer()) // Build 	参数默认是 local, 就是使用本地 channel&goroutine 去维护 brain. 当然也需要支持对接远端 MQ, 对接远端（grpc）维护用的计算资源/
	//
	//brain.SetContext()
	//brain.SetContextUnsafe()
	//brain.SetContextField()
	//brain.SetContextFieldunsafe()
	//brain.GetContext()
	//brain.GetOutput() // 如果是 stream , 则是 stream channel 中的元素拼接而成的结果
	//brain.WatchOutput()- > chan
	//
	//
	//brain.GetStatus()
	//
	//
	//// 触发所有 Entry Links
	//brain.Entry()
	//// 触发指定 links。可以从任意 links 来激活 brain
	//brain.TrigLinks()

	// goroutine 在 trigLinks 是ensure 开启,在 brain stop chan 收到消息时关闭

	fn1 := func(b zenmodel.Brain) error {
		fmt.Println("-----------> fn1")
		return nil
	}
	fn2 := func(b zenmodel.Brain) error {
		fmt.Println("-----------> fn2")
		return nil
	}

	bp := zenmodel.NewBrainPrint()
	n1 := bp.AddNeuron(fn1)
	n2 := bp.AddNeuron(fn2)
	_, err := bp.AddLink(n1, n2)
	if err != nil {
		fmt.Printf("add link error --------> %s\n", err)
		return
	}
	_, err = bp.AddEntryLink(n1)
	if err != nil {
		fmt.Printf("add entry link error --------> %s\n", err)
		return
	}
	brain := bp.Build()

	brain.Entry()

	time.Sleep(10 * time.Second)
}
