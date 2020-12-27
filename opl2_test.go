package opl2_test

import (
	"flag"
	"math"
	"os"
	"testing"
	"time"

	"github.com/gotracker/opl2"
	"github.com/pkg/errors"
)

var (
	ym3812 *opl2.Chip
	ymf262 *opl2.Chip

	sampleRate uint

	errIrqFailure          = errors.New("irq failure")
	errOPL2ChipNotDetected = errors.New("opl2 chip not detected")
	errOPL3ChipNotDetected = errors.New("opl3 chip not detected")
)

func detectChip(c *opl2.Chip) error {
	// Reset both timers
	c.WriteReg(0x04, 0x60)
	// Enable the interrupts
	c.WriteReg(0x04, 0x80)
	// Read status value
	status1 := c.ReadStatus()
	testStatus1 := status1 & 0xE0
	if testStatus1 != 0x00 {
		return errors.Wrapf(errIrqFailure, "expected status of 00 was not found, got %0.2X instead", testStatus1)
	}

	// Write FF to Timer 1
	c.WriteReg(0x02, 0xFF)
	// Start Timer 1
	c.WriteReg(0x04, 0x21)
	// Delay for at least 80 microseconds, simulated by generating 80us of data
	dataLen := math.Round((time.Microsecond * 80).Seconds() * float64(sampleRate))
	c.GenerateBlock2(uint(dataLen), nil)
	// Read status value
	status2 := c.ReadStatus()
	testStatus2 := status2 & 0xE0
	if testStatus2 != 0xC0 {
		return errors.Wrapf(errIrqFailure, "expected status of C0 was not found, got %0.2X instead", testStatus2)
	}

	// Reset both timers
	c.WriteReg(0x04, 0x60)
	// Enable the interrupts
	c.WriteReg(0x04, 0x80)

	return nil
}

func detectOPLVersion(c *opl2.Chip, opl3Expected bool) error {
	if err := detectChip(c); err != nil {
		return err
	}

	status3 := c.ReadStatus()
	testStatus3 := status3 & 0x06
	if opl3Expected {
		if testStatus3 != 0x00 {
			return errors.Wrapf(errOPL3ChipNotDetected, "expected status of 00 was not found, got %0.2X instead", testStatus3)
		}
	} else {
		if testStatus3 == 0x00 {
			return errors.Wrapf(errOPL2ChipNotDetected, "expected status of non-00 was not found, got %0.2X instead", testStatus3)
		}
	}

	return nil
}

func TestDetectOPL3WithOPL2(t *testing.T) {
	if err := detectOPLVersion(ym3812, true); err != nil {
		if !errors.Is(err, errOPL3ChipNotDetected) {
			t.Error(err)
		}
	}
}

func TestDetectOPL2WithOPL3(t *testing.T) {
	if err := detectOPLVersion(ymf262, false); err != nil {
		if !errors.Is(err, errOPL2ChipNotDetected) {
			t.Error(err)
		}
	}
}

func TestResetOPL2(t *testing.T) {
	if err := detectChip(ym3812); err != nil {
		t.Error(err)
	}
}

func TestResetOPL3(t *testing.T) {
	if err := detectChip(ymf262); err != nil {
		t.Error(err)
	}
}

func TestMain(m *testing.M) {
	flag.UintVar(&sampleRate, "s", uint(math.Round(opl2.OPLRATE)), "sample rate for OPL2/3 devices")
	flag.Parse()

	ym3812 = opl2.NewChip(uint32(sampleRate), false)
	ymf262 = opl2.NewChip(uint32(sampleRate), true)

	os.Exit(m.Run())
}
