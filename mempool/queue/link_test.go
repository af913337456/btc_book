package queue

import (
	"testing"
	"fmt"
)


func TestLinkNode_AddAtTail(t *testing.T) {
	link := NewLink(0)  // 以数字0作为头节点数据初始化一个链表
	link.AddAtTail(2)   // 从尾部添加一个节点，数据是 2
	link.AddAtTail(3)
	link.AddAtTail(1)
	link.AddAtTail(44)
	link.AddAtTail("2") // 添加一个字符串类型的 2

	link.Print() // 打印一次链表

	dataNode := link.RemoveFirst() // 移出链表头节点
	fmt.Println("dataNode ===> ",dataNode.Data) // 输出被移出头节点对应的数据

	link.Print() // 再打印次链表，查看节点的情况
}

























