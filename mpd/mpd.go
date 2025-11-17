package mpd

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Client struct {
	conn   net.Conn
	reader *bufio.Reader
}

var (
	CurrentStatus *Status
	statusMutex   sync.RWMutex
)

type Status struct {
	State          string // e.g., "play", "pause", "stop"
	Volume         int    // 0-100 or -1 if unavailable
	Repeat         bool   // Repeat mode
	Random         bool   // Random mode
	Single         bool   // Single mode
	Consume        bool   // Consume mode
	PlaylistLength int
	Song           int
	SongID         int // Current song ID
	NextSong       int
	NextSongID     int
	Duration       int
	Elapsed        float64 // Elapsed time of current song
	Bitrate        int     // kbit/s
	Error          string  // If an error occurred
	Artist         string
	Album          string
	Title          string
}

func NewClient(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("could not connect to MPD at %s: %w", addr, err)
	}

	reader := bufio.NewReader(conn)

	// Read the initial "OK MPD" line
	line, err := reader.ReadString('\n')
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to read MPD welcome message: %w", err)
	}

	if !strings.HasPrefix(line, "OK MPD") {
		conn.Close()
		return nil, fmt.Errorf("unexpected MPD welcome message: %s", line)
	}

	return &Client{
		conn:   conn,
		reader: reader,
	}, nil
}

// Close disconnects from the MPD server.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// sendCommand sends a command to MPD and returns the response lines.
func (c *Client) sendCommand(command string) ([]string, error) {
	// Send the command with a newline
	_, err := fmt.Fprintln(c.conn, command)
	if err != nil {
		return nil, fmt.Errorf("failed to send command '%s': %w", command, err)
	}

	var response []string
	for {
		line, err := c.reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read response for '%s': %w", command, err)
		}

		line = strings.TrimSpace(line)

		// Check for end of response
		if line == "OK" {
			break
		}

		// Check for an error response
		if strings.HasPrefix(line, "ACK") {
			return nil, fmt.Errorf("mpd command '%s' failed: %s", command, line)
		}

		response = append(response, line)
	}

	return response, nil
}

// parseKVP parses a list of "key: value" strings into a map.
func parseKVP(lines []string) map[string]string {
	m := make(map[string]string)
	for _, line := range lines {
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) == 2 {
			m[parts[0]] = parts[1]
		}
	}
	return m
}

// List sends a `list` command to MPD.
// It returns a list of values for the given tag.
// For example, `List("artist")` returns all artists.
// `List("album", "artist", "Daft Punk")` returns albums by Daft Punk.
func (c *Client) List(tag string, args ...string) ([]string, error) {
	cmd := fmt.Sprintf("list %s", tag)
	for i := 0; i < len(args); i += 2 {
		cmd += fmt.Sprintf(" %s \"%s\"", args[i], args[i+1])
	}

	lines, err := c.sendCommand(cmd)
	if err != nil {
		return nil, err
	}

	// The response is in "key: value" format, we just want the values.
	values := make([]string, 0, len(lines))
	prefix := tag + ": "
	for _, line := range lines {
		if strings.HasPrefix(strings.ToLower(line), prefix) {
			values = append(values, line[len(prefix):])
		}
	}
	return values, nil
}

// Status fetches the current status from MPD and populates a Status struct.
func (c *Client) Status() (*Status, error) {
	lines, err := c.sendCommand("status")
	if err != nil {
		return nil, err
	}

	kv := parseKVP(lines)
	s := &Status{Volume: -1, SongID: -1, NextSongID: -1} // Defaults

	if state, ok := kv["state"]; ok {
		s.State = state
	}
	if volumeStr, ok := kv["volume"]; ok {
		s.Volume, _ = strconv.Atoi(volumeStr)
	}
	if repeatStr, ok := kv["repeat"]; ok {
		s.Repeat = (repeatStr == "1")
	}
	if randomStr, ok := kv["random"]; ok {
		s.Random = (randomStr == "1")
	}
	if singleStr, ok := kv["single"]; ok {
		s.Single = (singleStr == "1")
	}
	if consumeStr, ok := kv["consume"]; ok {
		s.Consume = (consumeStr == "1")
	}
	if playlistLengthStr, ok := kv["playlistlength"]; ok {
		s.PlaylistLength, _ = strconv.Atoi(playlistLengthStr)
	}
	if songStr, ok := kv["song"]; ok {
		s.Song, _ = strconv.Atoi(songStr)
	}
	if songIDStr, ok := kv["songid"]; ok {
		s.SongID, _ = strconv.Atoi(songIDStr)
	}
	if nextSongStr, ok := kv["nextsong"]; ok {
		s.NextSong, _ = strconv.Atoi(nextSongStr)
	}
	if nextSongIDStr, ok := kv["nextsongid"]; ok {
		s.NextSongID, _ = strconv.Atoi(nextSongIDStr)
	}
	if durationStr, ok := kv["duration"]; ok {
		durationFloat, _ := strconv.ParseFloat(durationStr, 64)
		s.Duration = int(durationFloat)
	}
	if elapsedStr, ok := kv["elapsed"]; ok {
		s.Elapsed, _ = strconv.ParseFloat(elapsedStr, 64)
	}
	if bitrateStr, ok := kv["bitrate"]; ok {
		s.Bitrate, _ = strconv.Atoi(bitrateStr)
	}
	if errorStr, ok := kv["error"]; ok {
		s.Error = errorStr
	}

	// If a song is playing or paused, get its details
	if s.State == "play" || s.State == "pause" {
		currentSongLines, err := c.sendCommand("currentsong")
		if err != nil {
			// Log the error but don't fail the whole status update
			log.Printf("mpd: could not get current song: %v", err)
		} else {
			songKV := parseKVP(currentSongLines)
			s.Artist = songKV["Artist"]
			s.Album = songKV["Album"]
			s.Title = songKV["Title"]
		}
	}

	return s, nil
}

// Play starts playback.
func (c *Client) Play(song int) error {
	cmd := "play"
	if song >= 0 {
		cmd = fmt.Sprintf("play %d", song)
	}
	_, err := c.sendCommand(cmd)
	return err
}

// PlayID plays the song with the given ID in the playlist.
func (c *Client) PlayID(songID int) error {
	cmd := "playid"
	if songID >= 0 {
		cmd = fmt.Sprintf("playid %d", songID)
	}
	_, err := c.sendCommand(cmd)
	return err
}

// Pause toggles the pause state.
// Pass true to pause, false to unpause.
func (c *Client) Pause(p bool) error {
	pauseState := 0
	if p {
		pauseState = 1
	}
	cmd := fmt.Sprintf("pause %d", pauseState)
	_, err := c.sendCommand(cmd)
	return err
}

// Random enables or disables random mode.
func (c *Client) Random(r bool) error {
	randomState := 0
	if r {
		randomState = 1
	}
	cmd := fmt.Sprintf("random %d", randomState)
	_, err := c.sendCommand(cmd)
	return err
}

// Repeat enables or disables repeat mode.
func (c *Client) Repeat(r bool) error {
	repeatState := 0
	if r {
		repeatState = 1
	}
	cmd := fmt.Sprintf("repeat %d", repeatState)
	_, err := c.sendCommand(cmd)
	return err
}

// Single enables or disables single mode.
func (c *Client) Single(s bool) error {
	singleState := 0
	if s {
		singleState = 1
	}
	cmd := fmt.Sprintf("single %d", singleState)
	_, err := c.sendCommand(cmd)
	return err
}

// Next plays the next song in the playlist.
func (c *Client) Next() error {
	_, err := c.sendCommand("next")
	return err
}

// Previous plays the previous song in the playlist.
func (c *Client) Previous() error {
	_, err := c.sendCommand("previous")
	return err
}

// WatchStatus connects to the MPD server at the given address and periodically
// updates the public CurrentStatus variable. It handles reconnecting if the
// connection is lost. This function is designed to be run in a goroutine.
func WatchStatus(addr string, interval time.Duration) {
	for {
		client, err := NewClient(addr)
		if err != nil {
			log.Printf("mpd: failed to connect to %s: %v. Retrying in %s...", addr, err, interval)
			time.Sleep(interval)
			continue
		}

		log.Printf("mpd: connected to %s", addr)

		ticker := time.NewTicker(interval)
		for range ticker.C {
			status, err := client.Status()
			if err != nil {
				log.Printf("mpd: failed to get status: %v. Reconnecting...", err)
				client.Close()
				ticker.Stop()
				break // Break inner loop to reconnect
			}

			statusMutex.Lock()
			CurrentStatus = status
			statusMutex.Unlock()
		}

		// If the loop was broken, it means there was an error.
		// The outer loop will handle reconnection after a delay.
		// No need for an extra sleep here as the outer loop's `continue`
		// will be followed by a sleep if the next connection attempt fails.
	}
}
