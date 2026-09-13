package xlog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AeonDigital/Go-Core-xfs/pkg/xfs"
	"github.com/AeonDigital/Go-Core-xutils/module/xtime/pkg/xtime"
	"github.com/AeonDigital/Go-Core-xutils/module/xunits/pkg/xunits"
	"golang.org/x/term"
)

// The variables below are defined as mutable function references to enable
// controlled dependency injection and monkey patching during unit tests.
// Do not modify them in production code.

var (
	fnMockable_IsCLISupportsColors = isCLISupportsColors

	fnMockable_IsDir         = xfs.IsDir
	fnMockable_CreateDirPath = xfs.CreateDirPath
	fnMockable_OpenFileWrite = xfs.OpenFileWrite
	fnMockable_GetUserLogDir = xfs.GetUserLogDir
)

// LogHandler intercepts standard slog records to apply custom multi-destination
// routing, formatting layouts, and ANSI colorization.
type LogHandler struct {
	TimeFormat string `json:"timeFormat"`
	TimeZone   string `json:"timeZone"`

	UseTimeFormat string         `json:"useTimeFormat"`
	UseTimeZone   *time.Location `json:"useTimeZone"`

	LogCLI             bool             `json:"logCLI"`
	LogCLILevel        LogLevel         `json:"logCLILevel"`
	LogCLIColors       bool             `json:"logCLIColors"`
	LogCLIColorPallete *CLIColorPalette `json:"logCLIColorPallete"`

	LogRegistry            bool                `json:"logRegistry"`
	LogRegistryLevel       LogLevel            `json:"logRegistryLevel"`
	LogRegistryDirPath     string              `json:"logRegistryFilePath"`
	LogRegistryFileName    string              `json:"logRegistryFileName"`
	LogRegistryFileMaxSize xunits.Bytes        `json:"logRegistryFileMaxSize"`
	LogRegistryFileMaxAge  xunits.TimeDuration `json:"logRegistryFileMaxAge"`

	logRegistryFile *os.File
}

// isCLISupportsColors natively verifies if the current terminal runtime
// environment supports ANSI color sequences based on standard TTY checks
// and the industry-standard 'NO_COLOR' environment specification.
func isCLISupportsColors() bool {
	// 1. Honor the 'NO_COLOR' specification (https://no-color.org)
	if _, exists := os.LookupEnv("NO_COLOR"); exists {
		return false
	}

	// 2. Verify basic terminal compatibility capabilities
	termEnv := os.Getenv("TERM")
	if termEnv == "dumb" {
		return false
	}

	// 3. Ensure both standard descriptors are interactive TTY terminals
	// preventing ANSI leakages into piped streams (e.g., my-cli > output.txt)
	isStdoutTTY := term.IsTerminal(int(os.Stdout.Fd()))
	isStderrTTY := term.IsTerminal(int(os.Stderr.Fd()))

	return isStdoutTTY && isStderrTTY
}

// fixANSIEscape normalizes visual string representations of ANSI escape codes
// into real byte literals executable by terminals.
func fixANSIEscape(str string) string {
	str = strings.ReplaceAll(str, `\033`, "\x1b")
	str = strings.ReplaceAll(str, `\x1b`, "\x1b")
	return str
}

// checkTimeConfiguration evaluates and loads layout patterns and target time locations.
func (o *LogHandler) checkTimeConfiguration() error {
	var err error

	if o.TimeFormat == "" {
		o.TimeFormat = DefaultTimeFormat
	}
	if o.TimeZone == "" {
		o.TimeZone = DefaultTimeZone
	}

	o.UseTimeFormat = xtime.FormatGenericDateTimeToGolangLayout(o.TimeFormat)
	o.UseTimeZone, err = time.LoadLocation(o.TimeZone)
	if err != nil {
		return err
	}

	return nil
}

// checkCLIConfiguration normalizes and configures the interactive CLI color pallete settings.
func (o *LogHandler) checkCLIConfiguration() {
	if !o.LogCLI || o.LogCLILevel == LevelNone {
		o.LogCLILevel = LevelNone
		o.LogCLIColors = false
		o.LogCLIColorPallete = &CLIColorPalette{}
		return
	}
	if !o.LogCLIColors {
		o.LogCLIColorPallete = &CLIColorPalette{}
	}

	if o.LogCLIColors {
		o.LogCLIColors = fnMockable_IsCLISupportsColors()

		if o.LogCLIColors && o.LogCLIColorPallete == nil {
			o.LogCLIColorPallete = DefaultCLIPalette
		}

		o.LogCLIColorPallete.Debug = fixANSIEscape(o.LogCLIColorPallete.Debug)
		o.LogCLIColorPallete.Info = fixANSIEscape(o.LogCLIColorPallete.Info)
		o.LogCLIColorPallete.Warn = fixANSIEscape(o.LogCLIColorPallete.Warn)
		o.LogCLIColorPallete.Error = fixANSIEscape(o.LogCLIColorPallete.Error)
		o.LogCLIColorPallete.DateTime = fixANSIEscape(o.LogCLIColorPallete.DateTime)
	}
}

// checkRegistryConfiguration verifies, structures, and safely attaches the file descriptor targets for logging.
func (o *LogHandler) checkRegistryConfiguration(logFileName string) error {
	var err error

	// 1. Early exit if file registry tracking is deactivated
	if !o.LogRegistry || o.LogRegistryLevel == LevelNone {
		o.LogRegistryLevel = LevelNone
		o.LogRegistryDirPath = ""
		o.LogRegistryFileName = ""
		o.LogRegistryFileMaxSize = xunits.Bytes(0)
		o.LogRegistryFileMaxAge = xunits.TimeDuration{}
		return nil
	}

	// 2. Guarantee target fallback name conventions
	if logFileName == "" {
		logFileName = "current.log"
	}
	o.LogRegistryFileName = logFileName

	// 3. Define target working base directories if unassigned
	if o.LogRegistryDirPath == "" {
		o.LogRegistryDirPath, err = fnMockable_GetUserLogDir()
		if err != nil {
			return err
		}
	}

	// 4. Ensure directory tree topologies exist safely on disk
	if !fnMockable_IsDir(o.LogRegistryDirPath) {
		err = fnMockable_CreateDirPath(o.LogRegistryDirPath)
		if err != nil {
			return err
		}
	}

	// 5. Build absolute references and open target active descriptors
	fullFilePath := filepath.Join(o.LogRegistryDirPath, o.LogRegistryFileName)

	o.logRegistryFile, err = fnMockable_OpenFileWrite(fullFilePath, false)
	if err != nil {
		return err
	}

	// 6. Enforce safe architectural operational minimum boundaries
	minBytes := (1 * xunits.MB)
	minAge := xunits.TimeDuration{Duration: 1 * time.Minute}

	if o.LogRegistryFileMaxSize < minBytes {
		o.LogRegistryFileMaxSize = minBytes
	}
	if o.LogRegistryFileMaxAge.Duration < minAge.Duration {
		o.LogRegistryFileMaxAge.Duration = minAge.Duration
	}

	return nil
}

// generateLogMessage computes decoupled layout buffers ready for CLI text streams or file writing.
func (o *LogHandler) generateLogMessage(timeStr string, levelStr string, msg string) (string, string) {
	strTime := timeStr
	strLevel := "[" + strings.ToUpper(levelStr) + "]"

	logCLI := fmt.Sprintf("%s %s %s\n", strTime, strLevel, msg)
	logRegistry := logCLI

	if o.LogCLIColors {
		strTime = o.LogCLIColorPallete.DateTime + timeStr + ansiReset

		switch strLevel {
		case "[DEBUG]":
			strLevel = o.LogCLIColorPallete.Debug + strLevel + ansiReset
		case "[INFO]":
			strLevel = o.LogCLIColorPallete.Info + strLevel + ansiReset
		case "[WARN]":
			strLevel = o.LogCLIColorPallete.Warn + strLevel + ansiReset
		case "[ERROR]":
			strLevel = o.LogCLIColorPallete.Error + strLevel + ansiReset
		}
		logCLI = fmt.Sprintf("%s %s %s\n", strTime, strLevel, msg)
	}

	return logCLI, logRegistry
}

// CheckConfiguration processes structural validation across all target runtime parameters.
func (o *LogHandler) CheckConfiguration(logFileName string) error {
	var err error

	err = o.checkTimeConfiguration()
	if err != nil {
		return err
	}

	o.checkCLIConfiguration()

	err = o.checkRegistryConfiguration(logFileName)
	if err != nil {
		return err
	}

	return nil
}

// Enabled decides whether the framework allows processing logs under the specified context conditions.
func (o *LogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return o.LogCLI || o.LogRegistry
}

// Handle processes incoming record fields routing them into standard channels or output logs.
func (o *LogHandler) Handle(_ context.Context, r slog.Record) error {
	timeStr := r.Time.In(o.UseTimeZone).Format(o.UseTimeFormat)
	logCLI, logRegistry := o.generateLogMessage(timeStr, r.Level.String(), r.Message)

	if o.LogCLI {
		var target io.Writer = os.Stdout
		if r.Level == slog.LevelError {
			target = os.Stderr
		}

		fmt.Fprintf(target, "%s", logCLI)
	}

	if o.LogRegistry {
		_, err := o.logRegistryFile.WriteString(logRegistry)
		if err != nil {
			return err
		}
	}

	return nil
}

// WithAttrs returns a slice of the receiver handler keeping structural slog interfaces consistent.
func (o *LogHandler) WithAttrs(attrs []slog.Attr) slog.Handler { return o }

// WithGroup structures operational logging domains ensuring clean tracking scoping.
func (o *LogHandler) WithGroup(name string) slog.Handler { return o }
