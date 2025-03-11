package main

import (
	"container/heap"
	"fmt"
	"time"
)

// Packetはネットワークパケットを表し、送信元/宛先IPとMACアドレス、データペイロードを持つ。
type Packet struct {
	Data   string // パケットのデータ部分
	SrcIP  string // 送信元のIPアドレス
	DstIP  string // 宛先のIPアドレス
	SrcMAC string // 送信元のMACアドレス
	DstMAC string // 宛先のMACアドレス
}

// Stringはデバッグ用にパケットを人間が読める形式で返す。
func (p Packet) String() string {
	return fmt.Sprintf("From %s (%s) to %s (%s): %s", p.SrcIP, p.SrcMAC, p.DstIP, p.DstMAC, p.Data)
}

// Deviceはネットワークデバイス（ホスト、スイッチ、ルータ）のインターフェースを定義。
type Device interface {
	SendPacket(p Packet)    // パケットを次のデバイスに送信
	ReceivePacket(p Packet) // 他のデバイスからパケットを受信
	GetName() string        // デバイスの名前をログ用に返す
}

// Layerはプロトコル層（例：ネットワーク層、データリンク層）のインターフェースを定義。
type Layer interface {
	HandleOutgoing(p Packet) Packet // 送信パケットを処理（例：ヘッダ追加）
	HandleIncoming(p Packet) Packet // 受信パケットを処理（例：ヘッダ検証）
	GetName() string                // 層の名前をログ用に返す
}

// NetworkLayerはOSIモデルのIP層を表す。
type NetworkLayer struct {
	Name string // 層の名前（デバッグ用）
	IP   string // この層に割り当てられたIPアドレス
}

// HandleOutgoingは送信パケットに送信元IPを設定。
func (nl *NetworkLayer) HandleOutgoing(p Packet) Packet {
	p.SrcIP = nl.IP
	fmt.Printf("[IP] %s: パケット送信中 %s\n", nl.IP, p) // IP層の動作をログ
	return p
}

// HandleIncomingはパケットの宛先IPがこのデバイスのIPと一致するか確認。
func (nl *NetworkLayer) HandleIncoming(p Packet) Packet {
	if p.DstIP == nl.IP {
		fmt.Printf("[IP] %s: 自分宛のパケットを受信: %s\n", nl.IP, p) // 受信成功をログ
	} else {
		fmt.Printf("[IP] %s: IPが一致しないためパケットを破棄: %s\n", nl.IP, p) // 破棄をログ
	}
	return p
}

func (nl *NetworkLayer) GetName() string {
	return nl.Name
}

// DataLinkLayerはOSIモデルのMAC層を表す。
type DataLinkLayer struct {
	Name string // 層の名前（デバッグ用）
	MAC  string // この層に割り当てられたMACアドレス
}

// HandleOutgoingは送信パケットに送信元MACを設定。
func (dl *DataLinkLayer) HandleOutgoing(p Packet) Packet {
	p.SrcMAC = dl.MAC
	fmt.Printf("[MAC] %s: パケット送信中 %s\n", dl.Name, dl.MAC) // MAC層の動作をログ
	return p
}

// HandleIncomingはパケットの宛先MACがこのデバイスのMACと一致するか確認。
func (dl *DataLinkLayer) HandleIncoming(p Packet) Packet {
	if p.DstMAC == dl.MAC {
		fmt.Printf("[MAC] %s: 自分宛のパケットを受信: %s\n", dl.Name, p) // 受信成功をログ
	} else {
		fmt.Printf("[MAC] %s: MACが一致しないためパケットを破棄: %s\n", dl.Name, p) // 破棄をログ
	}
	return p
}

func (dl *DataLinkLayer) GetName() string {
	return dl.Name
}

// Eventはネットワークイベント（例：パケット送信）を表す。
type Event struct {
	Time    time.Time // イベントが発生する時刻
	Handler func()    // イベント発生時に実行する関数
}

// EventQueueは時間順にイベントを管理する優先度キュー。
type EventQueue []*Event

func (eq EventQueue) Len() int            { return len(eq) }
func (eq EventQueue) Less(i, j int) bool  { return eq[i].Time.Before(eq[j].Time) }
func (eq EventQueue) Swap(i, j int)       { eq[i], eq[j] = eq[j], eq[i] }
func (eq *EventQueue) Push(x interface{}) { *eq = append(*eq, x.(*Event)) }
func (eq *EventQueue) Pop() interface{} {
	old := *eq
	n := len(old)
	x := old[n-1]
	*eq = old[0 : n-1]
	return x
}

// EventBusは非同期パケット送信のためのイベントキューを管理。
type EventBus struct {
	Events EventQueue // スケジュールされたイベントのキュー
}

var eventBus = &EventBus{Events: make(EventQueue, 0)} // グローバルなイベントバス

// AddEventは遅延時間後に実行されるイベントを追加。
func (eb *EventBus) AddEvent(delay time.Duration, handler func()) {
	time := time.Now().Add(delay)
	event := &Event{Time: time, Handler: handler}
	heap.Push(&eb.Events, event)
	fmt.Printf("[EventBus] イベントを追加: 遅延 %v\n", delay) // イベント追加をログ
}

// Runはイベントキューを実行し、時間順にハンドラを呼び出す。
func (eb *EventBus) Run() {
	for eb.Events.Len() > 0 {
		event := heap.Pop(&eb.Events).(*Event)
		now := time.Now()
		if now.Before(event.Time) {
			fmt.Printf("[EventBus] 待機中: %v\n", event.Time.Sub(now)) // 待機時間をログ
			time.Sleep(event.Time.Sub(now))
		}
		event.Handler()
		fmt.Printf("[EventBus] イベント実行完了\n") // イベント実行をログ
	}
}

// Linkはデバイス間の接続を表し、遅延をシミュレート。
type Link struct {
	From  Device        // 送信元デバイス
	To    Device        // 宛先デバイス
	Delay time.Duration // 伝送遅延時間
}

// Transmitはパケットをリンク経由で送信（イベントバスを使用）。
func (l *Link) Transmit(p Packet) {
	fmt.Printf("リンク: %s から %s へパケット送信中、遅延 %v\n", l.From.GetName(), l.To.GetName(), l.Delay)
	eventBus.AddEvent(l.Delay, func() {
		l.To.ReceivePacket(p)
	})
}

// Networkはネットワークトポロジーを管理。
type Network struct {
	Devices []Device // ネットワーク内の全デバイス
	Links   []*Link  // デバイス間の全リンク
}

// AddDeviceはネットワークにデバイスを追加。
func (n *Network) AddDevice(d Device) {
	n.Devices = append(n.Devices, d)
	fmt.Printf("[Network] デバイス追加: %s\n", d.GetName()) // デバイス追加をログ
}

// AddLinkはデバイス間にリンクを追加。
func (n *Network) AddLink(from, to Device, delay time.Duration) {
	link := &Link{From: from, To: to, Delay: delay}
	n.Links = append(n.Links, link)
	fmt.Printf("[Network] リンク追加: %s -> %s\n", from.GetName(), to.GetName()) // リンク追加をログ
}

// GetLinkは指定されたデバイス間のリンクを返す（存在しない場合はnil）。
func (n *Network) GetLink(from, to Device) *Link {
	for _, link := range n.Links {
		if link.From == from && link.To == to {
			return link
		}
	}
	fmt.Printf("[Network] リンクが見つかりません: %s -> %s\n", from.GetName(), to.GetName()) // リンク未発見をログ
	return nil
}

var network = &Network{} // グローバルなネットワークインスタンス

// Hostはネットワークホストを表す。
type Host struct {
	Name         string  // ホストの名前
	Layers       []Layer // プロトコル層のスタック
	ConnectedDev Device  // 接続先デバイス（例：スイッチ）
}

// SendPacketはパケットを送信し、レイヤーを経由して接続先へ転送。
func (h *Host) SendPacket(p Packet) {
	fmt.Printf("%s がパケットを送信開始\n", h.Name)
	for i := len(h.Layers) - 1; i >= 0; i-- { // 高レイヤから低レイヤへ処理
		p = h.Layers[i].HandleOutgoing(p)
	}
	if h.ConnectedDev != nil {
		link := network.GetLink(h, h.ConnectedDev)
		if link != nil {
			link.Transmit(p)
			fmt.Printf("%s: %s へパケット送信完了\n", h.Name, h.ConnectedDev.GetName())
		} else {
			fmt.Printf("%s: %s へのリンクが見つかりません\n", h.Name, h.ConnectedDev.GetName()) // エラーケースをログ
		}
	} else {
		fmt.Printf("%s: 接続先デバイスが設定されていません\n", h.Name) // 接続先未設定をログ
	}
}

// ReceivePacketは受信パケットを低レイヤから高レイヤへ処理。
func (h *Host) ReceivePacket(p Packet) {
	fmt.Printf("%s がパケットを受信\n", h.Name)
	for _, layer := range h.Layers { // 低レイヤから高レイヤへ処理
		p = layer.HandleIncoming(p)
	}
}

func (h *Host) GetName() string {
	return h.Name
}

// SwitchはL2スイッチを表す。
type Switch struct {
	Name     string            // スイッチの名前
	Ports    map[string]Device // MACアドレスとデバイスのマッピング
	MACTable map[string]Device // 学習したMACアドレスとデバイスのテーブル
	Links    map[Device]*Link  // デバイスごとのリンク
}

// SendPacketはパケットを転送し、MACテーブルを更新。
func (s *Switch) SendPacket(p Packet) {
	if dev, ok := s.Ports[p.SrcMAC]; ok {
		s.MACTable[p.SrcMAC] = dev // 送信元MACを学習
		fmt.Printf("[Switch] %s: MACテーブル更新 %s -> %s\n", s.Name, p.SrcMAC, dev.GetName())
	}
	if dst, exists := s.MACTable[p.DstMAC]; exists {
		fmt.Printf("[Switch] %s: %s へパケット転送\n", s.Name, p.DstMAC)
		link := s.Links[dst]
		link.Transmit(p)
	} else {
		fmt.Printf("[Switch] %s: 不明なMAC %s、ブロードキャスト実行\n", s.Name, p.DstMAC)
		for mac, dev := range s.Ports {
			if mac != p.SrcMAC { // 送信元には送らない
				link := s.Links[dev]
				link.Transmit(p)
			}
		}
	}
}

// ReceivePacketは受信したパケットを転送処理に渡す。
func (s *Switch) ReceivePacket(p Packet) {
	fmt.Printf("[Switch] %s: パケット受信\n", s.Name)
	s.SendPacket(p)
}

func (s *Switch) GetName() string {
	return s.Name
}

// RouterはL3ルータを表す（現在未使用）。
type Router struct {
	Name  string            // ルータの名前
	Ports map[string]Device // IPアドレスとデバイスのマッピング
}

func (r *Router) SendPacket(p Packet) {
	if nextHop, exists := r.Ports[p.DstIP]; exists {
		fmt.Printf("[Router] %s: %s へパケット転送\n", r.Name, p.DstIP)
		nextHop.ReceivePacket(p)
	} else {
		fmt.Printf("[Router] %s: %s への経路なし\n", r.Name, p.DstIP)
	}
}

func (r *Router) ReceivePacket(p Packet) {
	fmt.Printf("[Router] %s: パケット受信\n", r.Name)
	r.SendPacket(p)
}

func (r *Router) GetName() string {
	return r.Name
}

// mainはシミュレーションのエントリーポイント。
func main() {
	// ホスト1の初期化
	host1 := &Host{
		Name: "Host1",
		Layers: []Layer{
			&DataLinkLayer{Name: "DataLink", MAC: "AA:BB:CC:DD:EE:01"},
			&NetworkLayer{Name: "Network", IP: "192.168.1.1"},
		},
	}
	// ホスト2の初期化
	host2 := &Host{
		Name: "Host2",
		Layers: []Layer{
			&DataLinkLayer{Name: "DataLink", MAC: "AA:BB:CC:DD:EE:02"},
			&NetworkLayer{Name: "Network", IP: "192.168.1.2"},
		},
	}
	// スイッチの初期化
	switch1 := &Switch{
		Name: "Switch1",
		Ports: map[string]Device{
			"AA:BB:CC:DD:EE:01": host1,
			"AA:BB:CC:DD:EE:02": host2,
		},
		MACTable: make(map[string]Device),
		Links:    make(map[Device]*Link),
	}

	// ネットワークトポロジーの設定
	network.AddDevice(host1)
	network.AddDevice(host2)
	network.AddDevice(switch1)
	network.AddLink(host1, switch1, 50*time.Millisecond) // ホスト1 -> スイッチ
	network.AddLink(switch1, host1, 50*time.Millisecond) // スイッチ -> ホスト1
	network.AddLink(switch1, host2, 50*time.Millisecond) // スイッチ -> ホスト2

	// ホストとスイッチの接続設定
	host1.ConnectedDev = switch1
	switch1.Links[host1] = network.GetLink(switch1, host1)
	switch1.Links[host2] = network.GetLink(switch1, host2)

	// パケットの作成と送信
	packet := Packet{Data: "Hello Network!!", SrcIP: "192.168.1.1", DstIP: "192.168.1.2", SrcMAC: "AA:BB:CC:DD:EE:01", DstMAC: "AA:BB:CC:DD:EE:02"}
	fmt.Printf("[Main] パケット送信開始: %s\n", packet)
	host1.SendPacket(packet)

	// イベントバスの実行
	fmt.Printf("[Main] イベントバス実行開始\n")
	eventBus.Run()
	fmt.Printf("[Main] シミュレーション終了\n")
}
