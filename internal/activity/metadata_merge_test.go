package activity

import (
	"reflect"
	"testing"
)

func TestMergeMetadata_shallow(t *testing.T) {
	dst := map[string]any{"source": "waken-wa", "a": 1}
	src := map[string]any{"b": "x"}
	MergeMetadata(dst, src)
	want := map[string]any{"source": "waken-wa", "a": 1, "b": "x"}
	if !reflect.DeepEqual(dst, want) {
		t.Fatalf("got %#v want %#v", dst, want)
	}
}

func TestMergeMetadata_mediaOneLevel(t *testing.T) {
	dst := map[string]any{
		"source": "waken-wa",
		"media":  map[string]any{"title": "T1"},
	}
	src := map[string]any{
		"media": map[string]any{"singer": "S1"},
	}
	MergeMetadata(dst, src)
	media, _ := dst["media"].(map[string]any)
	if media["title"] != "T1" || media["singer"] != "S1" {
		t.Fatalf("media got %#v", media)
	}
}

func TestMergeMetadata_mediaOverwriteScalar(t *testing.T) {
	dst := map[string]any{"media": "broken"}
	src := map[string]any{"media": map[string]any{"title": "T"}}
	MergeMetadata(dst, src)
	media, _ := dst["media"].(map[string]any)
	if media["title"] != "T" {
		t.Fatalf("got %#v", dst["media"])
	}
}
