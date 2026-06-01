package main

import (
	"bytes"
	"encoding/binary"
)

type ForzaHorizonPacket struct {
	IsRaceOn                    int32
	TimestampMS                 uint32
	EngineMaxRpm                float32
	EngineIdleRpm               float32
	CurrentEngineRpm            float32
	AccelerationX               float32
	AccelerationY               float32
	AccelerationZ               float32
	VelocityX                   float32
	VelocityY                   float32
	VelocityZ                   float32
	AngularVelocityX            float32
	AngularVelocityY            float32
	AngularVelocityZ            float32
	Yaw                         float32
	Pitch                       float32
	Roll                        float32
	SuspensionTravelFL          float32
	SuspensionTravelFR          float32
	SuspensionTravelRL          float32
	SuspensionTravelRR          float32
	TireSlipRatioFL             float32
	TireSlipRatioFR             float32
	TireSlipRatioRL             float32
	TireSlipRatioRR             float32
	WheelRotationSpeedFL        float32
	WheelRotationSpeedFR        float32
	WheelRotationSpeedRL        float32
	WheelRotationSpeedRR        float32
	WheelOnRumbleStripFL        int32
	WheelOnRumbleStripFR        int32
	WheelOnRumbleStripRL        int32
	WheelOnRumbleStripRR        int32
	WheelInPuddleFL             float32
	WheelInPuddleFR             float32
	WheelInPuddleRL             float32
	WheelInPuddleRR             float32
	SurfaceRumbleFL             float32
	SurfaceRumbleFR             float32
	SurfaceRumbleRL             float32
	SurfaceRumbleRR             float32
	TireSlipAngleFL             float32
	TireSlipAngleFR             float32
	TireSlipAngleRL             float32
	TireSlipAngleRR             float32
	TireCombinedSlipFL          float32
	TireCombinedSlipFR          float32
	TireCombinedSlipRL          float32
	TireCombinedSlipRR          float32
	SuspensionTravelMetersFL    float32
	SuspensionTravelMetersFR    float32
	SuspensionTravelMetersRL    float32
	SuspensionTravelMetersRR    float32
	CarOrdinal                  int32
	CarClass                    int32
	CarPerformanceIndex         int32
	DrivetrainType              int32
	NumCylinders                int32
	_                           [12]byte // Horizon-specific padding
	PositionX                   float32
	PositionY                   float32
	PositionZ                   float32
	Speed                       float32
	Power                       float32
	Torque                      float32
	TireTempFL                  float32
	TireTempFR                  float32
	TireTempRL                  float32
	TireTempRR                  float32
	Boost                       float32
	Fuel                        float32
	DistanceTraveled            float32
	BestLap                     float32
	LastLap                     float32
	CurrentLap                  float32
	CurrentRaceTime             float32
	LapNumber                   uint16
	RacePosition                uint8
	Accel                       uint8
	Brake                       uint8
	Clutch                      uint8
	HandBrake                   uint8
	Gear                        uint8
	Steer                       int8
	NormalizedDrivingLine       int8
	NormalizedAIBrakeDifference int8
	_                           [1]byte
}

const packetSize = 324

// TelemetrySnapshot is the JSON payload emitted to the frontend on every packet.
type TelemetrySnapshot struct {
	Speed          float32 `json:"speed"`
	RPM            float32 `json:"rpm"`
	MaxRPM         float32 `json:"maxRpm"`
	Gear           uint8   `json:"gear"`
	Throttle       float32 `json:"throttle"`
	Brake          float32 `json:"brake"`
	PositionX      float32 `json:"posX"`
	PositionY      float32 `json:"posY"`
	PositionZ      float32 `json:"posZ"`
	TireTempFL     float32 `json:"tireTempFL"`
	TireTempFR     float32 `json:"tireTempFR"`
	TireTempRL     float32 `json:"tireTempRL"`
	TireTempRR     float32 `json:"tireTempRR"`
	Boost          float32 `json:"boost"`
	Fuel           float32 `json:"fuel"`
	LapNumber      uint16  `json:"lapNumber"`
	RacePosition   uint8   `json:"racePosition"`
	CurrentLapTime float32 `json:"currentLapTime"`
	BestLap        float32 `json:"bestLap"`
	LastLap        float32 `json:"lastLap"`
	IsRaceOn       bool    `json:"isRaceOn"`
}

func parsePacket(raw []byte) (ForzaHorizonPacket, error) {
	var p ForzaHorizonPacket
	err := binary.Read(bytes.NewReader(raw), binary.LittleEndian, &p)
	return p, err
}

func packetToSnapshot(p ForzaHorizonPacket) TelemetrySnapshot {
	gear := p.Gear
	// FH6 sends gear=0 for both neutral and reverse.
	// When wheels spin backwards the average rotation speed is negative.
	if gear == 0 {
		avgWheel := (p.WheelRotationSpeedFL + p.WheelRotationSpeedFR +
			p.WheelRotationSpeedRL + p.WheelRotationSpeedRR) / 4
		if avgWheel < -0.5 {
			gear = 11 // treat as reverse
		}
	}
	return TelemetrySnapshot{
		Speed:          p.Speed * 3.6,
		RPM:            p.CurrentEngineRpm,
		MaxRPM:         p.EngineMaxRpm,
		Gear:           gear,
		Throttle:       float32(p.Accel) / 255.0 * 100,
		Brake:          float32(p.Brake) / 255.0 * 100,
		PositionX:      p.PositionX,
		PositionY:      p.PositionY,
		PositionZ:      p.PositionZ,
		TireTempFL:     (p.TireTempFL - 32) * 5 / 9,
		TireTempFR:     (p.TireTempFR - 32) * 5 / 9,
		TireTempRL:     (p.TireTempRL - 32) * 5 / 9,
		TireTempRR:     (p.TireTempRR - 32) * 5 / 9,
		Boost:          p.Boost,
		Fuel:           p.Fuel,
		LapNumber:      p.LapNumber,
		RacePosition:   p.RacePosition,
		CurrentLapTime: p.CurrentLap,
		BestLap:        p.BestLap,
		LastLap:        p.LastLap,
		IsRaceOn:       p.IsRaceOn != 0,
	}
}
