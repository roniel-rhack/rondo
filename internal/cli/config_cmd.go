package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/roniel/todo-app/internal/config"
	"github.com/spf13/cobra"
)

// configKey describes a single configuration key with get/set accessors.
type configKey struct {
	description string
	get         func(c config.Config) string
	set         func(c *config.Config, val string) error
}

var configKeys = map[string]configKey{
	"panel_ratio": {
		description: "Panel width ratio (0.2–0.8)",
		get: func(c config.Config) string {
			return strconv.FormatFloat(c.PanelRatio, 'f', 2, 64)
		},
		set: func(c *config.Config, val string) error {
			v, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return fmt.Errorf("panel_ratio must be a number between 0.2 and 0.8")
			}
			if v < 0.2 || v > 0.8 {
				return fmt.Errorf("panel_ratio must be between 0.2 and 0.8, got %g", v)
			}
			c.PanelRatio = v
			return nil
		},
	},
	"date_format": {
		description: "Date format (Go layout or preset: iso, european, us)",
		get:         func(c config.Config) string { return c.DateFormat },
		set: func(c *config.Config, val string) error {
			val = config.ResolvePreset(val, config.DateFormatPresets)
			if err := config.ValidateTimeLayout(val); err != nil {
				return fmt.Errorf("date_format: %w", err)
			}
			c.DateFormat = val
			return nil
		},
	},
	"time_format": {
		description: "Time format (Go layout or preset: 24h, 12h)",
		get:         func(c config.Config) string { return c.TimeFormat },
		set: func(c *config.Config, val string) error {
			val = config.ResolvePreset(val, config.TimeFormatPresets)
			if err := config.ValidateTimeLayout(val); err != nil {
				return fmt.Errorf("time_format: %w", err)
			}
			c.TimeFormat = val
			return nil
		},
	},
	"datetime_format": {
		description: "Date+time format (Go layout or preset: iso, european, us)",
		get:         func(c config.Config) string { return c.DateTimeFormat },
		set: func(c *config.Config, val string) error {
			val = config.ResolvePreset(val, config.DateTimeFormatPresets)
			if err := config.ValidateTimeLayout(val); err != nil {
				return fmt.Errorf("datetime_format: %w", err)
			}
			c.DateTimeFormat = val
			return nil
		},
	},
	"focus.work_duration_min": {
		description: "Work session duration in minutes (1–120)",
		get:         func(c config.Config) string { return strconv.Itoa(c.Focus.WorkDuration) },
		set: func(c *config.Config, val string) error {
			v, err := parseMinutes(val, 1, 120)
			if err != nil {
				return fmt.Errorf("focus.work_duration_min: %w", err)
			}
			c.Focus.WorkDuration = v
			return nil
		},
	},
	"focus.short_break_duration_min": {
		description: "Short break duration in minutes (1–60)",
		get:         func(c config.Config) string { return strconv.Itoa(c.Focus.ShortBreakDuration) },
		set: func(c *config.Config, val string) error {
			v, err := parseMinutes(val, 1, 60)
			if err != nil {
				return fmt.Errorf("focus.short_break_duration_min: %w", err)
			}
			c.Focus.ShortBreakDuration = v
			return nil
		},
	},
	"focus.long_break_duration_min": {
		description: "Long break duration in minutes (1–120)",
		get:         func(c config.Config) string { return strconv.Itoa(c.Focus.LongBreakDuration) },
		set: func(c *config.Config, val string) error {
			v, err := parseMinutes(val, 1, 120)
			if err != nil {
				return fmt.Errorf("focus.long_break_duration_min: %w", err)
			}
			c.Focus.LongBreakDuration = v
			return nil
		},
	},
	"focus.long_break_interval": {
		description: "Work sessions before a long break (1–10)",
		get:         func(c config.Config) string { return strconv.Itoa(c.Focus.LongBreakInterval) },
		set: func(c *config.Config, val string) error {
			v, err := parseMinutes(val, 1, 10)
			if err != nil {
				return fmt.Errorf("focus.long_break_interval: %w", err)
			}
			c.Focus.LongBreakInterval = v
			return nil
		},
	},
	"focus.daily_goal": {
		description: "Daily focus session goal",
		get:         func(c config.Config) string { return strconv.Itoa(c.Focus.DailyGoal) },
		set: func(c *config.Config, val string) error {
			v, err := parseMinutes(val, 1, 100)
			if err != nil {
				return fmt.Errorf("focus.daily_goal: %w", err)
			}
			c.Focus.DailyGoal = v
			return nil
		},
	},
	"focus.auto_start_break": {
		description: "Auto-start breaks after work sessions (true/false)",
		get: func(c config.Config) string {
			if c.Focus.AutoStartBreak {
				return "true"
			}
			return "false"
		},
		set: func(c *config.Config, val string) error {
			b, err := parseBool(val)
			if err != nil {
				return fmt.Errorf("focus.auto_start_break: %w", err)
			}
			c.Focus.AutoStartBreak = b
			return nil
		},
	},
	"focus.sound": {
		description: "Play sound on session completion (true/false)",
		get: func(c config.Config) string {
			if c.Focus.Sound {
				return "true"
			}
			return "false"
		},
		set: func(c *config.Config, val string) error {
			b, err := parseBool(val)
			if err != nil {
				return fmt.Errorf("focus.sound: %w", err)
			}
			c.Focus.Sound = b
			return nil
		},
	},
}

// orderedConfigKeys defines the display order for config list.
var orderedConfigKeys = []string{
	"panel_ratio",
	"date_format",
	"time_format",
	"datetime_format",
	"focus.work_duration_min",
	"focus.short_break_duration_min",
	"focus.long_break_duration_min",
	"focus.long_break_interval",
	"focus.daily_goal",
	"focus.auto_start_break",
	"focus.sound",
}

func parseMinutes(val string, min, max int) (int, error) {
	v, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("must be an integer, got %q", val)
	}
	if v < min || v > max {
		return 0, fmt.Errorf("must be between %d and %d, got %d", min, max, v)
	}
	return v, nil
}

func parseBool(val string) (bool, error) {
	switch strings.ToLower(val) {
	case "true", "1", "yes", "on":
		return true, nil
	case "false", "0", "no", "off":
		return false, nil
	default:
		return false, fmt.Errorf("must be true or false, got %q", val)
	}
}

func (c *CLI) configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "View and modify configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(c.configListCmd())
	cmd.AddCommand(c.configGetCmd())
	cmd.AddCommand(c.configSetCmd())
	cmd.AddCommand(c.configResetCmd())

	return cmd
}

func (c *CLI) configListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all configuration keys and values",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			p := c.printer(os.Stdout)
			switch strings.ToLower(c.format) {
			case "json":
				return p.JSON(cfg)
			default:
				rows := make([][]string, 0, len(orderedConfigKeys))
				for _, key := range orderedConfigKeys {
					kd := configKeys[key]
					rows = append(rows, []string{key, kd.get(cfg), kd.description})
				}
				p.Table([]string{"KEY", "VALUE", "DESCRIPTION"}, rows)
				return nil
			}
		},
	}
}

func (c *CLI) configGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			kd, ok := configKeys[key]
			if !ok {
				return fmt.Errorf("unknown config key %q; run 'rondo config list' to see valid keys", key)
			}

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			fmt.Fprintln(os.Stdout, kd.get(cfg))
			return nil
		},
	}
}

func (c *CLI) configSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, val := args[0], args[1]
			kd, ok := configKeys[key]
			if !ok {
				return fmt.Errorf("unknown config key %q; run 'rondo config list' to see valid keys", key)
			}

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			if err := kd.set(&cfg, val); err != nil {
				return err
			}

			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("save config: %w", err)
			}

			p := c.printer(os.Stdout)
			stored := kd.get(cfg)
			if stored != val {
				p.Success("Set %s = %s (resolved: %s)", key, val, stored)
			} else {
				p.Success("Set %s = %s", key, val)
			}
			return nil
		},
	}
}

func (c *CLI) configResetCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset configuration to defaults",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ok, err := Confirm("Reset all configuration to defaults?", force)
			if err != nil {
				return err
			}
			if !ok {
				fmt.Fprintln(os.Stderr, "Cancelled.")
				return nil
			}

			if err := config.Save(config.DefaultConfig()); err != nil {
				return fmt.Errorf("save config: %w", err)
			}

			p := c.printer(os.Stdout)
			p.Success("Configuration reset to defaults")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "y", false, "Skip confirmation prompt")

	return cmd
}
