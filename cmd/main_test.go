package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunKeepsCLIFlowOperational(t *testing.T) {
	input := strings.NewReader("105,5\n67,2\n184,72\n")
	var output bytes.Buffer

	if err := run(input, &output); err != nil {
		t.Fatal(err)
	}

	for _, expected := range []string{
		"Rateio Luz",
		"Consumo total: 172,7 kWh",
		"Proporção consumidor 1: 1055/1727 (61,09%)",
		"Proporção consumidor 2: 672/1727 (38,91%)",
		"Consumidor 1 paga: R$ 112,84",
		"Consumidor 2 paga: R$ 71,88",
	} {
		if !strings.Contains(output.String(), expected) {
			t.Errorf("output does not contain %q:\n%s", expected, output.String())
		}
	}
}

func TestRunReturnsValidationError(t *testing.T) {
	input := strings.NewReader("abc\n10\n50\n")
	var output bytes.Buffer

	err := run(input, &output)
	if err == nil || !strings.Contains(err.Error(), "consumo do consumidor 1") {
		t.Fatalf("run() error = %v", err)
	}
}
