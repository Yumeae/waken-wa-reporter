package media

import (
	"testing"
)

func TestInfo_IsEmpty(t *testing.T) {
	if !(Info{}).IsEmpty() {
		t.Fatal("zero Info must be empty")
	}
	if (Info{Title: "x"}).IsEmpty() {
		t.Fatal("non-empty title")
	}
}

func TestInfo_Signature(t *testing.T) {
	if g, w := (Info{}).Signature(), ""; g != w {
		t.Fatalf("empty sig: got %q want %q", g, w)
	}
	i := Info{Title: "T", Artist: "A", Album: "L", SourceAppID: "spotify"}
	if i.Signature() == "" {
		t.Fatal("expected non-empty signature")
	}
	i2 := Info{Title: "T", Artist: "A", Album: "L", SourceAppID: "spotify"}
	if i.Signature() != i2.Signature() {
		t.Fatal("signature must be stable")
	}
	i3 := Info{Title: "T", Artist: "A", Album: "L", SourceAppID: "chrome"}
	if i.Signature() == i3.Signature() {
		t.Fatal("signature must change when source app changes")
	}
}

func TestInfo_AsMap(t *testing.T) {
	if (Info{}).AsMap() != nil {
		t.Fatalf("empty AsMap: %#v", (Info{}).AsMap())
	}
	m := Info{Title: "T1", Artist: "A1", Album: "L1"}.AsMap()
	if m["title"] != "T1" || m["artist"] != "A1" || m["album"] != "L1" || m["singer"] != "A1" {
		t.Fatalf("AsMap: %#v", m)
	}
}
