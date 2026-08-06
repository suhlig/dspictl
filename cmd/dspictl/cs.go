package main

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/suhlig/dspi"
)

func newCsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cs",
		Short: "Control Surfaces (physical controls/indicators on user GPIOs)",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show the CS status packet (last SET result, dirty flag, active bindings)",
		Args:  cobra.NoArgs,
		RunE:  runCsStatus,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "caps",
		Short: "Show the CS capability header, type table, and per-noun descriptors",
		Args:  cobra.NoArgs,
		RunE:  runCsCaps,
	})

	bindingCmd := &cobra.Command{
		Use:   "binding",
		Short: "Get or set a binding slot",
	}

	bindingCmd.AddCommand(&cobra.Command{
		Use:   "get <slot>",
		Short: "Show the live binding of a slot (0-15)",
		Args:  cobra.ExactArgs(1),
		RunE:  runCsBindingGet,
	})

	setCmd := &cobra.Command{
		Use:   "set <slot>",
		Short: "Upload a binding to a slot (apply-live preview; cs save persists)",
		Args:  cobra.ExactArgs(1),
		RunE:  runCsBindingSet,
	}
	setCmd.Flags().String("type", "", "Component type: none, button, switch, pot, encoder, led, led-pwm, ir")
	setCmd.Flags().String("noun", "", "Parameter noun, e.g. user-volume, master-volume, filter-freq, preset")
	setCmd.Flags().String("action", "", "Action: adjust, step, inc, dec, toggle, set, follow, trigger, ind-equals, momentary, ind-above, ind-level")
	setCmd.Flags().String("gpio", "", "GPIO pin(s), comma-separated (e.g. 26 or 2,3)")
	setCmd.Flags().String("event", "", "Button event: press, long, double")
	setCmd.Flags().Int("target", 0, "Channel index for targeted nouns")
	setCmd.Flags().Int("index", 0, "Filter band for band-targeted nouns")
	setCmd.Flags().Int("value", 0, "SET/MOMENTARY target or IND_EQUALS/IND_ABOVE comparand")
	setCmd.Flags().Int("step", 0, "STEP/INC/DEC size (0 = per-unit default)")
	setCmd.Flags().Int("range-min", 0, "Pot/IND_LEVEL span minimum (0 = noun full range)")
	setCmd.Flags().Int("range-max", 0, "Pot/IND_LEVEL span maximum (0 = noun full range)")
	setCmd.Flags().Bool("invert", false, "Set the CS_FLAG_INVERT flag")
	bindingCmd.AddCommand(setCmd)
	cmd.AddCommand(bindingCmd)

	nameCmd := &cobra.Command{
		Use:   "name",
		Short: "Get or set a slot's user label",
	}

	nameCmd.AddCommand(&cobra.Command{
		Use:   "get <slot>",
		Short: "Show the live name of a slot",
		Args:  cobra.ExactArgs(1),
		RunE:  runCsNameGet,
	})

	nameCmd.AddCommand(&cobra.Command{
		Use:   "set <slot> <name>",
		Short: "Set the slot name (\"\" clears it; cs save persists)",
		Args:  cobra.ExactArgs(2),
		RunE:  runCsNameSet,
	})
	cmd.AddCommand(nameCmd)

	irCmd := &cobra.Command{
		Use:   "ir",
		Short: "IR remote commands (16 sub-slots)",
	}

	irCmd.AddCommand(&cobra.Command{
		Use:   "get <subslot>",
		Short: "Show the live IR command of a sub-slot (0-15)",
		Args:  cobra.ExactArgs(1),
		RunE:  runCsIRGet,
	})

	irSetCmd := &cobra.Command{
		Use:   "set <subslot>",
		Short: "Upload an IR command to a sub-slot (apply-live preview; cs save persists)",
		Args:  cobra.ExactArgs(1),
		RunE:  runCsIRSet,
	}
	irSetCmd.Flags().String("noun", "", "Parameter noun, e.g. preset, user-mute")
	irSetCmd.Flags().String("action", "", "Action: inc, dec, toggle, set, trigger, momentary")
	irSetCmd.Flags().String("protocol", "", "IR protocol: nec, rc5, rc6, hash")
	irSetCmd.Flags().String("code", "", "Learned code as hex, e.g. 0x00FF10EF")
	irSetCmd.Flags().Int("target", 0, "Channel index for targeted nouns")
	irSetCmd.Flags().Int("index", 0, "Filter band for band-targeted nouns")
	irSetCmd.Flags().Int("value", 0, "SET/MOMENTARY target")
	irSetCmd.Flags().Int("step", 0, "INC/DEC size (0 = per-unit default)")
	irCmd.AddCommand(irSetCmd)

	irCmd.AddCommand(&cobra.Command{
		Use:               "learn [arm|cancel|read]",
		Short:             "Arm the IR receiver to capture the next press, cancel, or read the result",
		Args:              cobra.MaximumNArgs(1),
		RunE:              runCsIRLearn,
		ValidArgsFunction: completeChoices([]string{"arm", "cancel", "read"}),
	})
	cmd.AddCommand(irCmd)

	cmd.AddCommand(&cobra.Command{
		Use:   "save",
		Short: "Persist the whole live CS config (bindings, IR commands, names) to flash",
		Args:  cobra.NoArgs,
		RunE:  runCsSave,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "revert",
		Short: "Discard the live preview and re-apply the stored CS config",
		Args:  cobra.NoArgs,
		RunE:  runCsRevert,
	})

	return cmd
}

func runCsStatus(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		st, err := d.GetCsStatus()
		if err != nil {
			slog.Error("getting CS status failed", "serial", d.Serial(), "error", err)
			continue
		}

		dirty := "clean"
		if st.Dirty {
			dirty = "dirty (unsaved preview)"
		}
		fmt.Printf("%s: %s, %s\n", d.Serial(), dspi.CsStatusName(st.LastStatus), dirty)

		lastSlot := st.LastSlot
		if lastSlot == dspi.CsLastSlotSave {
			fmt.Printf("  last SET: save/revert\n")
		} else if lastSlot&dspi.CsLastSlotIRFlag != 0 {
			fmt.Printf("  last SET: IR sub-slot %d\n", lastSlot&^dspi.CsLastSlotIRFlag)
		} else {
			fmt.Printf("  last SET: slot %d\n", lastSlot)
		}

		for i := range dspi.CsMaxBindings {
			if st.ActiveMask&(1<<i) != 0 {
				fmt.Printf("  slot %2d: active\n", i)
			} else if st.SlotStatus[i] != 0 {
				fmt.Printf("  slot %2d: %s\n", i, dspi.CsStatusName(st.SlotStatus[i]))
			}
		}
		for i := range dspi.CsMaxIRCommands {
			if st.IRActiveMask&(1<<i) != 0 {
				fmt.Printf("  IR sub-slot %2d: active\n", i)
			}
		}
		if st.IRLearnState != dspi.CsIrLearnIdle {
			fmt.Printf("  IR learn: %s\n", csIRLearnStateName(st.IRLearnState))
		}
	}

	return nil
}

func csIRLearnStateName(state uint8) string {
	switch state {
	case dspi.CsIrLearnIdle:
		return "idle"
	case dspi.CsIrLearnArmed:
		return "armed"
	case dspi.CsIrLearnDone:
		return "done"
	case dspi.CsIrLearnTimeout:
		return "timeout"
	default:
		return fmt.Sprintf("unknown(%d)", state)
	}
}

func runCsCaps(cmd *cobra.Command, args []string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		caps, err := d.GetCsCaps()
		if err != nil {
			slog.Error("getting CS caps failed", "serial", d.Serial(), "error", err)
			continue
		}

		fmt.Printf("%s: caps v%d, %d bindings, %d types, %d nouns, %d IR commands\n",
			d.Serial(), caps.CapsVersion, caps.MaxBindings, caps.TypeCount, caps.NounCount, caps.MaxIRCommands)
		for t := range dspi.CsTypeCount {
			desc := caps.Types[t]
			fmt.Printf("  %-8s actions=0x%04X pins=%d class=%d\n",
				dspi.CsTypeName(uint8(t)), desc.Actions, desc.PinCount, desc.PinClass)
		}
	}

	return nil
}

func runCsBindingGet(cmd *cobra.Command, args []string) error {
	slot, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid slot: %w", err)
	}

	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		b, err := d.GetCsBinding(slot)
		if err != nil {
			slog.Error("getting CS binding failed", "serial", d.Serial(), "error", err)
			continue
		}

		fmt.Printf("%s: slot %d: %s -> %s (%s)\n",
			d.Serial(), slot, dspi.CsTypeName(b.Type), dspi.CsNounName(b.Noun), csActionName(b.Action))
		if b.Type == dspi.CsTypeNone {
			continue
		}
		if b.GPIO[1] != dspi.CsGPIOUnused {
			fmt.Printf("  GPIO %d,%d", b.GPIO[0], b.GPIO[1])
		} else {
			fmt.Printf("  GPIO %d", b.GPIO[0])
		}
		if b.Type == dspi.CsTypeButton {
			fmt.Printf("  event=%s", csEventName(b.Event))
		}
		fmt.Printf("\n")
		if b.Target != 0 || b.Index != 0 {
			fmt.Printf("  target=%d index=%d\n", b.Target, b.Index)
		}
		fmt.Printf("  value=%d step=%d range=[%d,%d] flags=0x%02X\n",
			b.Value, b.Step, b.RangeMin, b.RangeMax, b.Flags)
	}

	return nil
}

func csActionName(action uint8) string {
	switch action {
	case dspi.CsActAdjust:
		return "adjust"
	case dspi.CsActStep:
		return "step"
	case dspi.CsActInc:
		return "inc"
	case dspi.CsActDec:
		return "dec"
	case dspi.CsActToggle:
		return "toggle"
	case dspi.CsActSet:
		return "set"
	case dspi.CsActFollow:
		return "follow"
	case dspi.CsActTrigger:
		return "trigger"
	case dspi.CsActIndEquals:
		return "ind-equals"
	case dspi.CsActMomentary:
		return "momentary"
	case dspi.CsActIndAbove:
		return "ind-above"
	case dspi.CsActIndLevel:
		return "ind-level"
	default:
		return fmt.Sprintf("unknown(%d)", action)
	}
}

func csEventName(event uint8) string {
	switch event {
	case dspi.CsEventPress:
		return "press"
	case dspi.CsEventLong:
		return "long"
	case dspi.CsEventDouble:
		return "double"
	default:
		return fmt.Sprintf("unknown(%d)", event)
	}
}

func runCsBindingSet(cmd *cobra.Command, args []string) error {
	slot, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid slot: %w", err)
	}

	typeName, _ := cmd.Flags().GetString("type")
	if typeName == "" {
		return fmt.Errorf("--type is required")
	}
	typ, err := dspi.ParseCsType(typeName)
	if err != nil {
		return err
	}

	var noun, action uint8
	if typ == dspi.CsTypeIR {
		// An IR binding is a container: noun/action must stay 0.
		if cmd.Flags().Changed("noun") || cmd.Flags().Changed("action") {
			return fmt.Errorf("an ir binding is a container and takes no noun/action")
		}
	} else {
		nounName, _ := cmd.Flags().GetString("noun")
		if nounName == "" {
			return fmt.Errorf("--noun is required")
		}
		noun, err = dspi.ParseCsNoun(nounName)
		if err != nil {
			return err
		}

		actionName, _ := cmd.Flags().GetString("action")
		if actionName == "" {
			return fmt.Errorf("--action is required")
		}
		action, err = dspi.ParseCsAction(actionName)
		if err != nil {
			return err
		}
	}

	b := &dspi.CsBinding{
		Type:   typ,
		Noun:   noun,
		Action: action,
		GPIO:   [2]uint8{dspi.CsGPIOUnused, dspi.CsGPIOUnused},
	}

	gpio, _ := cmd.Flags().GetString("gpio")
	if gpio != "" {
		pins := strings.Split(gpio, ",")
		for i, p := range pins {
			pin, err := strconv.Atoi(strings.TrimSpace(p))
			if err != nil {
				return fmt.Errorf("invalid GPIO %q: %w", p, err)
			}
			if i > 1 {
				return fmt.Errorf("a binding uses at most 2 GPIOs")
			}
			b.GPIO[i] = uint8(pin)
		}
	}

	if cmd.Flags().Changed("event") {
		eventName, _ := cmd.Flags().GetString("event")
		switch eventName {
		case "press":
			b.Event = dspi.CsEventPress
		case "long":
			b.Event = dspi.CsEventLong
		case "double":
			b.Event = dspi.CsEventDouble
		default:
			return fmt.Errorf("invalid event %q (expected press, long, or double)", eventName)
		}
	}

	b.Target = uint8(mustFlagInt(cmd, "target"))
	b.Index = uint8(mustFlagInt(cmd, "index"))
	b.Value = int16(mustFlagInt(cmd, "value"))
	b.Step = int16(mustFlagInt(cmd, "step"))
	b.RangeMin = int16(mustFlagInt(cmd, "range-min"))
	b.RangeMax = int16(mustFlagInt(cmd, "range-max"))
	if invert, _ := cmd.Flags().GetBool("invert"); invert {
		b.Flags |= dspi.CsFlagInvert
	}

	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		if err := d.SetCsBinding(slot, b); err != nil {
			slog.Error("setting CS binding failed", "serial", d.Serial(), "error", err)
			continue
		}
		fmt.Printf("%s: slot %d: %s -> %s (%s) staged\n",
			d.Serial(), slot, dspi.CsTypeName(b.Type), dspi.CsNounName(b.Noun), csActionName(b.Action))
	}

	return nil
}

func mustFlagInt(cmd *cobra.Command, name string) int {
	v, _ := cmd.Flags().GetInt(name)
	return v
}

func runCsNameGet(cmd *cobra.Command, args []string) error {
	slot, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid slot: %w", err)
	}

	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		name, err := d.GetCsName(slot)
		if err != nil {
			slog.Error("getting CS name failed", "serial", d.Serial(), "error", err)
			continue
		}
		if name == "" {
			name = "(unset)"
		}
		fmt.Printf("%s: slot %d name: %s\n", d.Serial(), slot, name)
	}

	return nil
}

func runCsNameSet(cmd *cobra.Command, args []string) error {
	slot, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid slot: %w", err)
	}

	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		if err := d.SetCsName(slot, args[1]); err != nil {
			slog.Error("setting CS name failed", "serial", d.Serial(), "error", err)
			continue
		}
		fmt.Printf("%s: slot %d name: %s\n", d.Serial(), slot, args[1])
	}

	return nil
}

func runCsIRGet(cmd *cobra.Command, args []string) error {
	subslot, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid sub-slot: %w", err)
	}

	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		c, err := d.GetCsIrCommand(subslot)
		if err != nil {
			slog.Error("getting CS IR command failed", "serial", d.Serial(), "error", err)
			continue
		}

		if c.Protocol == dspi.CsIRProtoNone {
			fmt.Printf("%s: IR sub-slot %d: (empty)\n", d.Serial(), subslot)
			continue
		}

		fmt.Printf("%s: IR sub-slot %d: %s (%s) proto=%s code=0x%08X\n",
			d.Serial(), subslot, dspi.CsNounName(c.Noun), csActionName(c.Action),
			csIRProtoName(c.Protocol), c.Code)
		if c.Target != 0 || c.Index != 0 {
			fmt.Printf("  target=%d index=%d\n", c.Target, c.Index)
		}
		fmt.Printf("  value=%d step=%d flags=0x%02X\n", c.Value, c.Step, c.Flags)
	}

	return nil
}

func csIRProtoName(proto uint8) string {
	switch proto {
	case dspi.CsIRProtoNone:
		return "none"
	case dspi.CsIRProtoNEC:
		return "nec"
	case dspi.CsIRProtoRC5:
		return "rc5"
	case dspi.CsIRProtoRC6:
		return "rc6"
	case dspi.CsIRProtoHash:
		return "hash"
	default:
		return fmt.Sprintf("unknown(%d)", proto)
	}
}

func runCsIRSet(cmd *cobra.Command, args []string) error {
	subslot, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid sub-slot: %w", err)
	}

	nounName, _ := cmd.Flags().GetString("noun")
	if nounName == "" {
		return fmt.Errorf("--noun is required")
	}
	noun, err := dspi.ParseCsNoun(nounName)
	if err != nil {
		return err
	}

	actionName, _ := cmd.Flags().GetString("action")
	if actionName == "" {
		return fmt.Errorf("--action is required")
	}
	action, err := dspi.ParseCsAction(actionName)
	if err != nil {
		return err
	}

	protoName, _ := cmd.Flags().GetString("protocol")
	if protoName == "" {
		return fmt.Errorf("--protocol is required")
	}
	var proto uint8
	switch protoName {
	case "nec":
		proto = dspi.CsIRProtoNEC
	case "rc5":
		proto = dspi.CsIRProtoRC5
	case "rc6":
		proto = dspi.CsIRProtoRC6
	case "hash":
		proto = dspi.CsIRProtoHash
	default:
		return fmt.Errorf("invalid protocol %q (expected nec, rc5, rc6, or hash)", protoName)
	}

	codeName, _ := cmd.Flags().GetString("code")
	if codeName == "" {
		return fmt.Errorf("--code is required (hex, e.g. 0x00FF10EF)")
	}
	code, err := strconv.ParseUint(strings.TrimPrefix(codeName, "0x"), 16, 32)
	if err != nil {
		return fmt.Errorf("invalid code %q: %w", codeName, err)
	}

	c := &dspi.IrCommand{
		Noun:     noun,
		Action:   action,
		Target:   uint8(mustFlagInt(cmd, "target")),
		Index:    uint8(mustFlagInt(cmd, "index")),
		Protocol: proto,
		Value:    int16(mustFlagInt(cmd, "value")),
		Step:     int16(mustFlagInt(cmd, "step")),
		Code:     uint32(code),
	}

	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		if err := d.SetCsIrCommand(subslot, c); err != nil {
			slog.Error("setting CS IR command failed", "serial", d.Serial(), "error", err)
			continue
		}
		fmt.Printf("%s: IR sub-slot %d: %s (%s) proto=%s code=0x%08X staged\n",
			d.Serial(), subslot, dspi.CsNounName(c.Noun), csActionName(c.Action),
			csIRProtoName(c.Protocol), c.Code)
	}

	return nil
}

func runCsIRLearn(cmd *cobra.Command, args []string) error {
	action := "arm"
	if len(args) > 0 {
		action = args[0]
	}

	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		switch action {
		case "arm":
			if err := d.CsIrLearnArm(); err != nil {
				slog.Error("arming CS IR learn failed", "serial", d.Serial(), "error", err)
				continue
			}
			fmt.Printf("%s: IR learn armed; press the remote button, then run `cs ir learn read`\n", d.Serial())
		case "cancel":
			if err := d.CsIrLearnCancel(); err != nil {
				slog.Error("cancelling CS IR learn failed", "serial", d.Serial(), "error", err)
				continue
			}
			fmt.Printf("%s: IR learn cancelled\n", d.Serial())
		case "read":
			res, err := d.CsIrLearnRead()
			if err != nil {
				slog.Error("reading CS IR learn result failed", "serial", d.Serial(), "error", err)
				continue
			}
			fmt.Printf("%s: IR learn state=%s proto=%s code=0x%08X\n",
				d.Serial(), csIRLearnStateName(res.State), csIRProtoName(res.Protocol), res.Code)
		default:
			return fmt.Errorf("invalid learn action %q (expected arm, cancel, or read)", action)
		}
	}

	return nil
}

func runCsSave(cmd *cobra.Command, args []string) error {
	return csSaveRevertForDevices(func(d *dspi.Device) error { return d.CsSave() }, "saved")
}

func runCsRevert(cmd *cobra.Command, args []string) error {
	return csSaveRevertForDevices(func(d *dspi.Device) error { return d.CsRevert() }, "reverted")
}

func csSaveRevertForDevices(do func(*dspi.Device) error, verb string) error {
	devices, err := openDevices()
	if err != nil {
		return fmt.Errorf("opening DSPi devices: %w", err)
	}
	defer closeDevices(devices)

	for _, d := range devices {
		if err := do(d); err != nil {
			slog.Error("CS save/revert failed", "serial", d.Serial(), "error", err)
			continue
		}
		fmt.Printf("%s: CS config %s (poll `cs status` for the outcome)\n", d.Serial(), verb)
	}

	return nil
}
