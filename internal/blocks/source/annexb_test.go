package source

import (
	"bytes"
	"testing"
)

func TestFirstParameterSets_PicksFirstSPSAndPPSAcrossNALUs(t *testing.T) {
	nalus := [][]byte{
		{0x09, 0x00},                   // AUD (type 9) — non-PS
		{0x67, 0x42, 0xC0, 0x1E},       // SPS (type 7)
		{0x68, 0xCE, 0x38, 0x80},       // PPS (type 8)
		{0x65, 0xB8, 0x00, 0x00, 0x00}, // IDR (type 5) — VCL after, ignored
	}
	sps, pps := firstParameterSets(nalus)
	if !bytes.Equal(sps, nalus[1]) {
		t.Fatalf("sps = %x, want %x", sps, nalus[1])
	}
	if !bytes.Equal(pps, nalus[2]) {
		t.Fatalf("pps = %x, want %x", pps, nalus[2])
	}
}

func TestFirstParameterSets_ReturnsNilForMissingSets(t *testing.T) {
	nalus := [][]byte{{0x09}, {0x65, 0xB8}}
	sps, pps := firstParameterSets(nalus)
	if sps != nil || pps != nil {
		t.Fatalf("expected nil, got sps=%x pps=%x", sps, pps)
	}
}
