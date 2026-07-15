package assets

import (
	"bytes"
	"image/png"
	"testing"
)

func TestEmbeddedIconIsA512PixelPNG(t *testing.T) {
	if Icon == nil || len(Icon.Content()) == 0 {
		t.Fatal("application icon is empty")
	}
	imageData, err := png.Decode(bytes.NewReader(Icon.Content()))
	if err != nil {
		t.Fatalf("decode embedded icon: %v", err)
	}
	if got := imageData.Bounds().Size(); got.X != 512 || got.Y != 512 {
		t.Errorf("icon size = %v, want 512x512", got)
	}
}
