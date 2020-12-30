package opl2

// Pure Go conversion of original C++ file

// This is the Opal OPL3 emulator from Reality Adlib Tracker v2.0a (http://www.3eality.com/productions/reality-adlib-tracker).
// It was released by Shayde/Reality into the public domain.
// Minor modifications to silence some warnings and fix a bug in the envelope generator have been applied.
// Additional fixes by JP Cimalando.

/*

   The Opal OPL3 emulator.

   Note: this is not a complete emulator, just enough for Reality Adlib Tracker tunes.

   Missing features compared to a real OPL3:

       - Timers/interrupts
       - OPL3 enable bit (it defaults to always on)
       - CSW mode
       - Test register
       - Percussion mode

*/

// Various constants
const (
	OPL3SampleRate = 49716

	NumChannels  = 18
	NumOperators = 36
)

type envStage int

const (
	envStageOff = envStage(-1)
	envStageAtt = envStage(0 + iota)
	envStageDec
	envStageSus
	envStageRel
)

var opalRateTables = [4][8]uint16{
	{1, 0, 1, 0, 1, 0, 1, 0},
	{1, 0, 1, 0, 0, 0, 1, 0},
	{1, 0, 0, 0, 1, 0, 0, 0},
	{1, 0, 0, 0, 0, 0, 0, 0},
}

var opalExpTable = [0x100]uint16{
	1018, 1013, 1007, 1002, 996, 991, 986, 980, 975, 969, 964, 959, 953, 948, 942, 937,
	932, 927, 921, 916, 911, 906, 900, 895, 890, 885, 880, 874, 869, 864, 859, 854,
	849, 844, 839, 834, 829, 824, 819, 814, 809, 804, 799, 794, 789, 784, 779, 774,
	770, 765, 760, 755, 750, 745, 741, 736, 731, 726, 722, 717, 712, 708, 703, 698,
	693, 689, 684, 680, 675, 670, 666, 661, 657, 652, 648, 643, 639, 634, 630, 625,
	621, 616, 612, 607, 603, 599, 594, 590, 585, 581, 577, 572, 568, 564, 560, 555,
	551, 547, 542, 538, 534, 530, 526, 521, 517, 513, 509, 505, 501, 496, 492, 488,
	484, 480, 476, 472, 468, 464, 460, 456, 452, 448, 444, 440, 436, 432, 428, 424,
	420, 416, 412, 409, 405, 401, 397, 393, 389, 385, 382, 378, 374, 370, 367, 363,
	359, 355, 352, 348, 344, 340, 337, 333, 329, 326, 322, 318, 315, 311, 308, 304,
	300, 297, 293, 290, 286, 283, 279, 276, 272, 268, 265, 262, 258, 255, 251, 248,
	244, 241, 237, 234, 231, 227, 224, 220, 217, 214, 210, 207, 204, 200, 197, 194,
	190, 187, 184, 181, 177, 174, 171, 168, 164, 161, 158, 155, 152, 148, 145, 142,
	139, 136, 133, 130, 126, 123, 120, 117, 114, 111, 108, 105, 102, 99, 96, 93,
	90, 87, 84, 81, 78, 75, 72, 69, 66, 63, 60, 57, 54, 51, 48, 45,
	42, 40, 37, 34, 31, 28, 25, 22, 20, 17, 14, 11, 8, 6, 3, 0,
}

var opalLogSinTable = [0x100]uint16{
	2137, 1731, 1543, 1419, 1326, 1252, 1190, 1137, 1091, 1050, 1013, 979, 949, 920, 894, 869,
	846, 825, 804, 785, 767, 749, 732, 717, 701, 687, 672, 659, 646, 633, 621, 609,
	598, 587, 576, 566, 556, 546, 536, 527, 518, 509, 501, 492, 484, 476, 468, 461,
	453, 446, 439, 432, 425, 418, 411, 405, 399, 392, 386, 380, 375, 369, 363, 358,
	352, 347, 341, 336, 331, 326, 321, 316, 311, 307, 302, 297, 293, 289, 284, 280,
	276, 271, 267, 263, 259, 255, 251, 248, 244, 240, 236, 233, 229, 226, 222, 219,
	215, 212, 209, 205, 202, 199, 196, 193, 190, 187, 184, 181, 178, 175, 172, 169,
	167, 164, 161, 159, 156, 153, 151, 148, 146, 143, 141, 138, 136, 134, 131, 129,
	127, 125, 122, 120, 118, 116, 114, 112, 110, 108, 106, 104, 102, 100, 98, 96,
	94, 92, 91, 89, 87, 85, 83, 82, 80, 78, 77, 75, 74, 72, 70, 69,
	67, 66, 64, 63, 62, 60, 59, 57, 56, 55, 53, 52, 51, 49, 48, 47,
	46, 45, 43, 42, 41, 40, 39, 38, 37, 36, 35, 34, 33, 32, 31, 30,
	29, 28, 27, 26, 25, 24, 23, 23, 22, 21, 20, 20, 19, 18, 17, 17,
	16, 15, 15, 14, 13, 13, 12, 12, 11, 10, 10, 9, 9, 8, 8, 7,
	7, 7, 6, 6, 5, 5, 5, 4, 4, 4, 3, 3, 3, 2, 2, 2,
	2, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0,
}

// A single FM operator
type operator struct {
	Master         *Opal    // Master object
	Chan           *channel // Owning channel
	Phase          uint32   // The current offset in the selected waveform
	Waveform       uint16   // The waveform id this operator is using
	FreqMultTimes2 uint16   // Frequency multiplier * 2
	EnvelopeStage  envStage // Which stage the envelope is at (see Env* enums above)
	EnvelopeLevel  int16    // 0 - $1FF, 0 being the loudest
	OutputLevel    uint16   // 0 - $FF
	AttackRate     uint16
	DecayRate      uint16
	SustainLevel   uint16
	ReleaseRate    uint16
	AttackShift    uint16
	AttackMask     uint16
	AttackAdd      uint16
	AttackTab      []uint16
	DecayShift     uint16
	DecayMask      uint16
	DecayAdd       uint16
	DecayTab       []uint16
	ReleaseShift   uint16
	ReleaseMask    uint16
	ReleaseAdd     uint16
	ReleaseTab     []uint16
	KeyScaleShift  uint16
	KeyScaleLevel  uint16
	Out            [2]int16
	KeyOn          bool
	KeyScaleRate   bool // Affects envelope rate scaling
	SustainMode    bool // Whether to sustain during the sustain phase, or release instead
	TremoloEnable  bool
	VibratoEnable  bool
}

// Init - Operator constructor.
func (o *operator) Init() {
	o.Master = nil
	o.Chan = nil
	o.Phase = 0
	o.Waveform = 0
	o.FreqMultTimes2 = 1
	o.EnvelopeStage = envStageOff
	o.EnvelopeLevel = 0x1FF
	o.AttackRate = 0
	o.DecayRate = 0
	o.SustainLevel = 0
	o.ReleaseRate = 0
	o.KeyScaleShift = 0
	o.KeyScaleLevel = 0
	o.Out[0] = 0
	o.Out[1] = 0
	o.KeyOn = false
	o.KeyScaleRate = false
	o.SustainMode = false
	o.TremoloEnable = false
	o.VibratoEnable = false
}

// Output - Produce output from operator.
func (o *operator) Output(_ uint16, phaseStep uint32, vibrato int16, mod int16, fbshift int16) int16 {

	// Advance wave phase
	if o.VibratoEnable {
		phaseStep += uint32(vibrato)
	}
	o.Phase += (phaseStep * uint32(o.FreqMultTimes2)) / 2

	leveltemp := uint16(0)
	if o.TremoloEnable {
		leveltemp = o.Master.TremoloLevel
	}
	level := (uint16(o.EnvelopeLevel) + o.OutputLevel + o.KeyScaleLevel + leveltemp) << 3

	switch o.EnvelopeStage {
	case envStageAtt: // Attack stage
		add := uint16((o.AttackAdd>>o.AttackTab[o.Master.Clock>>o.AttackShift&7])*^uint16(o.EnvelopeLevel)) >> 3
		if o.AttackRate == 0 {
			add = 0
		}
		if o.AttackMask != 0 && (o.Master.Clock&o.AttackMask) != 0 {
			add = 0
		}
		o.EnvelopeLevel += int16(add)
		if o.EnvelopeLevel <= 0 {
			o.EnvelopeLevel = 0
			o.EnvelopeStage = envStageDec
		}

	case envStageDec: // Decay stage
		add := uint16(o.DecayAdd >> o.DecayTab[o.Master.Clock>>o.DecayShift&7])
		if o.DecayRate == 0 {
			add = 0
		}
		if o.DecayMask != 0 && (o.Master.Clock&o.DecayMask) != 0 {
			add = 0
		}
		o.EnvelopeLevel += int16(add)
		if o.EnvelopeLevel >= int16(o.SustainLevel) {
			o.EnvelopeLevel = int16(o.SustainLevel)
			o.EnvelopeStage = envStageSus
		}

	case envStageSus: // Sustain stage
		if o.SustainMode {
			break
		}
		o.EnvelopeStage = envStageRel
		fallthrough

	case envStageRel: // Release stage
		add := o.ReleaseAdd >> o.ReleaseTab[o.Master.Clock>>o.ReleaseShift&7]
		if o.ReleaseRate == 0 {
			add = 0
		}
		if o.ReleaseMask != 0 && (o.Master.Clock&o.ReleaseMask) != 0 {
			add = 0
		}
		o.EnvelopeLevel += int16(add)
		if o.EnvelopeLevel >= 0x1FF {
			o.EnvelopeLevel = 0x1FF
			o.EnvelopeStage = envStageOff
			o.Out[0] = 0
			o.Out[1] = 0
			return 0
		}

	// Envelope, and therefore the operator, is not running
	default:
		o.Out[0] = 0
		o.Out[1] = 0
		return 0
	}

	// Feedback?  In that case we modulate by a blend of the last two samples
	if fbshift != 0 {
		mod += (o.Out[0] + o.Out[1]) >> fbshift
	}

	phase := uint16(uint32(o.Phase>>10) + uint32(mod))
	offset := phase & 0xFF
	var logsin uint16
	negate := false

	switch o.Waveform {
	case 0: // Standard sine wave
		if (phase & 0x100) != 0 {
			offset ^= 0xFF
		}
		logsin = opalLogSinTable[offset]
		negate = (phase & 0x200) != 0
		break

	case 1: // Half sine wave
		if (phase & 0x200) != 0 {
			offset = 0
		} else if (phase & 0x100) != 0 {
			offset ^= 0xFF
		}
		logsin = opalLogSinTable[offset]
		break

	case 2: // Positive sine wave
		if (phase & 0x100) != 0 {
			offset ^= 0xFF
		}
		logsin = opalLogSinTable[offset]
		break

	case 3: // Quarter positive sine wave
		if (phase & 0x100) != 0 {
			offset = 0
		}
		logsin = opalLogSinTable[offset]
		break

	case 4: // Double-speed sine wave
		if (phase & 0x200) != 0 {
			offset = 0
		} else {
			if (phase & 0x80) != 0 {
				offset ^= 0xFF
			}

			offset = (offset + offset) & 0xFF
			negate = (phase & 0x100) != 0
		}

		logsin = opalLogSinTable[offset]
		break

	case 5: // Double-speed positive sine wave
		if (phase & 0x200) != 0 {
			offset = 0
		} else {
			offset = (offset + offset) & 0xFF
			if (phase & 0x80) != 0 {
				offset ^= 0xFF
			}
		}

		logsin = opalLogSinTable[offset]
		break

	case 6: // Square wave
		logsin = 0
		negate = (phase & 0x200) != 0
		break

	case 7: // Exponentiation wave
		logsin = phase & 0x1FF
		if (phase & 0x200) != 0 {
			logsin ^= 0x1FF
			negate = true
		}
		logsin <<= 3
		break

	default: // unknown
		panic("unknown wave function!")
	}

	mix := uint16(logsin + level)
	if mix > 0x1FFF {
		mix = 0x1FFF
	}

	// From the OPLx decapsulated docs:
	// "When such a table is used for calculation of the exponential, the table is read at the
	// position given by the 8 LSB's of the input. The value + 1024 (the hidden bit) is then the
	// significand of the floating point output and the yet unused MSB's of the input are the
	// exponent of the floating point output."
	v := int16(opalExpTable[mix&0xFF]+1024) >> (mix >> 8)
	v += v
	if negate {
		v = ^v
	}

	// Keep last two results for feedback calculation
	o.Out[1] = o.Out[0]
	o.Out[0] = v

	return v
}

// SetKeyOn - Trigger operator.
func (o *operator) SetKeyOn(on bool) {
	// Already on/off?
	if o.KeyOn == on {
		return
	}

	o.KeyOn = on

	if on {
		// The highest attack rate is instant; it bypasses the attack phase
		if o.AttackRate == 15 {
			o.EnvelopeStage = envStageDec
			o.EnvelopeLevel = 0
		} else {
			o.EnvelopeStage = envStageAtt
		}

		o.Phase = 0
	} else {
		// Stopping current sound?
		if o.EnvelopeStage != envStageOff && o.EnvelopeStage != envStageRel {
			o.EnvelopeStage = envStageRel
		}
	}
}

// SetTremeloEnable - Enable amplitude vibrato.
func (o *operator) SetTremoloEnable(on bool) {
	o.TremoloEnable = on
}

// SetVibratoEnable - Enable frequency vibrato.
func (o *operator) SetVibratoEnable(on bool) {
	o.VibratoEnable = on
}

// SetSustainMode - Sets whether we release or sustain during the sustain phase of the envelope. 'true' is to
// sustain, otherwise release.
func (o *operator) SetSustainMode(on bool) {
	o.SustainMode = on
}

// SetEnvelopeScaling - Key scale rate. Sets how much the Key Scaling Number affects the envelope rates.
func (o *operator) SetEnvelopeScaling(on bool) {
	o.KeyScaleRate = on
	o.ComputeRates()
}

// Needs to be multiplied by two (and divided by two later when we use it) because the first
// entry is actually .5
var mulTimes2 = []uint16{
	1, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 20, 24, 24, 30, 30,
}

// SetFrequencyMultiplier - Multiplies the phase frequency.
func (o *operator) SetFrequencyMultiplier(scale uint16) {
	o.FreqMultTimes2 = mulTimes2[scale&15]
}

var kslShift = [4]uint8{8, 1, 2, 0}

// SetKeyScale - Attenuates output level towards higher pitch.
func (o *operator) SetKeyScale(scale uint16) {
	o.KeyScaleShift = uint16(kslShift[scale])
	o.ComputeKeyScaleLevel()
}

// SetOutputLevel - Sets the output level (volume) of the operator.
func (o *operator) SetOutputLevel(level uint16) {
	o.OutputLevel = level * 4
}

// SetAttackRate - Operator attack rate.
func (o *operator) SetAttackRate(rate uint16) {
	o.AttackRate = rate
	o.ComputeRates()
}

// SetDecayRate - Operator decay rate.
func (o *operator) SetDecayRate(rate uint16) {
	o.DecayRate = rate
	o.ComputeRates()
}

// SetSustainLevel - Operator sustain level.
func (o *operator) SetSustainLevel(level uint16) {
	o.SustainLevel = 31
	if level < 15 {
		o.SustainLevel = level
	}
	o.SustainLevel *= 16
}

// SetReleaseRate - Operator release rate.
func (o *operator) SetReleaseRate(rate uint16) {
	o.ReleaseRate = rate
	o.ComputeRates()
}

// SetWaveform - Assign the waveform this operator will use.
func (o *operator) SetWaveform(wave uint16) {
	o.Waveform = wave & 7
}

// ComputeRates - Compute actual rate from register rate.  From the Yamaha data sheet:
//
// Actual rate = Rate value * 4 + Rof, if Rate value = 0, actual rate = 0
//
// Rof is set as follows depending on the KSR setting:
//
//  Key scale   0   1   2   3   4   5   6   7   8   9   10  11  12  13  14  15
//  KSR = 0     0   0   0   0   1   1   1   1   2   2   2   2   3   3   3   3
//  KSR = 1     0   1   2   3   4   5   6   7   8   9   10  11  12  13  14  15
//
// Note: zero rates are infinite, and are treated separately elsewhere
func (o *operator) ComputeRates() {
	scaleSh := 2
	if o.KeyScaleRate {
		scaleSh = 0
	}
	combinedRate := o.AttackRate*4 + (o.Chan.GetKeyScaleNumber() >> scaleSh)
	rateHi := combinedRate >> 2
	rateLo := combinedRate & 3

	o.AttackShift = uint16(0)
	if rateHi < 12 {
		o.AttackShift = 12 - rateHi
	}
	o.AttackMask = (1 << o.AttackShift) - 1
	o.AttackAdd = 1 << (rateHi - 12)
	if rateHi < 12 {
		o.AttackAdd = 1
	}
	o.AttackTab = opalRateTables[rateLo][:]

	// Attack rate of 15 is always instant
	if o.AttackRate == 15 {
		o.AttackAdd = 0xFFF
	}

	combinedRate = o.DecayRate*4 + (o.Chan.GetKeyScaleNumber() >> scaleSh)
	rateHi = combinedRate >> 2
	rateLo = combinedRate & 3

	o.DecayShift = uint16(0)
	if rateHi < 12 {
		o.DecayShift = 12 - rateHi
	}
	o.DecayMask = (1 << o.DecayShift) - 1
	o.DecayAdd = 1 << (rateHi - 12)
	if rateHi < 12 {
		o.DecayAdd = 1
	}
	o.DecayTab = opalRateTables[rateLo][:]

	combinedRate = o.ReleaseRate*4 + (o.Chan.GetKeyScaleNumber() >> scaleSh)
	rateHi = combinedRate >> 2
	rateLo = combinedRate & 3

	o.ReleaseShift = uint16(0)
	if rateHi < 12 {
		o.ReleaseShift = 12 - rateHi
	}
	o.ReleaseMask = (1 << o.ReleaseShift) - 1
	o.ReleaseAdd = 1 << (rateHi - 12)
	if rateHi < 12 {
		o.ReleaseAdd = 1
	}
	o.ReleaseTab = opalRateTables[rateLo][:]
}

var levtab = []uint8{
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 8, 12, 16, 20, 24, 28, 32,
	0, 0, 0, 0, 0, 12, 20, 28, 32, 40, 44, 48, 52, 56, 60, 64,
	0, 0, 0, 20, 32, 44, 52, 60, 64, 72, 76, 80, 84, 88, 92, 96,
	0, 0, 32, 52, 64, 76, 84, 92, 96, 104, 108, 112, 116, 120, 124, 128,
	0, 32, 64, 84, 96, 108, 116, 124, 128, 136, 140, 144, 148, 152, 156, 160,
	0, 64, 96, 116, 128, 140, 148, 156, 160, 168, 172, 176, 180, 184, 188, 192,
	0, 96, 128, 148, 160, 172, 180, 188, 192, 200, 204, 208, 212, 216, 220, 224,
}

// ComputeKeyScaleLevel - Compute the operator's key scale level. This changes based on the channel frequency/octave and
// operator key scale value.
func (o *operator) ComputeKeyScaleLevel() {
	// This uses a combined value of the top four bits of frequency with the octave/block
	i := uint16(o.Chan.GetOctave()<<4) | (o.Chan.GetFreq() >> 6)
	o.KeyScaleLevel = uint16(levtab[i]) >> o.KeyScaleShift
}

func (o *operator) SetMaster(opal *Opal) {
	o.Master = opal
}

func (o *operator) SetChannel(ch *channel) {
	o.Chan = ch
}

// A single channel, which can contain two or more operators
type channel struct {
	Op [4]*operator

	Master         *Opal  // Master object
	Freq           uint16 // Frequency; actually it's a phase stepping value
	Octave         uint16 // Also known as "block" in Yamaha parlance
	PhaseStep      uint32
	KeyScaleNumber uint16
	FeedbackShift  uint16
	ModulationType uint16
	ChannelPair    *channel
	Enable         bool
	LeftEnable     bool
	RightEnable    bool
}

// Init - Channel constructor.
func (c *channel) Init() {
	c.Master = nil
	c.Freq = 0
	c.Octave = 0
	c.PhaseStep = 0
	c.KeyScaleNumber = 0
	c.FeedbackShift = 0
	c.ModulationType = 0
	c.ChannelPair = nil
	c.Enable = true
}

func (c *channel) SetMaster(opal *Opal) {
	c.Master = opal
}

func (c *channel) SetOperators(ops ...*operator) {
	for i, o := range ops {
		c.Op[i] = o
		if o != nil {
			o.SetChannel(c)
		}
	}
}

func (c *channel) SetEnable(on bool) {
	c.Enable = on
}

func (c *channel) SetChannelPair(pair *channel) {
	c.ChannelPair = pair
}

func (c *channel) GetFreq() uint16 {
	return c.Freq
}

func (c *channel) GetOctave() uint16 {
	return c.Octave
}

func (c *channel) GetKeyScaleNumber() uint16 {
	return c.KeyScaleNumber
}

func (c *channel) GetModulationType() uint16 {
	return c.ModulationType
}

func (c *channel) GetChannelPair() *channel {
	return c.ChannelPair
}

// Output - Produce output from channel.
func (c *channel) Output() (int16, int16) {
	// Has the channel been disabled?  This is usually a result of the 4-op enables being used to
	// disable the secondary channel in each 4-op pair
	if !c.Enable {
		return 0, 0
	}

	vibrato := int16(c.Freq>>7) & 7
	if !c.Master.VibratoDepth {
		vibrato >>= 1
	}

	// 0  3  7  3  0  -3  -7  -3
	clk := c.Master.VibratoClock
	if (clk & 3) == 0 {
		vibrato = 0 // Position 0 and 4 is zero
	} else {
		if (clk & 1) != 0 {
			vibrato >>= 1 // Odd positions are half the magnitude
		}
		if (clk & 4) != 0 {
			vibrato = -vibrato // The second half positions are negative
		}
	}

	vibrato <<= c.Octave

	// Combine individual operator outputs
	var out, acc int16

	// Running in 4-op mode?
	if c.ChannelPair != nil {
		// Get the secondary channel's modulation type.  This is the only thing from the secondary
		// channel that is used
		if c.ChannelPair.GetModulationType() == 0 {
			if c.ModulationType == 0 {
				// feedback -> modulator -> modulator -> modulator -> carrier
				out = c.Op[0].Output(c.KeyScaleNumber, c.PhaseStep, vibrato, 0, int16(c.FeedbackShift))
				out = c.Op[1].Output(c.KeyScaleNumber, c.PhaseStep, vibrato, out, 0)
				out = c.Op[2].Output(c.KeyScaleNumber, c.PhaseStep, vibrato, out, 0)
				out = c.Op[3].Output(c.KeyScaleNumber, c.PhaseStep, vibrato, out, 0)
			} else {
				// (feedback -> carrier) + (modulator -> modulator -> carrier)
				out = c.Op[0].Output(c.KeyScaleNumber, c.PhaseStep, vibrato, 0, int16(c.FeedbackShift))
				acc = c.Op[1].Output(c.KeyScaleNumber, c.PhaseStep, vibrato, 0, 0)
				acc = c.Op[2].Output(c.KeyScaleNumber, c.PhaseStep, vibrato, acc, 0)
				out += c.Op[3].Output(c.KeyScaleNumber, c.PhaseStep, vibrato, acc, 0)
			}
		} else {
			if c.ModulationType == 0 {
				// (feedback -> modulator -> carrier) + (modulator -> carrier)
				out = c.Op[0].Output(c.KeyScaleNumber, c.PhaseStep, vibrato, 0, int16(c.FeedbackShift))
				out = c.Op[1].Output(c.KeyScaleNumber, c.PhaseStep, vibrato, out, 0)
				acc = c.Op[2].Output(c.KeyScaleNumber, c.PhaseStep, vibrato, 0, 0)
				out += c.Op[3].Output(c.KeyScaleNumber, c.PhaseStep, vibrato, acc, 0)
			} else {
				// (feedback -> carrier) + (modulator -> carrier) + carrier
				out = c.Op[0].Output(c.KeyScaleNumber, c.PhaseStep, vibrato, 0, int16(c.FeedbackShift))
				acc = c.Op[1].Output(c.KeyScaleNumber, c.PhaseStep, vibrato, 0, 0)
				out += c.Op[2].Output(c.KeyScaleNumber, c.PhaseStep, vibrato, acc, 0)
				out += c.Op[3].Output(c.KeyScaleNumber, c.PhaseStep, vibrato, 0, 0)
			}
		}
	} else {
		// Standard 2-op mode
		if c.ModulationType == 0 {
			// Frequency modulation (well, phase modulation technically)
			out = c.Op[0].Output(c.KeyScaleNumber, c.PhaseStep, vibrato, 0, int16(c.FeedbackShift))
			out = c.Op[1].Output(c.KeyScaleNumber, c.PhaseStep, vibrato, out, 0)
		} else {
			// Additive
			out = c.Op[0].Output(c.KeyScaleNumber, c.PhaseStep, vibrato, 0, int16(c.FeedbackShift))
			out += c.Op[1].Output(c.KeyScaleNumber, c.PhaseStep, vibrato, 0, 0)
		}
	}

	l := int16(0)
	if c.LeftEnable {
		l = out
	}
	r := int16(0)
	if c.RightEnable {
		r = out
	}

	return l, r
}

// SetFrequencyLow - Set phase step for operators using this channel.
func (c *channel) SetFrequencyLow(freq uint16) {
	c.Freq = (c.Freq & 0x300) | (freq & 0xFF)
	c.ComputePhaseStep()
}

// SetFrequencyHigh - Set phase step for operators using this channel.
func (c *channel) SetFrequencyHigh(freq uint16) {
	c.Freq = (c.Freq & 0xFF) | ((freq & 3) << 8)
	c.ComputePhaseStep()

	// Only the high bits of Freq affect the Key Scale No.
	c.ComputeKeyScaleNumber()
}

// SetOctave - Set the octave of the channel (0 to 7).
func (c *channel) SetOctave(oct uint16) {
	c.Octave = oct & 7
	c.ComputePhaseStep()
	c.ComputeKeyScaleNumber()
}

// SetKeyOn - Keys the channel on/off.
func (c *channel) SetKeyOn(on bool) {
	c.Op[0].SetKeyOn(on)
	c.Op[1].SetKeyOn(on)
}

// SetLeftEnable - Enable left stereo channel.
func (c *channel) SetLeftEnable(on bool) {
	c.LeftEnable = on
}

// SetRightEnable - Enable right stereo channel.
func (c *channel) SetRightEnable(on bool) {
	c.RightEnable = on
}

// SetFeedback - Set the channel feedback amount.
func (c *channel) SetFeedback(val uint16) {
	if val == 0 {
		c.FeedbackShift = 0
		return
	}
	c.FeedbackShift = 9 - val
}

// SetModulationType - Set frequency modulation/additive modulation
func (c *channel) SetModulationType(typ uint16) {
	c.ModulationType = typ
}

// ComputePhaseStep - Compute the stepping factor for the operator waveform phase based on the frequency and octave
// values of the channel.
func (c *channel) ComputePhaseStep() {
	c.PhaseStep = uint32(c.Freq) << c.Octave
}

// ComputeKeyScaleNumber - Compute the key scale number and key scale levels.
// From the Yamaha data sheet this is the block/octave number as bits 3-1, with bit 0 coming from
// the MSB of the frequency if NoteSel is 1, and the 2nd MSB if NoteSel is 0.
func (c *channel) ComputeKeyScaleNumber() {
	lsb := uint16(c.Freq>>8) & 1
	if c.Master.NoteSel {
		lsb = c.Freq >> 9
	}
	c.KeyScaleNumber = c.Octave<<1 | lsb

	// Get the channel operators to recompute their rates as they're dependent on this number.  They
	// also need to recompute their key scale level
	for _, op := range c.Op {
		if op == nil {
			continue
		}

		op.ComputeRates()
		op.ComputeKeyScaleLevel()
	}
}

// Opal - Opal class.
type Opal struct {
	SampleRate   int32
	SampleAccum  int32
	LastOutput   [2]int16
	CurrOutput   [2]int16
	Chan         [NumChannels]channel
	Op           [NumOperators]operator
	Clock        uint16
	TremoloClock uint16
	TremoloLevel uint16
	VibratoTick  uint16
	VibratoClock uint16
	NoteSel      bool
	TremoloDepth bool
	VibratoDepth bool
	//ExpTable     [256]uint16
	//LogSinTable  [256]uint16
}

var chanOps = []int{
	0, 1, 2, 6, 7, 8, 12, 13, 14, 18, 19, 20, 24, 25, 26, 30, 31, 32,
}

// Init - Initialise the emulation.
func (o *Opal) Init(sampleRate int) {
	o.Clock = 0
	o.TremoloClock = 0
	o.TremoloLevel = 0
	o.VibratoTick = 0
	o.VibratoClock = 0
	o.NoteSel = false
	o.TremoloDepth = false
	o.VibratoDepth = false

	//	// Build the exponentiation table (reversed from the official OPL3 ROM)
	//	for i := 0; i < 0x100; i++ {
	//		o.ExpTable[i] = uint16(math.Round((math.Pow(2, float64(0xFF - i) / 256.0) - 1) * 1024));
	//	}
	//
	//	// Build the log-sin table
	//	for i := 0; i < 0x100; i++ {
	//	    o.LogSinTable[i] = uint16(math.Round(-math.Log2(math.Sin((float64(i) + 0.5) * math.Pi / 256 / 2)) * 256))
	//	}

	// Let sub-objects know where to find us
	for i := range o.Op {
		op := &o.Op[i]
		op.Init()
		op.SetMaster(o)
	}

	for i := range o.Chan {
		ch := &o.Chan[i]
		ch.Init()
		ch.SetMaster(o)
	}

	// Add the operators to the channels.  Note, some channels can't use all the operators
	// FIXME: put this into a separate routine
	for i := range o.Chan {
		ch := &o.Chan[i]
		op := chanOps[i]
		switch {
		case i < 3, i >= 9 && i < 12:
			ch.SetOperators(&o.Op[op], &o.Op[op+3], &o.Op[op+6], &o.Op[op+9])
		default:
			ch.SetOperators(&o.Op[op], &o.Op[op+3], nil, nil)
		}
	}

	// Initialise the operator rate data.  We can't do this in the Operator constructor as it
	// relies on referencing the master and channel objects
	for i := range o.Op {
		o.Op[i].ComputeRates()
	}

	o.SetSampleRate(sampleRate)
}

// SetSampleRate - Change the sample rate.
func (o *Opal) SetSampleRate(sampleRate int) {
	// Sanity
	if sampleRate == 0 {
		sampleRate = OPL3SampleRate
	}

	o.SampleRate = int32(sampleRate)
	o.SampleAccum = 0
	o.LastOutput[0] = 0
	o.LastOutput[1] = 0
	o.CurrOutput[0] = 0
	o.CurrOutput[1] = 0
}

var opLookup = []int8{
	//  00  01  02  03  04  05  06  07  08  09  0A  0B  0C  0D  0E  0F
	0, 1, 2, 3, 4, 5, -1, -1, 6, 7, 8, 9, 10, 11, -1, -1,
	//  10  11  12  13  14  15  16  17  18  19  1A  1B  1C  1D  1E  1F
	12, 13, 14, 15, 16, 17, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
}

// Port - Write a value to an OPL3 register.
func (o *Opal) Port(regNum uint16, val uint8) {

	// Is it BD, the one-off register stuck in the middle of the register array?
	if regNum == 0xBD {
		o.TremoloDepth = (val & 0x80) != 0
		o.VibratoDepth = (val & 0x40) != 0
		return
	}

	typ := regNum & 0xE0
	switch {
	case typ == 0x00: // Global registers
		o.portGlobalRegs(regNum, val)
	case typ >= 0xA0 && typ <= 0xC0: // Channel registers
		o.portChannelRegs(regNum, val)
	case typ >= 0x20 && typ <= 0x80, typ == 0xE0: // Operator registers
		o.portOperatorRegs(regNum, val)
	}
}

// Sample - Generate sample. Every time you call this you will get two signed 16-bit samples (one for each
// stereo channel) which will sound correct when played back at the sample rate given when the
// class was constructed.
func (o *Opal) Sample() (int16, int16) {
	// If the destination sample rate is higher than the OPL3 sample rate, we need to skip ahead
	for o.SampleAccum >= o.SampleRate {
		o.LastOutput[0] = o.CurrOutput[0]
		o.LastOutput[1] = o.CurrOutput[1]

		o.CurrOutput[0], o.CurrOutput[1] = o.Output()

		o.SampleAccum -= o.SampleRate
	}

	// Mix with the partial accumulation
	omblend := int32(o.SampleRate - o.SampleAccum)
	l := int16((int32(o.LastOutput[0])*omblend + int32(o.CurrOutput[0])*o.SampleAccum) / o.SampleRate)
	r := int16((int32(o.LastOutput[1])*omblend + int32(o.CurrOutput[1])*o.SampleAccum) / o.SampleRate)

	o.SampleAccum += OPL3SampleRate

	return l, r
}

// Output - Produce final output from the chip.  This is at the OPL3 sample-rate.
func (o *Opal) Output() (int16, int16) {
	lmix := int32(0)
	rmix := int32(0)

	// Sum the output of each channel
	for i := range o.Chan {
		chanL, chanR := o.Chan[i].Output()
		lmix += int32(chanL)
		rmix += int32(chanR)
	}

	// Clamp
	l := int16(0)
	switch {
	case lmix < -0x8000:
		l = -0x8000
	case lmix > 0x7FFF:
		l = 0x7FFF
	default:
		l = int16(lmix)
	}

	r := int16(0)
	switch {
	case rmix < -0x8000:
		r = -0x8000
	case rmix > 0x7FFF:
		r = 0x7FFF
	default:
		r = int16(rmix)
	}

	o.Clock++

	// Tremolo.  According to this post, the OPL3 tremolo is a 13,440 sample length triangle wave
	// with a peak at 26 and a trough at 0 and is simply added to the logarithmic level accumulator
	//      http://forums.submarine.org.uk/phpBB/viewtopic.php?f=9&t=1171
	o.TremoloClock = (o.TremoloClock + 1) % 13440
	o.TremoloLevel = o.TremoloClock
	if o.TremoloClock >= 13440/2 {
		o.TremoloLevel = (13440 - o.TremoloClock)
	}
	o.TremoloLevel /= 256
	if !o.TremoloDepth {
		o.TremoloLevel >>= 2
	}

	// Vibrato.  This appears to be a 8 sample long triangle wave with a magnitude of the three
	// high bits of the channel frequency, positive and negative, divided by two if the vibrato
	// depth is zero.  It is only cycled every 1,024 samples.
	o.VibratoTick++
	if o.VibratoTick >= 1024 {
		o.VibratoTick = 0
		o.VibratoClock = (o.VibratoClock + 1) & 7
	}

	return l, r
}

func (o *Opal) portOperatorRegs(regNum uint16, val uint8) {
	// Convert to operator number
	opNum := opLookup[regNum&0x1F]

	// Valid register?
	if opNum < 0 {
		return
	}

	// Is it the other bank of operators?
	if (regNum & 0x100) != 0 {
		opNum += 18
	}

	op := &o.Op[opNum]

	// Do specific registers
	typ := regNum & 0xE0
	switch typ {
	case 0x20: // Tremolo Enable / Vibrato Enable / Sustain Mode / Envelope Scaling / Frequency Multiplier
		op.SetTremoloEnable((val & 0x80) != 0)
		op.SetVibratoEnable((val & 0x40) != 0)
		op.SetSustainMode((val & 0x20) != 0)
		op.SetEnvelopeScaling((val & 0x10) != 0)
		op.SetFrequencyMultiplier(uint16(val & 15))

	case 0x40: // Key Scale / Output Level
		op.SetKeyScale(uint16(val >> 6))
		op.SetOutputLevel(uint16(val & 0x3F))

	case 0x60: // Attack Rate / Decay Rate
		op.SetAttackRate(uint16(val >> 4))
		op.SetDecayRate(uint16(val & 15))

	case 0x80: // Sustain Level / Release Rate
		op.SetSustainLevel(uint16(val >> 4))
		op.SetReleaseRate(uint16(val & 15))

	case 0xE0: // Waveform
		op.SetWaveform(uint16(val & 7))
	}
}

func (o *Opal) portChannelRegs(regNum uint16, val uint8) {
	// Convert to channel number
	chanNum := int(regNum) & 15

	// Valid channel?
	if chanNum >= 9 {
		return
	}

	// Is it the other bank of channels?
	if (regNum & 0x100) != 0 {
		chanNum += 9
	}

	ch := &o.Chan[chanNum]

	// Registers Ax and Bx affect both channels
	chans := []*channel{ch}
	if cp := ch.GetChannelPair(); cp != nil {
		chans = append(chans, cp)
	}

	// Do specific registers
	switch regNum & 0xF0 {
	case 0xA0: // Frequency low
		for _, chn := range chans {
			chn.SetFrequencyLow(uint16(val))
		}

	case 0xB0: // Key-on / Octave / Frequency High
		for _, chn := range chans {
			chn.SetKeyOn((val & 0x20) != 0)
			chn.SetOctave(uint16(val >> 2 & 7))
			chn.SetFrequencyHigh(uint16(val & 3))
		}

	case 0xC0: // Right Stereo Channel Enable / Left Stereo Channel Enable / Feedback Factor / Modulation Type
		ch.SetRightEnable((val & 0x20) != 0)
		ch.SetLeftEnable((val & 0x10) != 0)
		ch.SetFeedback(uint16(val >> 1 & 7))
		ch.SetModulationType(uint16(val & 1))
	}
}

func (o *Opal) portGlobalRegs(regNum uint16, val uint8) {
	switch regNum {
	case 0x104: // 4-OP enables
		o.port104(val)

	case 0x08: // CSW / Note-sel
		o.NoteSel = (val & 0x40) != 0
		// Get the channels to recompute the Key Scale No. as this varies based on NoteSel
		for i := range o.Chan {
			o.Chan[i].ComputeKeyScaleNumber()
		}
	}
}

func (o *Opal) port104(val uint8) {
	// Enable/disable channels based on which 4-op enables
	mask := uint8(1)
	for i := 0; i < 6; i++ {
		// The 4-op channels are 0, 1, 2, 9, 10, 11
		ch := uint16(i)
		if i >= 3 {
			ch = uint16(i) + 6
		}
		primary := &o.Chan[ch]
		secondary := &o.Chan[ch+3]

		if (val & mask) != 0 {
			// Let primary channel know it's controlling the secondary channel
			primary.SetChannelPair(secondary)

			// Turn off the second channel in the pair
			secondary.SetEnable(false)
		} else {
			// Let primary channel know it's no longer controlling the secondary channel
			primary.SetChannelPair(nil)

			// Turn on the second channel in the pair
			secondary.SetEnable(true)
		}
		mask <<= 1
	}
}

// NewOpal create a new Opal instance
func NewOpal(sampleRate uint32) *Opal {
	o := Opal{}
	o.Init(int(sampleRate))
	return &o
}

// WriteReg write to the Opal on a specific register
func (o *Opal) WriteReg(reg uint32, val uint8) {
	o.Port(uint16(reg), uint8(val))
}

// GenerateBlock2 generates a block of mono 16-bit output data from the Opal
func (o *Opal) GenerateBlock2(count uint, output []int32) {
	for i := uint(0); i < count; i++ {
		l, r := o.Sample()
		output[i] = (int32(l) + int32(r)) / 2
	}
}
