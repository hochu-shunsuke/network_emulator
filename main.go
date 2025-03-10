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

type Layer interface {
	HandleOutgoing(p Packet) Packet
	HandleIncoming(p Packet) Packet
	GetName() string
}

type NetworkLayer struct {
	Name string
	IP   string
}

func (nl *NetworkLayer) HandleOutgoing(p Packet) Packet {
	p.SrcIP = nl.IP
	fmt.Printf("[IP] %s: Sending packet %s\n", nl.IP, p)
	return p
}

func (nl *NetworkLayer) HandleIncoming(p Packet) Packet {
	if p.DstIP == nl.IP {
		fmt.Printf("[IP] %s: Received packet for me: %s\n", nl.IP, p)
	} else {
		fmt.Printf("[IP] %s: Dropping packet (wrong IP): %s\n", nl.IP, p)
	}
	return p
}

func (nl *NetworkLayer) GetName() string {
	return nl.Name
}

type DataLinkLayer struct {
	Name string
	MAC  string
}

func (dl *DataLinkLayer) HandleOutgoing(p Packet) Packet {
	p.SrcMAC = dl.MAC
	fmt.Printf("[MAC] %s: Sending packet %s\n", dl.Name, dl.MAC)
	return p
}

func (dl *DataLinkLayer) HandleIncoming(p Packet) Packet {
	if p.DstMAC == dl.MAC {
		fmt.Printf("[MAC] %s: Received packet for me: %s\n", dl.Name, p)
	} else {
		fmt.Printf("[MAC] %s: Dropping packet (wrong MAC): %s\n", dl.Name, p)
	}
	return p
}

func (dl *DataLinkLayer) GetName() string {
	return dl.Name
}

type Host struct {
	Name   string
	Layers []Layer
}

// Packetを引数で受け取り送信する
func (h *Host) SendPacket(p Packet) {
	fmt.Printf("%s is sending packet\n", h.Name)
	//高レイヤから低レイヤへ
	for i := len(h.Layers) - 1; i >= 0; i-- {
		p = h.Layers[i].HandleOutgoing(p)
	}
}

func (h *Host) ReceivePacket(p Packet) {
	fmt.Printf("%s is receiving packet\n", h.Name)
	//低レイヤから高レイヤへ
	for _, layer := range h.Layers {
		p = layer.HandleIncoming(p)
	}
}

func (h *Host) GetName() string {
	return h.Name
}

// Switch（Deviceインターフェースを満たす）
type Switch struct {
	Name  string
	Ports map[string]Device
}

func (s *Switch) SendPacket(p Packet) {
	if dst, exists := s.Ports[p.DstMAC]; exists { //dst, existsで、値と存在の2つを返す
		fmt.Printf("[Switch] Forwarding packet to %s\n", p.DstMAC)
		dst.ReceivePacket(p)
	} else {
		fmt.Printf("[Switch] Destination MAC %s not found\n", p.DstMAC)
	}
}

func (s *Switch) ReceivePacket(p Packet) {
	s.SendPacket(p)
}

func (s *Switch) GetName() string {
	return s.Name
}

// Router（Deviceインターフェースを満たす）
type Router struct {
	Name  string
	Ports map[string]Device
}

func (r *Router) SendPacket(p Packet) {
	if nextHop, exists := r.Ports[p.DstIP]; exists {
		fmt.Printf("[Router] Forwarding packet to %s\n", p.DstIP)
		nextHop.ReceivePacket(p)
	} else {
		fmt.Printf("[Router] No route to %s\n", p.DstIP)
	}
}
func (r *Router) ReceivePacket(p Packet) {
	r.SendPacket(p)
}

func (r *Router) GetName() string {
	return r.Name
}

// Packetを作成、送信と受信を行う。
func main() {
	host1 := &Host{Name: "Host1", Layers: []Layer{
		&DataLinkLayer{Name: "DataLink", MAC: "AA:BB:CC:DD:EE:01"},
		&NetworkLayer{Name: "Network", IP: "192.168.1.1"},
	}}
	host2 := &Host{Name: "Host2", Layers: []Layer{
		&DataLinkLayer{Name: "DataLink", MAC: "AA:BB:CC:DD:EE:02"},
		&NetworkLayer{Name: "Network", IP: "192.168.1.2"},
	}}
	switch1 := &Switch{Name: "Switch1", Ports: map[string]Device{"AA:BB:CC:DD:EE:01": host1, "AA:BB:CC:DD:EE:02": host2}}
	packet := Packet{Data: "Hello Network!!", SrcIP: "192.168.1.1", DstIP: "192.168.1.2", SrcMAC: "AA:BB:CC:DD:EE:01", DstMAC: "AA:BB:CC:DD:EE:02"}
	host1.SendPacket(packet)
	switch1.ReceivePacket(packet)
}

//現在の問題点
/*
162、163行目に記述していることを確認すると、
host1がsendするとswitch1がreceiveする。
switch1はパケットの宛先MACアドレスを確認して、そのホストのreceive関数を呼び出している。
そのため、実際には通信のモックが行われている。

通信ではなく、独立した関数の呼び出しになっているため、
改善しないと支障が出ると予想するし、何よりロマンがない

改善方法についてはまだ未定。
*/
