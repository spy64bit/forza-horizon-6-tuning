package main

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"time"
)

func modeLive(ctx context.Context, onPacket func(ForzaHorizonPacket)) {
	addr, err := net.ResolveUDPAddr("udp", getListenAddr())
	if err != nil {
		fmt.Printf("Error resolving address: %v\n", err)
		return
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Printf("Error opening socket: %v\n", err)
		return
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			fmt.Printf("Read error: %v\n", err)
			continue
		}
		if n != packetSize {
			continue
		}
		p, err := parsePacket(buf[:n])
		if err == nil {
			onPacket(p)
		}
	}
}

func modeRecord(ctx context.Context, recordFile string, onPacket func(ForzaHorizonPacket)) {
	addr, err := net.ResolveUDPAddr("udp", getListenAddr())
	if err != nil {
		fmt.Printf("Error resolving address: %v\n", err)
		return
	}
	var conn *net.UDPConn
	for i := 0; i < 10; i++ {
		conn, err = net.ListenUDP("udp", addr)
		if err == nil {
			break
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(100 * time.Millisecond):
		}
	}
	if err != nil {
		fmt.Printf("Error opening socket: %v\n", err)
		return
	}
	defer conn.Close()

	f, err := os.Create(recordFile)
	if err != nil {
		fmt.Printf("Cannot create file: %v\n", err)
		return
	}

	writer := bufio.NewWriterSize(f, 64*1024)
	defer func() {
		writer.Flush()
		f.Close()
	}()

	buf := make([]byte, 1024)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			fmt.Printf("Read error: %v\n", err)
			continue
		}
		if n != packetSize {
			continue
		}

		ts := time.Now().UnixNano()
		binary.Write(writer, binary.LittleEndian, ts)
		writer.Write(buf[:n])

		p, err := parsePacket(buf[:n])
		if err == nil {
			onPacket(p)
		}
	}
}
