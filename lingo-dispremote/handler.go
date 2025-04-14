package dispremote

import (
	"errors"
	"time"

	ws "github.com/leo82309/ipod/ws"
	"github.com/oandrew/ipod"
)

type DeviceDispRemote interface {
}

func ackSuccess(req *ipod.Command) *ACK {
	return &ACK{Status: ACKStatusSuccess, CmdID: uint8(req.ID.CmdID())}
}

func HandleDispRemote(req *ipod.Command, tr ipod.CommandWriter, dev DeviceDispRemote) error {
	switch msg := req.Payload.(type) {

	case *GetCurrentEQProfileIndex:
		ipod.Respond(req, tr, &RetCurrentEQProfileIndex{
			CurrentEQIndex: 0,
		})

	case *SetCurrentEQProfileIndex:
		ipod.Respond(req, tr, ackSuccess(req))

	case *GetNumEQProfiles:
		ipod.Respond(req, tr, &RetNumEQProfiles{
			NumEQProfiles: 1,
		})
	case *GetIndexedEQProfileName:
		ipod.Respond(req, tr, &RetIndexedEQProfileName{
			EQProfileName: ipod.StringToBytes("Default"),
		})
	case *SetRemoteEventNotification:
		ipod.Respond(req, tr, ackSuccess(req))

	case *GetRemoteEventStatus:
		ipod.Respond(req, tr, &RetRemoteEventStatus{
			EventStatus: 0,
		})

	case *GetiPodStateInfo:
		t := &RetiPodStateInfo{
			InfoType: msg.InfoType,
		}

		switch msg.InfoType {
		case InfoTypeTrackPositionMs:
			t.InfoData = &InfoTrackPositionMs{TrackPositionMs: uint32(ws.WSState.ElapsedTime * 1000)}
		case InfoTypeTrackIndex:
			t.InfoData = &InfoTrackIndex{TrackIndex: uint32(ws.WSState.SongPos)}
		case InfoTypeChapterIndex:
			t.InfoData = &InfoChapterIndex{
				TrackIndex:   uint32(ws.WSState.SongPos),
				ChapterCount: 0,
				ChapterIndex: 0,
			}
		case InfoTypePlayStatus:
			t.InfoData = &InfoPlayStatus{
				PlayStatus: PlayStatusPlaying,
			}
		case InfoTypeVolume:
			t.InfoData = &InfoVolume{MuteState: 0x00, UIVolumeLevel: 255}
		case InfoTypePower:
			t.InfoData = &InfoPower{PowerState: 0x05, BatteryLevel: 255}
		case InfoTypeEqualizer:
			t.InfoData = &InfoEqualizer{0x00}
		case InfoTypeShuffle:
			t.InfoData = &InfoShuffle{0x00}
		case InfoTypeRepeat:
			t.InfoData = &InfoRepeat{0x00}
		case InfoTypeDateTime:
			d := time.Now()
			t.InfoData = &InfoDateTime{
				Year:   uint16(d.Year()),
				Month:  uint8(d.Month()),
				Day:    uint8(d.Day()),
				Hour:   uint8(d.Hour()),
				Minute: uint8(d.Minute()),
			}
		case InfoTypeBacklight:
			t.InfoData = &InfoBacklight{BacklightLevel: 255}
		case InfoTypeHoldSwitch:
			t.InfoData = &InfoHoldSwitch{HoldSwitchState: 0x00}
		case InfoTypeSoundCheck:
			t.InfoData = &InfoSoundCheck{SoundCheckState: 0x00}
		case InfoTypeAudiobookSpeed:
			t.InfoData = &InfoAudiobookSpeed{0x00}
		case InfoTypeTrackPositionSec:
			t.InfoData = &InfoTrackPositionSec{uint16(ws.WSState.ElapsedTime)}
		case InfoTypeVolume2:
			t.InfoData = &InfoVolume2{
				MuteState:           0x00,
				UIVolumeLevel:       255,
				AbsoluteVolumeLevel: 255,
			}
		default:
			return errors.New("unknown info type")
		}

		ipod.Respond(req, tr, t)

	case *SetiPodStateInfo:
		ipod.Respond(req, tr, ackSuccess(req))

	case *GetPlayStatus:
		ipod.Respond(req, tr, &RetPlayStatus{
			PlayState:   1, //playing
			TrackIndex:  uint32(ws.WSState.SongPos),
			TrackLength: uint32(ws.WSState.TotalTime) * 1000,
			TrackPos:    uint32(ws.WSState.SongPos),
		})

	case *SetCurrentPlayingTrack:
		ipod.Respond(req, tr, ackSuccess(req))
	case *GetIndexedPlayingTrackInfo:
		t := &RetIndexedPlayingTrackInfo{
			InfoType: msg.InfoType,
		}

		switch msg.InfoType {
		case TrackInfoTypeCaps:
			t.InfoData = &TrackInfoCaps{
				Caps:         0x00,
				TrackTotalMs: 300000,
				ChapterCount: 0,
			}
		case TrackInfoTypeChapterTimeName:
			t.InfoData = &TrackInfoChapterTimeName{
				ChapterTime: 0,
				ChapterName: ipod.StringToBytes(""),
			}
		case TrackInfoTypeArtist:
			t.InfoData = &TrackInfoArtist{
				Name: ipod.StringToBytes(ws.WSSongInfo.Artist),
			}
		case TrackInfoTypeAlbum:
			t.InfoData = &TrackInfoAlbum{
				Name: ipod.StringToBytes(ws.WSSongInfo.Album),
			}
		case TrackInfoTypeGenre:
			t.InfoData = &TrackInfoGenre{
				Name: ipod.StringToBytes(""),
			}
		case TrackInfoTypeTrack:
			t.InfoData = &TrackInfoTrack{
				Title: ipod.StringToBytes(ws.WSSongInfo.Title),
			}
		case TrackInfoTypeComposer:
			t.InfoData = &TrackInfoComposer{
				Name: ipod.StringToBytes(""),
			}
		case TrackInfoTypeLyrics:
			t.InfoData = &TrackInfoLyrics{
				Flags:       0x00,
				PacketIndex: 0,
				Lyrics:      ipod.StringToBytes(""),
			}
		case TrackInfoTypeArtworkCount:
			t.InfoData = &TrackInfoArtworkCount{
				None: 0x08,
			}
		default:
			return errors.New("unknown info type")
		}

		ipod.Respond(req, tr, t)
	case *GetNumPlayingTracks:
		ipod.Respond(req, tr, &RetNumPlayingTracks{
			NumPlayTracks: uint32(len(ws.WSQueue)),
		})
	case *GetArtworkFormats:
		ipod.Respond(req, tr, &RetArtworkFormats{})
	case *GetTrackArtworkData:
	// RetTrackArtworkData:
	//todo
	case *GetPowerBatteryState:
		ipod.Respond(req, tr, &RetPowerBatteryState{
			BatteryLevel: 255, // 100%
			PowerState:   0x01,
		})
	case *GetSoundCheckState:
		ipod.Respond(req, tr, &RetSoundCheckState{
			Enabled: false,
		})
	case *SetSoundCheckState:
		ipod.Respond(req, tr, ackSuccess(req))
	case *GetTrackArtworkTimes:
		ipod.Respond(req, tr, &RetTrackArtworkTimes{})

	default:
		_ = msg
	}
	return nil
}
