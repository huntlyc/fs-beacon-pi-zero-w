package lcd

import (
	"time"

	"github.com/d2r2/go-i2c"
)

// DATASHEET: https://cdn-shop.adafruit.com/product-files/399/399+spec+sheet.pdf

// Constants needed to manipulate the display
// Bit 7 -> Backlight
// Bit 6 -> D7
// Bit 5 -> D6
// Bit 4 -> D5
// Bit 3 -> D4
// Bit 2 -> Enable
// Bit 1 -> RS
// Bit 0 -> unused
const (
	regIODIR = 0x00
	regGPIO  = 0x09

	RS = 1 << 1
	EN = 1 << 2

	D4 = 1 << 3
	D5 = 1 << 4
	D6 = 1 << 5
	D7 = 1 << 6

	BACKLIGHT = 1 << 7
)

type LCD struct {
	i2c  *i2c.I2C
	gpio byte
}

// Initialise the LCD.
func New(addr uint8, bus int) (*LCD, error) {
	dev, err := i2c.NewI2C(addr, bus)
	if err != nil {
		return nil, err
	}

	l := &LCD{
		i2c: dev,
	}

	// MCP23008 all outputs to 0
	if err := l.writeRegister(regIODIR, 0x00); err != nil {
		return nil, err
	}

	l.gpio = BACKLIGHT
	l.writeGPIO()

	time.Sleep(50 * time.Millisecond)

	// HD44780 init sequence
	l.write4Bits(0x03, false)
	time.Sleep(5 * time.Millisecond)

	l.write4Bits(0x03, false)
	time.Sleep(5 * time.Millisecond)

	l.write4Bits(0x03, false)
	time.Sleep(1 * time.Millisecond)

	l.write4Bits(0x02, false)

	l.command(0x28) // 4 bit, 2 line, 5x8 font
	l.command(0x0C) // display on
	l.command(0x06) // entry mode
	l.command(0x01) // clear

	time.Sleep(5 * time.Millisecond)

	return l, nil
}

// Close the i2c connection.
func (l *LCD) Close() {
	l.i2c.Close()
}

// Change the LCD backlight state to on or off.
// Uses:
// |= OR to turn on
//
//	&^= to turn off
//
// Leaves all other bits as is
func (l *LCD) Backlight(on bool) {
	if on {
		l.gpio |= BACKLIGHT
	} else {
		l.gpio &^= BACKLIGHT
	}

	l.writeGPIO()
}

// Converts s into a byte sequence and writes it to the LCD.
func (l *LCD) Print(s string) {
	for _, c := range []byte(s) {
		l.writeByte(c, true)
	}
}

// Sends clear screen command to lcd.
func (l *LCD) Clear() {
	l.command(0x01)
	time.Sleep(2 * time.Millisecond)
}

// Change the cursor position.
func (l *LCD) SetCursor(col, row int) {
	rowOffsets := []byte{0x00, 0x40}

	l.command(0x80 | (rowOffsets[row] + byte(col)))
}

func (l *LCD) command(v byte) {
	l.writeByte(v, false)
}

func (l *LCD) writeByte(v byte, rs bool) {
	// toggle register select
	if rs {
		l.gpio |= RS
	} else {
		l.gpio &^= RS
	}

	l.write4Bits(v>>4, rs)
	l.write4Bits(v&0x0F, rs)
}

func (l *LCD) WriteByte(b byte) {
	l.writeByte(b, true)
}

func (l *LCD) write4Bits(v byte, rs bool) {

	// turn off all data bits
	l.gpio &^= D4 | D5 | D6 | D7

	// individually turn on based on byte
	if v&0x01 != 0 {
		l.gpio |= D4
	}
	if v&0x02 != 0 {
		l.gpio |= D5
	}
	if v&0x04 != 0 {
		l.gpio |= D6
	}
	if v&0x08 != 0 {
		l.gpio |= D7
	}

	l.writeGPIO()
	l.pulseEnable()
}

// make the screen do our bidding
func (l *LCD) pulseEnable() {
	// turn on enable pin
	l.gpio |= EN
	l.writeGPIO()

	// wait
	time.Sleep(1 * time.Microsecond)

	// turn off enable pin
	l.gpio &^= EN
	l.writeGPIO()

	// wait
	time.Sleep(50 * time.Microsecond)
}

// lock in the data to send
func (l *LCD) writeGPIO() error {
	return l.writeRegister(regGPIO, l.gpio)
}

// send down the i2c wire
func (l *LCD) writeRegister(reg byte, value byte) error {
	_, err := l.i2c.WriteBytes([]byte{reg, value})
	return err
}
