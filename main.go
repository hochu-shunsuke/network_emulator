package main

import "fmt"

// 基本的なパケット構造体
type Packet struct {
	Data string
}

// {Data: data}のdataのみを返すために書いている
func (p Packet) String() string {
	return p.Data
}

// Packetを引数で受け取り送信する
func sendPacket(p Packet) {
	fmt.Printf("Sending packet: %s\n", p)
}

func receivePacket(p Packet) {
	fmt.Printf("Received packet: %s\n", p)
}

// Packetを作成、送信と受信を行う。
func main() {
	packet := Packet{Data: "Hello ,Network!!"}
	sendPacket(packet)
	receivePacket(packet)
}
