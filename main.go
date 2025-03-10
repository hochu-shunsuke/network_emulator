package main

import "fmt"

// 基本的なパケット構造体
type Packet struct {
	Data   string
	SrcIP  string
	DstIP  string
	SrcMAC string
	DstMAC string
}

func (p Packet) String() string {
	return fmt.Sprintf("From %s (%s) to %s (%s): %s", p.SrcIP, p.SrcMAC, p.DstIP, p.DstMAC, p.Data)
}

// ネットワークデバイスのインターフェース（パケットの送受信を定義）
type Device interface {
	SendPacket(p Packet)    //パケット送信メソッド
	ReceivePacket(p Packet) //パケット受信メソッド
	GetName() string        //デバイス名取得メソッド
}

type Host struct {
	Name string
}

// Packetを引数で受け取り送信する
func (h *Host) SendPacket(p Packet) {
	fmt.Printf("%s is sending packet: %s\n", h.Name, p)
}

func (h *Host) ReceivePacket(p Packet) {
	fmt.Printf("%s is receiving packet: %s\n", h.Name, p)
}

func (h *Host) GetName() string {
	return h.Name
}

// Packetを作成、送信と受信を行う。
func main() {
	host1 := &Host{Name: "Host1"}
	host2 := &Host{Name: "Host2"}
	packet := Packet{Data: "Hello Network!!", SrcIP: "192.168.1.1", DstIP: "192.168.1.2", SrcMAC: "AA:BB:CC:DD:EE:01", DstMAC: "AA:BB:CC:DD:EE:02"}
	host1.SendPacket(packet)
	host2.ReceivePacket(packet)
}
