package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGlob(t *testing.T) {
	for path, expected := range map[string][]string{
		"fixtures":      {"fixtures"},
		"fixtures/":     {"fixtures/"},
		"fixtures/*":    {"fixtures/cron"},
		"fixtures/cron": {"fixtures/cron"},
	} {
		actual, err := filepath.Glob(path)
		assert.NoError(t, err)
		assert.EqualValues(t, expected, actual)
	}
}

func TestCronRead(t *testing.T) {
	f, err := os.Open("fixtures/cron")
	assert.NoError(t, err)

	c := new(Cron)
	assert.NoError(t, c.Read(f))
}

func TestCronReadLine(t *testing.T) {
	now := time.Date(2016, time.February, 1, 0, 0, 0, 0, time.UTC)
	t.Log(now.Format(time.Kitchen))
	c := new(Cron)
	for cron, nextRun := range map[string]time.Time{
		"* * * * * COMMAND":       now.Add(1 * time.Minute),
		"*/10 * * * * COMMAND":    now.Add(10 * time.Minute),
		"* */2 * * * COMMAND":     now.Add(1 * time.Hour),
		"0 * * * * COMMAND":       now.Add(1 * time.Hour),
		"0 */2 * * * COMMAND ARG": now.Add(1 * time.Hour),
		"@daily COMMAND ARG":      now.Add(23 * time.Hour),
		"@hourly COMMAND":         now.Add(1 * time.Hour),
	} {
		id, err := c.ReadLine(cron)
		assert.NoError(t, err)
		assert.Equal(t, nextRun, c.Entry(id).Schedule.Next(now))
		t.Log(cron, nextRun.Format(time.Kitchen), c.Entry(id).Schedule.Next(now).Format(time.Kitchen))
	}
}
