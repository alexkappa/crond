package main

import (
	"fmt"
	"testing"
)

func TestAnnotation(t *testing.T) {
	for _, id := range []string{
		"080063bc-9a07-460b-b825-9bb1f4e0a578",
		"80cd5101-07ed-47e1-ba61-029b44ec035c",
		"ccaee346-4a37-45b7-8079-2237bffca9f3",
		"29779685-f8ba-4c48-b124-de09b993d76c",
	} {
		cmd := fmt.Sprintf("command a b # hc:%s", id)

		hc := newHealthCheck()

		if !hc.HasAnnotation(cmd) {
			t.Fatalf("HasAnnotation should return true for command %q", cmd)
		}
		if hc.FromAnnotation(cmd) != id {
			t.Errorf("FromAnnotation should return %q for command %q", id, cmd)
		}
	}
}
