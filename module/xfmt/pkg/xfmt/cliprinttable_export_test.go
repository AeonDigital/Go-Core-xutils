package xfmt

// Export private functions for white-box testing in xfmt_test package.
var (
	ExportCalcColumnWidths   = cliPrintTableCalculateColumnWidths
	ExportWrapString         = cliPrintTableWrapString
	ExportBuildTableRows     = cliPrintTableBuildTableRows
	ExportCliIsTerminal      = &cliIsTerminal
	ExportCliGetTerminalSize = &cliGetTerminalSize

	ExportJsonPrettyEncoderExecute = &jsonPrettyEncoderExecute
)
