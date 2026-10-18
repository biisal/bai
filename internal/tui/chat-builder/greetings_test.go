package chatbuilder

import (
	"testing"
	"time"
)

func at(day time.Weekday, hour int) time.Time {
	// 2024-01-01 was a Monday
	return time.Date(2024, 1, 1+int(day-time.Monday+7)%7, hour, 0, 0, 0, time.UTC)
}

func TestGreetingFor(t *testing.T) {
	cases := []struct {
		day  time.Weekday
		hour int
		want string
	}{
		{time.Monday, 9, "Happy Monday"}, // day-specific wins over the generic morning ones
		{time.Friday, 9, "Happy Friday"},
		{time.Saturday, 10, "Welcome to the weekend"},
		{time.Wednesday, 14, "Good afternoon"},
		{time.Wednesday, 19, "Good evening"},
		{time.Wednesday, 2, "How’s it going?"}, // 21-6 wraps past midnight
		{time.Wednesday, 23, "How’s it going?"},
	}
	for _, c := range cases {
		g := greetingFor(at(c.day, c.hour))
		if g == nil {
			t.Fatalf("%v %dh: no greeting", c.day, c.hour)
		}
		if g.text != c.want {
			t.Errorf("%v %dh: got %q want %q", c.day, c.hour, g.text, c.want)
		}
	}
}

func TestRenderPrefersAsciiWhenItFits(t *testing.T) {
	g := greeting{text: "plain", ascii: "██\n██"}
	if got := g.render(2); got != g.ascii {
		t.Errorf("fits: got %q", got)
	}
	if got := g.render(1); got != "plain" {
		t.Errorf("too narrow: got %q", got)
	}
	if got := (greeting{text: "plain"}).render(80); got != "plain" {
		t.Errorf("empty ascii: got %q", got)
	}
}

func TestInRange(t *testing.T) {
	if !inRange(2, timeRange{21, 6}) || !inRange(21, timeRange{21, 6}) || inRange(6, timeRange{21, 6}) {
		t.Error("wrap-around range 21-6 is wrong")
	}
	if !inRange(6, timeRange{6, 12}) || inRange(12, timeRange{6, 12}) {
		t.Error("plain range 6-12 is wrong")
	}
}
