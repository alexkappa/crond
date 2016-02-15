package main

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCronRead(t *testing.T) {
	f, err := os.Open("fixtures/cron")
	assert.NoError(t, err)

	c := new(Cron)
	assert.NoError(t, c.Read(f))
}

func TestCronReadLine(t *testing.T) {
	c := new(Cron)
	for cron, expected := range map[string]time.Duration{
		"* * * * * COMMAND":       1 * time.Minute,
		"*/10 * * * * COMMAND":    10 * time.Minute,
		"0 * * * * COMMAND":       1 * time.Hour,
		"0 */2 * * * COMMAND ARG": 2 * time.Hour,
		"0 0 */2 * * COMMAND":     2 * 24 * time.Hour,
		"@daily COMMAND ARG":      24 * time.Hour,
		"@hourly COMMAND":         1 * time.Hour,
		"@every 1h COMMAND":       1 * time.Hour,
		"@every 30s COMMAND":      30 * time.Second,
	} {
		id, err := c.ReadLine(cron)
		assert.NoError(t, err)
		run1 := c.Entry(id).Schedule.Next(time.Now())
		run2 := c.Entry(id).Schedule.Next(run1)
		assert.Equal(t, expected, run2.Sub(run1), cron)
	}
}

func TestCronReadIgnoredLine(t *testing.T) {
	c := new(Cron)
	for _, cron := range []string{
		"",
		"    ",
		"#",
		"# comment",
	} {
		_, err := c.ReadLine(cron)
		assert.NoError(t, err)
	}
}
