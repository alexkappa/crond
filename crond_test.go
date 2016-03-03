package main

import (
	"os"
	"testing"
	"time"
)

func TestCronRead(t *testing.T) {
	f, err := os.Open("fixtures/cron")
	if err != nil {
		t.Fatal(err)
	}
	c := new(Crond)
	if err = c.Read(f); err != nil {
		t.Fatal(err)
	}
}

func TestCronReadLine(t *testing.T) {
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
		c := new(Crond)
		id, err := c.ReadLine(cron)
		if err != nil {
			t.Fatal(err)
		}

		run1 := c.Entry(id).Schedule.Next(time.Now())
		run2 := c.Entry(id).Schedule.Next(run1)

		actual := run2.Sub(run1)

		if expected != actual {
			t.Errorf("Failed asserting that %v equals %v", expected, actual)
		}
	}
}

func TestCronReadIgnoredLine(t *testing.T) {
	c := new(Crond)
	for _, cron := range []string{
		"",
		"    ",
		"#",
		"# comment",
	} {
		_, err := c.ReadLine(cron)
		if err != nil {
			t.Error(err)
		}
	}
}
