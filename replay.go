package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"
)

const frameSize = 8 + packetSize // int64 timestamp + raw packet

// replayCmd is sent through cmdCh to control an in-progress replay.
type replayCmd struct {
	seek     int  // frame index to jump to; -1 = no seek
	setPause bool // whether to update the paused state
	pause    bool
}

func modeReplay(
	ctx context.Context,
	recordFile string,
	realtime bool,
	cmdCh <-chan replayCmd,
	onPacket func(ForzaHorizonPacket),
	onProgress func(frame, total int),
) {
	f, err := os.Open(recordFile)
	if err != nil {
		fmt.Printf("Cannot open file: %v\n", err)
		return
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		fmt.Printf("Cannot stat file: %v\n", err)
		return
	}
	total := int(info.Size() / int64(frameSize))
	if total == 0 {
		return
	}

	readFrame := func(idx int) (int64, []byte, error) {
		if _, err := f.Seek(int64(idx)*int64(frameSize), io.SeekStart); err != nil {
			return 0, nil, err
		}
		var ts int64
		if err := binary.Read(f, binary.LittleEndian, &ts); err != nil {
			return 0, nil, err
		}
		raw := make([]byte, packetSize)
		if _, err := io.ReadFull(f, raw); err != nil {
			return 0, nil, err
		}
		return ts, raw, nil
	}

	applyCmd := func(cmd replayCmd, paused *bool, frame *int, prevNano *int64) {
		if cmd.setPause {
			*paused = cmd.pause
		}
		if cmd.seek >= 0 {
			*frame = cmd.seek
			if *frame >= total {
				*frame = total - 1
			}
			*prevNano = 0
		}
	}

	paused := false
	frame := 0
	var prevNano int64

	for frame < total {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Drain all queued commands.
	drain:
		for {
			select {
			case cmd := <-cmdCh:
				applyCmd(cmd, &paused, &frame, &prevNano)
			default:
				break drain
			}
		}

		if paused {
			// Block until a command arrives or context is cancelled.
			select {
			case <-ctx.Done():
				return
			case cmd := <-cmdCh:
				applyCmd(cmd, &paused, &frame, &prevNano)
			}
			continue
		}

		ts, raw, err := readFrame(frame)
		if err != nil {
			return
		}

		// Honour realtime pacing; allow interruption during the sleep.
		needReread := false
		if realtime && prevNano != 0 {
			delta := time.Duration(ts - prevNano)
			if delta > 0 && delta < 200*time.Millisecond {
				select {
				case <-ctx.Done():
					return
				case cmd := <-cmdCh:
					applyCmd(cmd, &paused, &frame, &prevNano)
					needReread = true
				case <-time.After(delta):
				}
			}
		}
		if needReread {
			continue
		}

		prevNano = ts
		if p, err := parsePacket(raw); err == nil {
			onPacket(p)
		}
		onProgress(frame, total)
		frame++
	}
}
