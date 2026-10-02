package collector

import (
	"bufio"
	"context"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ITlusions/ITL.Amalia.EdgeSensor/internal/event"
)

const initialSyslogWindow = int64(64 * 1024)

type SyslogCollector struct {
	DeviceID string
	Path     string

	mu     sync.Mutex
	offset int64
}

type SyslogMessage struct {
	Message string `json:"message"`
}

func (c *SyslogCollector) Name() string { return "syslog" }

func (c *SyslogCollector) Collect(_ context.Context) ([]event.Event, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	file, err := os.Open(c.Path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if info.Size() < c.offset {
		c.offset = 0
	}
	skipPartialLine := false
	if c.offset == 0 && info.Size() > initialSyslogWindow {
		c.offset = info.Size() - initialSyslogWindow
		skipPartialLine = true
	}
	if _, err := file.Seek(c.offset, 0); err != nil {
		return nil, err
	}

	var events []event.Event
	scanner := bufio.NewScanner(file)
	if skipPartialLine {
		scanner.Scan()
	}
	for scanner.Scan() {
		message := strings.TrimSpace(scanner.Text())
		if message == "" {
			continue
		}
		e, err := event.New("edge.syslog", c.DeviceID, "info", time.Now(), SyslogMessage{Message: message})
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	position, err := file.Seek(0, 1)
	if err == nil {
		c.offset = position
	}
	return events, scanner.Err()
}
