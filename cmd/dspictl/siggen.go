package main

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/suhlig/dspi"
)

func newSiggenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "siggen",
		Short: "Onboard test signal generator",
		Long: `Control the DSPi onboard test signal generator.

The generator can inject 15 measurement/diagnostic signals directly into the
output pipeline without a host audio stream. Use this for sweeps, channel ID,
noise, impulses, and more.

Examples:
  dspictl siggen types
  dspictl siggen start --type sine --channels 0,1 --level -20 --freq 1000
  dspictl siggen start --type sweep-log --channels 0 --level -20 --f1 20 --f2 20000 --duration 10000
  dspictl siggen start --type channel-id --channels 0,1,2,3
  dspictl siggen stop`,
	}

	cmd.AddCommand(newSiggenTypesCmd())
	cmd.AddCommand(newSiggenStatusCmd())
	cmd.AddCommand(newSiggenStartCmd())
	cmd.AddCommand(newSiggenConfigCmd())
	cmd.AddCommand(newSiggenStopCmd())

	return cmd
}

func newSiggenTypesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "types",
		Short: "List supported signal types",
		RunE:  runSiggenTypes,
	}
}

func newSiggenStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current signal-generator status",
		RunE:  runSiggenStatus,
	}
}

func newSiggenStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Configure and start the signal generator",
		Args:  cobra.NoArgs,
		RunE:  runSiggenStart,
	}

	addSiggenConfigFlags(cmd)
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("channels")

	return cmd
}

func newSiggenConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Stage a signal-generator config without starting",
		Args:  cobra.NoArgs,
		RunE:  runSiggenConfig,
	}

	addSiggenConfigFlags(cmd)
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("channels")

	return cmd
}

func newSiggenStopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the signal generator",
		Args:  cobra.NoArgs,
		RunE:  runSiggenStop,
	}

	cmd.Flags().Bool("now", false, "Hard stop immediately without fade")

	return cmd
}

func addSiggenConfigFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("type", "t", "", "Signal type (name or ID)")
	cmd.Flags().StringSliceP("channels", "c", nil, "Output channel numbers to drive")
	cmd.Flags().StringSlice("invert", nil, "Channels to invert polarity")
	cmd.Flags().Float64P("level", "l", -20, "Peak level in dBFS")
	cmd.Flags().Bool("raw", false, "Bypass per-channel crossover + PEQ")
	cmd.Flags().Bool("decorr", false, "Decorrelate noise per channel")
	cmd.Flags().Bool("walk", false, "Play masked channels one at a time")
	cmd.Flags().Uint32P("duration", "d", 0, "Duration in ms (continuous total / sweep length)")
	cmd.Flags().Uint16P("repeat", "r", 0, "Repeat count (0 = infinite for sweeps/patterns)")
	cmd.Flags().Uint16P("gap", "g", 0, "Inter-cycle silence in ms")

	cmd.Flags().Float64("p1", 0, "Parameter 1")
	cmd.Flags().Float64("p2", 0, "Parameter 2")
	cmd.Flags().Float64("p3", 0, "Parameter 3")
	cmd.Flags().Float64("p4", 0, "Parameter 4")

	// Convenience aliases for common parameters.
	cmd.Flags().Float64("freq", 0, "Alias for p1 (sine/square)")
	cmd.Flags().Float64("f1", 0, "Alias for p1 (sweeps / tone pair)")
	cmd.Flags().Float64("f2", 0, "Alias for p2 (sweeps / tone pair)")
	cmd.Flags().Float64("period", 0, "Alias for p1 (impulse / clicks / polarity)")
	cmd.Flags().Float64("pulse-width", 0, "Alias for p1 (polarity)")
	cmd.Flags().Float64("on-cycles", 0, "Alias for p2 (tone burst)")
	cmd.Flags().Float64("off-cycles", 0, "Alias for p3 (tone burst)")
	cmd.Flags().Float64("edge-cycles", 0, "Alias for p4 (tone burst)")
	cmd.Flags().Float64("ratio", 0, "Alias for p3 (tone pair)")
	cmd.Flags().Float64("count", 0, "Alias for p1 (multitone)")
	cmd.Flags().Float64("tone-lo", 0, "Alias for p2 (multitone)")
	cmd.Flags().Float64("tone-hi", 0, "Alias for p3 (multitone)")
	cmd.Flags().Float64("pattern", 0, "Alias for p1 (ISP)")
	cmd.Flags().Float64("blip", 0, "Alias for p1 (channel ID)")
	cmd.Flags().Float64("steps", 0, "Alias for p3 (sweep step)")
	cmd.Flags().Float64("dwell", 0, "Alias for p4 (sweep step)")

	_ = cmd.RegisterFlagCompletionFunc("type", completeSiggenTypes)
	_ = cmd.RegisterFlagCompletionFunc("channels", completeOutputChannels)
	_ = cmd.RegisterFlagCompletionFunc("invert", completeOutputChannels)
}

func completeSiggenTypes(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	choices := make([]string, 0, dspi.SiggenTypeCount)

	for i := range dspi.SiggenTypeCount {
		choices = append(choices, dspi.SiggenType(i).String())
	}

	return choices, cobra.ShellCompDirectiveNoFileComp
}

func runSiggenTypes(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		caps, err := d.GetSiggenCaps()
		if err != nil {
			slog.Error("getting siggen caps failed", "serial", d.Serial(), "error", err)
			continue
		}

		fmt.Printf("%s: %d signal types, %d output channels, multitone max %d\n",
			d.Serial(), caps.TypeCount, caps.OutputChannels, caps.MultitoneMax)

		for i := 0; i < int(caps.TypeCount); i++ {
			desc, err := d.GetSiggenTypeDesc(i)
			if err != nil {
				slog.Error("getting siggen type desc failed", "serial", d.Serial(), "index", i, "error", err)
				continue
			}

			fmt.Printf("  %2d %-8s %-10s", int(desc.ID), desc.Name, desc.TimingModel)

			params := make([]string, 0, 4)
			for _, p := range desc.Params {
				if p.Semantic == dspi.SiggenParamUnused {
					continue
				}
				params = append(params, fmt.Sprintf("%s=%.2f..%.2f (def %.2f)",
					p.Semantic, p.Min, p.Max, p.Default))
			}

			if len(params) > 0 {
				fmt.Printf(" %s", strings.Join(params, ", "))
			}

			fmt.Println()
		}
	}

	return nil
}

func runSiggenStatus(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		status, err := d.GetSiggenStatus()
		if err != nil {
			slog.Error("getting siggen status failed", "serial", d.Serial(), "error", err)
			continue
		}

		fmt.Printf("%s: state=%s type=%s elapsed=%dms cycles=%d stop=%s",
			d.Serial(), status.State, status.SignalType,
			status.ElapsedMs, status.CyclesDone, status.StopReason)

		if status.ActiveChannel >= 0 {
			fmt.Printf(" walk-ch=%d", status.ActiveChannel)
		}

		if status.CurrentFreq > 0 {
			fmt.Printf(" freq=%.1fHz", status.CurrentFreq)
		}

		fmt.Println()
	}

	return nil
}

func runSiggenStart(cmd *cobra.Command, args []string) error {
	return runSiggenSetAndControl(cmd, true)
}

func runSiggenConfig(cmd *cobra.Command, args []string) error {
	return runSiggenSetAndControl(cmd, false)
}

func runSiggenSetAndControl(cmd *cobra.Command, start bool) error {
	cfg, err := buildSiggenConfig(cmd)
	if err != nil {
		return err
	}

	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		caps, err := d.GetSiggenCaps()
		if err != nil {
			slog.Error("getting siggen caps failed", "serial", d.Serial(), "error", err)
			continue
		}

		deviceCfg, err := cfg.forDevice(caps)
		if err != nil {
			slog.Error("invalid siggen config", "serial", d.Serial(), "error", err)
			continue
		}

		if err := d.SetSiggenConfig(deviceCfg); err != nil {
			slog.Error("setting siggen config failed", "serial", d.Serial(), "error", err)
			continue
		}

		if !start {
			fmt.Printf("%s: signal generator configured\n", d.Serial())
			continue
		}

		if err := d.SiggenStart(); err != nil {
			slog.Error("starting siggen failed", "serial", d.Serial(), "error", err)
			continue
		}

		fmt.Printf("%s: signal generator started (%s on %s)\n",
			d.Serial(), deviceCfg.SignalType, formatMaskU16(deviceCfg.ChannelMask, int(caps.OutputChannels)))
	}

	return nil
}

func runSiggenStop(cmd *cobra.Command, args []string) error {
	now, err := cmd.Flags().GetBool("now")
	if err != nil {
		return fmt.Errorf("getting now flag: %w", err)
	}

	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		if now {
			err = d.SiggenStopNow()
		} else {
			err = d.SiggenStop()
		}

		if err != nil {
			slog.Error("stopping siggen failed", "serial", d.Serial(), "error", err)
			continue
		}

		if now {
			fmt.Printf("%s: signal generator stopped immediately\n", d.Serial())
		} else {
			fmt.Printf("%s: signal generator stopping\n", d.Serial())
		}
	}

	return nil
}

// cliSiggenConfig holds the parsed, not-yet-validated CLI configuration.
type cliSiggenConfig struct {
	typ      dspi.SiggenType
	channels []string
	invert   []string
	level    float64
	raw      bool
	decorr   bool
	walk     bool
	duration uint32
	repeat   uint16
	gap      uint16
	p1       float64
	p2       float64
	p3       float64
	p4       float64
}

// forDevice builds a validated SiggenConfig for a specific device's capabilities.
func (c *cliSiggenConfig) forDevice(caps *dspi.SiggenCaps) (*dspi.SiggenConfig, error) {
	mask, err := parseChannelList(c.channels, int(caps.OutputChannels))
	if err != nil {
		return nil, fmt.Errorf("channels: %w", err)
	}

	if mask == 0 {
		return nil, fmt.Errorf("no output channels selected")
	}

	invert, err := parseChannelList(c.invert, int(caps.OutputChannels))
	if err != nil {
		return nil, fmt.Errorf("invert: %w", err)
	}

	invert &= mask

	switch c.typ {
	case dspi.SiggenTypeSweepLog, dspi.SiggenTypeSweepLin, dspi.SiggenTypeSweepStep:
		if c.duration == 0 {
			return nil, fmt.Errorf("duration must be > 0 for sweep types")
		}
	}

	cfg := &dspi.SiggenConfig{
		SignalType:  c.typ,
		ChannelMask: mask,
		InvertMask:  invert,
		LevelDB:     c.level,
		DurationMs:  c.duration,
		Repeat:      c.repeat,
		GapMs:       c.gap,
		P1:          c.p1,
		P2:          c.p2,
		P3:          c.p3,
		P4:          c.p4,
	}
	cfg.SetFlag(c.raw, c.decorr, c.walk)

	return cfg, nil
}

func buildSiggenConfig(cmd *cobra.Command) (*cliSiggenConfig, error) {
	typStr, err := cmd.Flags().GetString("type")
	if err != nil {
		return nil, fmt.Errorf("getting type flag: %w", err)
	}

	typ, err := dspi.ParseSiggenType(typStr)
	if err != nil {
		return nil, fmt.Errorf("parsing type: %w", err)
	}

	channels, err := cmd.Flags().GetStringSlice("channels")
	if err != nil {
		return nil, fmt.Errorf("getting channels flag: %w", err)
	}

	invert, err := cmd.Flags().GetStringSlice("invert")
	if err != nil {
		return nil, fmt.Errorf("getting invert flag: %w", err)
	}

	level, err := cmd.Flags().GetFloat64("level")
	if err != nil {
		return nil, fmt.Errorf("getting level flag: %w", err)
	}

	raw, err := cmd.Flags().GetBool("raw")
	if err != nil {
		return nil, fmt.Errorf("getting raw flag: %w", err)
	}

	decorr, err := cmd.Flags().GetBool("decorr")
	if err != nil {
		return nil, fmt.Errorf("getting decorr flag: %w", err)
	}

	walk, err := cmd.Flags().GetBool("walk")
	if err != nil {
		return nil, fmt.Errorf("getting walk flag: %w", err)
	}

	duration, err := cmd.Flags().GetUint32("duration")
	if err != nil {
		return nil, fmt.Errorf("getting duration flag: %w", err)
	}

	repeat, err := cmd.Flags().GetUint16("repeat")
	if err != nil {
		return nil, fmt.Errorf("getting repeat flag: %w", err)
	}

	gap, err := cmd.Flags().GetUint16("gap")
	if err != nil {
		return nil, fmt.Errorf("getting gap flag: %w", err)
	}

	p1, p2, p3, p4, err := getSiggenParams(cmd)
	if err != nil {
		return nil, err
	}

	return &cliSiggenConfig{
		typ:      typ,
		channels: channels,
		invert:   invert,
		level:    level,
		raw:      raw,
		decorr:   decorr,
		walk:     walk,
		duration: duration,
		repeat:   repeat,
		gap:      gap,
		p1:       p1,
		p2:       p2,
		p3:       p3,
		p4:       p4,
	}, nil
}

// getSiggenParams reads the generic p1..p4 flags and the convenience aliases,
// giving precedence to explicit p1..p4 values when they were set.
func getSiggenParams(cmd *cobra.Command) (p1, p2, p3, p4 float64, err error) {
	p1, err = getSiggenParam(cmd, "p1", "freq", "f1", "period", "pulse-width", "count", "pattern", "blip")
	if err != nil {
		return 0, 0, 0, 0, err
	}

	p2, err = getSiggenParam(cmd, "p2", "f2", "tone-lo", "on-cycles")
	if err != nil {
		return 0, 0, 0, 0, err
	}

	p3, err = getSiggenParam(cmd, "p3", "ratio", "tone-hi", "off-cycles", "steps")
	if err != nil {
		return 0, 0, 0, 0, err
	}

	p4, err = getSiggenParam(cmd, "p4", "edge-cycles", "dwell")
	if err != nil {
		return 0, 0, 0, 0, err
	}

	return p1, p2, p3, p4, nil
}

func getSiggenParam(cmd *cobra.Command, base string, aliases ...string) (float64, error) {
	if cmd.Flags().Changed(base) {
		return cmd.Flags().GetFloat64(base)
	}

	for _, name := range aliases {
		if cmd.Flags().Changed(name) {
			return cmd.Flags().GetFloat64(name)
		}
	}

	return cmd.Flags().GetFloat64(base)
}

// parseChannelList parses a comma/space-separated list of channel numbers
// into a bit mask. Empty input returns 0.
func parseChannelList(values []string, maxBits int) (uint16, error) {
	var mask uint16
	seen := make(map[int]bool)

	for _, value := range values {
		for part := range strings.SplitSeq(value, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}

			n, err := strconv.Atoi(part)
			if err != nil {
				return 0, fmt.Errorf("invalid channel number %q: %w", part, err)
			}

			if n < 0 || n >= maxBits {
				return 0, fmt.Errorf("channel %d out of range (0-%d)", n, maxBits-1)
			}

			if seen[n] {
				continue
			}

			seen[n] = true
			mask |= 1 << uint(n)
		}
	}

	return mask, nil
}
