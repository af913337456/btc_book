package queue

import "fmt"


// LinkNode 是链表的节点
type LinkNode struct {
	Data interface{} // 数据域，使用 interface 泛型类型，可以存储任何数据类型
	Next *LinkNode   // 指针对象域
}

// 实例化一个链表，返回头节点
func NewLink(data interface{}) *LinkNode {
	return &LinkNode{Data: data}
}

// 将一个元素添加到链表尾部
func (head *LinkNode) AddAtTail(data interface{}) {
	if head == nil {
		return
	}
	temp := head
	for temp.Next != nil {
		temp = temp.Next // 移动指针位置
	}
	if temp.Next == nil { // 移动到了后面
		newNode := &LinkNode{Data: data}
		temp.Next = newNode // 设置新节点
	}
}

func (head *LinkNode) FillData(dataList []interface{}) {
	if head == nil {
		return
	}
	for _,item := range dataList {
		head.AddAtTail(item)
	}
}

// 从头部移出一个数据
func (head *LinkNode) RemoveFirst() *LinkNode {
	if head == nil {
		return nil
	}
	data := head.Data  // 先读出头节点数据
	// 下面移动链表
	temp := head.Next
	*head = *temp
	return &LinkNode{Data:data} // 以头节点数据实例化个节点返回
}

// 辅助函数：打印链表数据
func (head *LinkNode) Print() {
	if head == nil {
		return
	}
	temp := head
	for temp.Next != nil {
		fmt.Println(temp.Data)
		temp = temp.Next
	}
	fmt.Println(temp.Data,"\n-----")
}

















